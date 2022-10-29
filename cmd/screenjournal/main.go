package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	gorilla "github.com/mtlynch/gorilla-handlers"

	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/handlers/auth/simple"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
)

func main() {
	log.Print("Starting screenjournal server")

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	dbPath := flag.String("db", "data/store.db", "path to database")
	flag.Parse()

	authenticator, err := simple.New(requireEnv("SJ_ADMIN_USERNAME"), requireEnv("SJ_ADMIN_PASSWORD"))
	if err != nil {
		log.Fatalf("invalid shared secret: %v", err)
	}

	ensureDirExists(filepath.Dir(*dbPath))
	store := sqlite.New(*dbPath, isLitestreamEnabled())

	h := gorilla.LoggingHandler(os.Stdout, handlers.New(authenticator, store).Router())
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
