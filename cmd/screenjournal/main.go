package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	gorilla "github.com/mtlynch/gorilla-handlers"

	"github.com/mtlynch/screenjournal/v2"
	simple_auth "github.com/mtlynch/screenjournal/v2/auth/simple"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
	simple_sessions "github.com/mtlynch/screenjournal/v2/sessions/simple"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

func main() {
	log.Print("Starting screenjournal server")

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	dbPath := flag.String("db", "data/store.db", "path to database")
	flag.Parse()

	adminUsername := screenjournal.Username(requireEnv("SJ_ADMIN_USERNAME"))
	adminPassword := screenjournal.Password(requireEnv("SJ_ADMIN_PASSWORD"))

	authenticator, err := simple_auth.New(adminUsername, adminPassword)
	if err != nil {
		log.Fatalf("invalid shared secret: %v", err)
	}

	sessionManager, err := simple_sessions.New(adminUsername, adminPassword)
	if err != nil {
		log.Fatalf("failed to create session manager: %v", err)
	}

	ensureDirExists(filepath.Dir(*dbPath))
	store := sqlite.New(*dbPath, isLitestreamEnabled())

	metadataFinder, err := tmdb.New(requireEnv("SJ_TMDB_API"))
	if err != nil {
		log.Fatalf("failed to create metadata finder: %v", err)
	}

	h := gorilla.LoggingHandler(os.Stdout, handlers.New(authenticator, sessionManager, store, metadataFinder).Router())
	if os.Getenv("SJ_BEHIND_PROXY") != "" {
		h = gorilla.ProxyIPHeadersHandler(h)
	}
	http.Handle("/", h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4001"
	}
	log.Printf("Listening on %s", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
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
