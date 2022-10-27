package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func (s Server) indexGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := fmt.Fprint(w, "Hello from ScreenJournal"); err != nil {
			log.Fatalf("failed to write HTTP response: %v", err)
		}
	}
}
