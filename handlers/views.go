package handlers

import (
	"net/http"
	"path"
	"text/template"
)

func (s Server) indexGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderTemplate(w, "index.html", struct {
			Title string
		}{
			Title: "ScreenJournal",
		}, template.FuncMap{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func renderTemplate(w http.ResponseWriter, templateFilename string, templateVars interface{}, funcMap template.FuncMap) error {
	const templatesRootDir = "./templates"
	const baseTemplate = "base"
	const baseTemplateFilename = "base.html"

	templateFiles := []string{}
	templateFiles = append(templateFiles, path.Join(templatesRootDir, templateFilename))
	templateFiles = append(templateFiles, path.Join(templatesRootDir, baseTemplateFilename))

	t := template.Must(template.New(templateFilename).Funcs(funcMap).
		ParseFiles(templateFiles...))
	if err := t.ExecuteTemplate(w, baseTemplate, templateVars); err != nil {
		return err
	}
	return nil
}
