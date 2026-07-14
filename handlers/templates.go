package handlers

import (
	"bytes"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

type htmlRenderer struct {
	templateFS          fs.FS
	sharedTemplateFiles []string
}

func newHTMLRenderer(templateFS fs.FS, sharedTemplateFiles ...string) *htmlRenderer {
	return &htmlRenderer{
		templateFS:          templateFS,
		sharedTemplateFiles: sharedTemplateFiles,
	}
}

func (h *htmlRenderer) mustParse(templateFiles ...string) *template.Template {
	return h.mustParseWithFuncs(nil, templateFiles...)
}

func (h *htmlRenderer) mustParseWithFuncs(funcs template.FuncMap, templateFiles ...string) *template.Template {
	files := make([]string, 0, len(h.sharedTemplateFiles)+len(templateFiles))
	files = append(files, h.sharedTemplateFiles...)
	files = append(files, templateFiles...)

	t := template.New("base.html")
	if funcs != nil {
		t = t.Funcs(funcs)
	}
	return template.Must(t.ParseFS(h.templateFS, files...))
}

// render executes the named template with data and writes the result to w. It
// renders into a buffer first so a mid-render failure becomes a clean 500
// instead of a partially written response.
func (h *htmlRenderer) render(w http.ResponseWriter, t *template.Template, name string, data any) bool {
	var body bytes.Buffer
	if err := t.ExecuteTemplate(&body, name, data); err != nil {
		log.Printf("failed to render %q template: %v", name, err)
		http.Error(w, "Failed to render HTML template", http.StatusInternalServerError)
		return false
	}

	w.Header().Add("Vary", "HX-Request")
	_, _ = body.WriteTo(w)
	return true
}
