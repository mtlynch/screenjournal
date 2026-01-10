package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	gorilla "github.com/mtlynch/gorilla-handlers"

	email_announce "github.com/mtlynch/screenjournal/v2/announce/email"
	"github.com/mtlynch/screenjournal/v2/announce/quiet"
	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/email/smtp"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

func main() {
	log.Print("starting screenjournal server")

	log.SetFlags(log.LstdFlags | log.Llongfile)
	dbPath := flag.String("db", "data/store.db", "path to database")
	flag.Parse()

	ensureDirExists(filepath.Dir(*dbPath))
	db := sqlite.MustOpen(*dbPath)
	store := sqlite.New(db, isLitestreamEnabled())

	authenticator := auth.New(store)

	useTls := isTlsRequired()
	if !useTls {
		log.Printf("TLS has not been marked as required, so session cookies will not have Secure flag")
	}
	sessionManager, err := sessions.NewManager(db, useTls)
	if err != nil {
		log.Fatalf("failed to create session manager: %v", err)
	}

	var announcer handlers.Announcer
	if isSmtpEnabled() {
		smtpHost := requireEnv("SJ_SMTP_HOST")
		smtpPort, err := strconv.Atoi(requireEnv("SJ_SMTP_PORT"))
		if err != nil {
			log.Printf("failed to parse SMTP port: %v", err)
		}
		log.Printf("SMTP is enabled using server at %s:%d", smtpHost, smtpPort)
		mailSender, err := smtp.New(smtpHost, smtpPort, requireEnv("SJ_SMTP_USERNAME"), requireEnv("SJ_SMTP_PASSWORD"))
		if err != nil {
			log.Fatalf("failed to create mail sender: %v", err)
		}
		announcer = email_announce.New(requireEnv("SJ_BASE_URL"), mailSender, store)
	} else {
		log.Printf("SMTP not configured. Transactional emails are disabled")
		announcer = quiet.New()
	}

	metadataFinder, err := tmdb.New(requireEnv("SJ_TMDB_API"))
	if err != nil {
		log.Fatalf("failed to create metadata finder: %v", err)
	}

	h := gorilla.LoggingHandler(os.Stdout, handlers.New(authenticator, announcer, sessionManager, store, metadataFinder).Router())
	if os.Getenv("SJ_BEHIND_PROXY") != "" {
		h = gorilla.ProxyIPHeadersHandler(h)
	}
	http.Handle("/", h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4003"
	}
	log.Printf("listening on %s", port)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("missing required environment variable: %s", key)
	}
	return val
}

func ensureDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func isLitestreamEnabled() bool {
	return os.Getenv("LITESTREAM_BUCKET") != ""
}

func isSmtpEnabled() bool {
	return os.Getenv("SJ_SMTP_USERNAME") != ""
}

func isTlsRequired() bool {
	if os.Getenv("SJ_REQUIRE_TLS") == "false" {
		return false
	}
	return defaultIsTlsRequired
}
