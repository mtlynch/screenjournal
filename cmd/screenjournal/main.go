package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	gorilla "github.com/mtlynch/gorilla-handlers"

	"github.com/mtlynch/screenjournal/v2/handlers"
)

func main() {
	log.Print("Starting screenjournal server")

	h := gorilla.LoggingHandler(os.Stdout, handlers.New().Router())
	if os.Getenv("PS_BEHIND_PROXY") != "" {
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
