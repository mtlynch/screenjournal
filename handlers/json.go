package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Fatalf("failed to encode JSON response: %v", err)
	}
}

func respondHTML(w http.ResponseWriter, data string) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(data)); err != nil {
		log.Fatalf("failed to encode JSON response: %v", err)
	}
}
