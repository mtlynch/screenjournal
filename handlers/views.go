package handlers

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/store"
)

type commonProps struct {
	Title           string
	IsAuthenticated bool
	IsAdmin         bool
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

func (s Server) signUpGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templateFilename := "sign-up.html"

		inviteCode, err := inviteCodeFromQueryParams(r)
		if err != nil {
			log.Printf("invalid invite code: %v", err)
			http.Error(w, "Invalid invite code", http.StatusBadRequest)
			return
		}

		var invite screenjournal.SignupInvitation

		if inviteCode != "" {
			invites, err := s.store.ReadSignupInvitations()
			if err != nil {
				log.Printf("failed to read signup invitations: %v", err)
				http.Error(w, "Failed to read signup invitations", http.StatusInternalServerError)
				return
			}

			findInvite := func(code screenjournal.InviteCode) (screenjournal.SignupInvitation, error) {
				for _, invite := range invites {
					if invite.InviteCode == code {
						return invite, nil
					}
				}
				return screenjournal.SignupInvitation{}, errors.New("invite code does not exist")
			}

			invite, err = findInvite(inviteCode)
			if err != nil {
				log.Printf("invalid invite code: %v", err)
				http.Error(w, "Invalid invite code", http.StatusUnauthorized)
				return
			}
		}

		uc, err := s.store.CountUsers()
		if err != nil {
			log.Printf("failed to count users: %v", err)
			http.Error(w, "Failed to load signup template", http.StatusInternalServerError)
			return
		}

		if uc > 0 && invite.Empty() {
			templateFilename = "sign-up-by-invitation.html"
		}
		if err := renderTemplate(w, templateFilename, struct {
			commonProps
			Invitee screenjournal.Invitee
		}{
			commonProps: makeCommonProps("Sign Up", r.Context()),
			Invitee:     invite.Invitee,
		}, template.FuncMap{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) reviewsGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, err := usernameFromRequestPath(r)
		if err != nil {
			owner = screenjournal.Username("")
		}

		reviews, err := s.store.ReadReviews(store.ReviewFilters{Username: owner})
		if err != nil {
			log.Printf("failed to read reviews: %v", err)
			http.Error(w, "Failed to read reviews", http.StatusInternalServerError)
			return
		}

		// Sort reviews starting with most recent watch dates.
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].Watched.Time().After(reviews[j].Watched.Time())
		})

		var loggedInUserIsOwner bool
		if !owner.IsEmpty() {
			loggedInUserIsOwner = usernameFromContext(r.Context()).Equal(owner)
		}

		title := "Ratings"
		if !owner.IsEmpty() {
			title = fmt.Sprintf("%s's %s", owner, title)
		}

		if err := renderTemplate(w, "reviews-index.html", struct {
			commonProps
			Reviews             []screenjournal.Review
			Owner               screenjournal.Username
			LoggedInUserIsOwner bool
		}{
			commonProps:         makeCommonProps(title, r.Context()),
			Reviews:             reviews,
			Owner:               owner,
			LoggedInUserIsOwner: loggedInUserIsOwner,
		}, template.FuncMap{
			"relativeWatchDate": func(t screenjournal.WatchDate) string {
				daysAgo := int(time.Since(t.Time()).Hours() / 24)
				weeksAgo := int(daysAgo / 7)
				if daysAgo < 1 {
					return "today"
				} else if daysAgo == 1 {
					return "yesterday"
				} else if daysAgo <= 14 {
					return fmt.Sprintf("%d days ago", daysAgo)
				} else if weeksAgo < 8 {
					return fmt.Sprintf("%d weeks ago", weeksAgo)
				}
				monthsAgo := int(daysAgo / 30)
				return fmt.Sprintf("%d months ago", monthsAgo)
			},
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
			"splitByNewline": func(s string) []string {
				return strings.Split(s, "\n")
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

func (s Server) reviewsEditGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := reviewIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, err := s.store.ReadReview(id)
		if err == store.ErrReviewNotFound {
			http.Error(w, "Invalid review ID", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to read review: %v", err)
			http.Error(w, "Failed to read review", http.StatusInternalServerError)
			return
		}

		if err := renderTemplate(w, "reviews-edit.html", struct {
			commonProps
			RatingOptions []int
			Review        screenjournal.Review
			Today         time.Time
		}{
			commonProps:   makeCommonProps("Edit Review", r.Context()),
			RatingOptions: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			Review:        review,
			Today:         time.Now(),
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
			"formatDate": func(t time.Time) string {
				return t.Format("2006-01-02")
			},
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) reviewsNewGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderTemplate(w, "reviews-new.html", struct {
			commonProps
			RatingOptions []int
			Today         time.Time
		}{
			commonProps:   makeCommonProps("Add Review", r.Context()),
			RatingOptions: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			Today:         time.Now(),
		}, template.FuncMap{
			"formatDate": func(t time.Time) string {
				return t.Format("2006-01-02")
			},
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) invitesGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		invites, err := s.store.ReadSignupInvitations()
		if err != nil {
			log.Printf("failed to read signup invitations: %v", err)
			http.Error(w, "Failed to read signup invitations", http.StatusInternalServerError)
			return
		}
		if err := renderTemplate(w, "invites.html", struct {
			commonProps
			Invites []screenjournal.SignupInvitation
		}{
			commonProps: makeCommonProps("Invites", r.Context()),
			Invites:     invites,
		}, template.FuncMap{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) invitesNewGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderTemplate(w, "invites-new.html", struct {
			commonProps
		}{
			commonProps: makeCommonProps("Create Invite Link", r.Context()),
		}, template.FuncMap{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func makeCommonProps(title string, ctx context.Context) commonProps {
	return commonProps{
		Title:           title,
		IsAuthenticated: isAuthenticated(ctx),
		IsAdmin:         isAdmin(ctx),
	}
}

//go:embed templates
var templatesFS embed.FS

func renderTemplate(w http.ResponseWriter, templateFilename string, templateVars interface{}, funcMap template.FuncMap) error {
	t := template.New(templateFilename).Funcs(funcMap)
	t = template.Must(
		t.ParseFS(
			templatesFS,
			"templates/custom-elements/*.html",
			"templates/layouts/*.html",
			"templates/partials/*.html",
			path.Join("templates/pages", templateFilename)))

	if err := t.ExecuteTemplate(w, "base", templateVars); err != nil {
		return err
	}
	return nil
}
