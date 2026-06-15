package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

// renderTemplate executes the named template with data and writes the result to
// w. It renders into a buffer first so a mid-render failure becomes a clean 500
// instead of a partially written response.
func renderTemplate(w http.ResponseWriter, t *template.Template, name string, data any) bool {
	var body bytes.Buffer
	if err := t.ExecuteTemplate(&body, name, data); err != nil {
		log.Printf("failed to render %q template: %v", name, err)
		http.Error(w, "Failed to render HTML template", http.StatusInternalServerError)
		return false
	}

	_, _ = body.WriteTo(w)
	return true
}
