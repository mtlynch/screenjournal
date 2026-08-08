package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	gorilla "github.com/mtlynch/gorilla-handlers"

	email_announce "github.com/mtlynch/screenjournal/v2/announce/email"
	"github.com/mtlynch/screenjournal/v2/announce/quiet"
	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/email/smtp"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
	"github.com/mtlynch/screenjournal/v2/passwordreset"
	passwordreset_email "github.com/mtlynch/screenjournal/v2/passwordreset/email"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

// shutdownTimeout bounds in-flight work so Litestream has time for its final
// sync before Fly terminates the machine.
const shutdownTimeout = 15 * time.Second

func main() {
	log.Print("starting screenjournal server")

	log.SetFlags(log.LstdFlags | log.Llongfile)
	dbPath := flag.String("db", "data/store.db", "path to database")
	flag.Parse()

	ensureDirExists(filepath.Dir(*dbPath))
	db := sqlite.MustOpen(*dbPath, isLitestreamEnabled())
	store := sqlite.New(db)

	authenticator := auth.New(store)

	useTls := isTlsRequired()
	if !useTls {
		log.Printf("TLS has not been marked as required, so session cookies will not have Secure flag")
	}
	sessionManager := sessions.NewManager(store, useTls)

	var announcer handlers.Announcer
	var passwordResetter handlers.PasswordResetter
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
		baseURL := requireEnv("SJ_BASE_URL")
		announcer = email_announce.New(baseURL, mailSender, store)
		passwordResetter = passwordreset.New(store, passwordreset_email.New(baseURL, mailSender), time.Now)
	} else {
		log.Printf("SMTP not configured. Transactional emails are disabled")
		announcer = quiet.New()
	}

	tmdbBaseURL := os.Getenv("SJ_TMDB_API_BASE_URL")
	if tmdbBaseURL == "" {
		tmdbBaseURL = tmdb.DefaultBaseURL
	}
	metadataFinder, err := tmdb.New(tmdbBaseURL, requireEnv("SJ_TMDB_API"))
	if err != nil {
		log.Fatalf("failed to create metadata finder: %v", err)
	}

	h := gorilla.LoggingHandler(os.Stdout, handlers.New(handlers.ServerParams{
		Authenticator:    authenticator,
		Announcer:        announcer,
		SessionManager:   sessionManager,
		Store:            store,
		MetadataFinder:   metadataFinder,
		PasswordResetter: passwordResetter,
	}).Router())
	if os.Getenv("SJ_BEHIND_PROXY") != "" {
		h = gorilla.ProxyIPHeadersHandler(h)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "4003"
	}
	log.Printf("listening on %s", port)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	shutdownCtx, stopShutdownSignals := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopShutdownSignals()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("failed to serve: %v", err)
			panic(err)
		}
	case <-shutdownCtx.Done():
		shutdown(server, store)
	}
}

// shutdown drains requests and closes SQLite before Litestream's final sync.
// It logs shutdown errors and returns successfully because a nonzero exit code
// prevents Litestream from running its final database sync.
func shutdown(server *http.Server, store sqlite.Store) {
	log.Print("shutdown signal received, draining in-flight requests")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("failed to drain in-flight requests: %v", err)
	}
	if err := store.Close(); err != nil {
		log.Printf("failed to close database: %v", err)
	}

	log.Print("shutdown complete")
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
