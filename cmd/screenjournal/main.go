package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	gorilla "github.com/mtlynch/gorilla-handlers"

	simple_auth "github.com/mtlynch/screenjournal/v2/auth/simple"
	"github.com/mtlynch/screenjournal/v2/handlers"
	jeff_sessions "github.com/mtlynch/screenjournal/v2/handlers/sessions/jeff"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

func main() {
	log.Print("Starting screenjournal server")

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	dbPath := flag.String("db", "data/store.db", "path to database")
	flag.Parse()

	ensureDirExists(filepath.Dir(*dbPath))
	store := sqlite.New(*dbPath, isLitestreamEnabled())

	authenticator, err := simple_auth.New(store)
	if err != nil {
		log.Fatalf("failed to create authenticator: %v", err)
	}

	sessionManager, err := jeff_sessions.New(*dbPath)
	if err != nil {
		log.Fatalf("failed to create session manager: %v", err)
	}

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
