package handlers

import (
	"context"
	"embed"
	"log"
	"net/http"
	"path"
	"sort"
	"text/template"

	"github.com/mtlynch/screenjournal/v2"
)

type commonProps struct {
	Title           string
	IsAuthenticated bool
}

func (s Server) indexGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderTemplate(w, "index.html", struct {
			commonProps
		}{
			commonProps: makeCommonProps("ScreenJournal", r.Context()),
		}, template.FuncMap{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) aboutGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderTemplate(w, "about.html", struct {
			commonProps
		}{
			commonProps: makeCommonProps("About ScreenJournal", r.Context()),
		}, template.FuncMap{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) logInGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderTemplate(w, "login.html", struct {
			commonProps
		}{
			commonProps: makeCommonProps("Sign In", r.Context()),
		}, template.FuncMap{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) logOutGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderTemplate(w, "logout.html", struct {
			commonProps
		}{
			commonProps: makeCommonProps("Log Out", r.Context()),
		}, template.FuncMap{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) signUpGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderTemplate(w, "sign-up.html", struct {
			commonProps
		}{
			commonProps: makeCommonProps("Sign Up", r.Context()),
		}, template.FuncMap{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) reviewsGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reviews, err := s.store.ReadReviews()
		if err != nil {
			log.Printf("failed to read reviews: %v", err)
			http.Error(w, "Failed to read reviews", http.StatusInternalServerError)
			return
		}

		// Sort reviews starting with most recent watch dates.
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].Watched.Time().After(reviews[j].Watched.Time())
		})

		if err := renderTemplate(w, "reviews-index.html", struct {
			commonProps
			Reviews []screenjournal.Review
		}{
			commonProps: makeCommonProps("Ratings", r.Context()),
			Reviews:     reviews,
		}, template.FuncMap{
			"formatWatchDate": func(t screenjournal.WatchDate) string {
				return t.Time().Format("2006-01-02")
			},
			"iterate": func(n uint8) []uint8 {
				var arr []uint8
				var i uint8
				for i = 0; i < n; i++ {
					arr = append(arr, i)
				}
				return arr
			},
			"minus": func(a, b uint8) uint8 {
				return a - b
			},
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func makeCommonProps(title string, ctx context.Context) commonProps {
	return commonProps{
		Title:           title,
		IsAuthenticated: isAuthenticated(ctx),
	}
}

//go:embed templates
var templatesFS embed.FS

func renderTemplate(w http.ResponseWriter, templateFilename string, templateVars interface{}, funcMap template.FuncMap) error {
	t := template.New(templateFilename).Funcs(funcMap)
	t = template.Must(t.ParseFS(templatesFS, "templates/layouts/*.html", "templates/partials/*.html", path.Join("templates/pages", templateFilename)))

	if err := t.ExecuteTemplate(w, "base", templateVars); err != nil {
		return err
	}
	return nil
}
