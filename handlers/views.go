package handlers

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/store"
)

type commonProps struct {
	Title            string
	IsAuthenticated  bool
	IsAdmin          bool
	LoggedInUsername screenjournal.Username
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
		if !inviteCode.Empty() {
			invite, err = s.store.ReadSignupInvitation(inviteCode)
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

		var suggestedUsername string
		if !invite.Empty() {
			nonSuggestedCharsPattern := regexp.MustCompile(`(?i)[^a-z0-9]`)
			firstPart := strings.SplitN(invite.Invitee.String(), " ", 2)[0]
			suggestedUsername = nonSuggestedCharsPattern.ReplaceAllString(strings.ToLower(firstPart), "")
		}

		if err := renderTemplate(w, templateFilename, struct {
			commonProps
			Invitee           screenjournal.Invitee
			SuggestedUsername string
		}{
			commonProps:       makeCommonProps("Sign Up", r.Context()),
			Invitee:           invite.Invitee,
			SuggestedUsername: suggestedUsername,
		}, template.FuncMap{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) reviewsGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionOwner, err := usernameFromRequestPath(r)
		if err != nil {
			collectionOwner = screenjournal.Username("")
		}

		reviews, err := s.store.ReadReviews(store.ReviewFilters{Username: collectionOwner})
		if err != nil {
			log.Printf("failed to read reviews: %v", err)
			http.Error(w, "Failed to read reviews", http.StatusInternalServerError)
			return
		}

		// Sort reviews starting with most recent watch dates.
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].Watched.Time().After(reviews[j].Watched.Time())
		})

		title := "Ratings"
		if !collectionOwner.IsEmpty() {
			title = fmt.Sprintf("%s's %s", collectionOwner, title)
		}

		if err := renderTemplate(w, "reviews-index.html", struct {
			commonProps
			Reviews         []screenjournal.Review
			CollectionOwner screenjournal.Username
		}{
			commonProps:     makeCommonProps(title, r.Context()),
			Reviews:         reviews,
			CollectionOwner: collectionOwner,
		}, template.FuncMap{
			"relativeWatchDate": relativeWatchDate,
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
			"elideBlurb": func(b screenjournal.Blurb) string {
				score := 0
				var elidedChars []rune
				for _, c := range b.String() {
					if c == '\n' {
						score += 50
					} else {
						score += 1
					}
					if score > 350 {
						// Add ellipsis.
						elidedChars = append(elidedChars, '.', '.', '.')
						break
					}
					elidedChars = append(elidedChars, c)
				}
				return string(elidedChars)
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

func (s Server) reviewsReadGet() http.HandlerFunc {
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

		if err := renderTemplate(w, "reviews-view.html", struct {
			commonProps
			Review screenjournal.Review
		}{
			commonProps: makeCommonProps(review.Movie.Title.String(), r.Context()),
			Review:      review,
		}, template.FuncMap{
			"relativeWatchDate": relativeWatchDate,
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
			"splitByNewline": func(s string) []string {
				return strings.Split(s, "\n")
			},
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) reviewsEditGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loggedInUser, ok := userFromContext(r.Context())
		if !ok {
			http.Error(w, "You must be logged in to edit reviews", http.StatusUnauthorized)
			return
		}

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

		if !review.Owner.Equal(loggedInUser.Username) {
			http.Error(w, "You can't edit another user's review", http.StatusForbidden)
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

func relativeWatchDate(t screenjournal.WatchDate) string {
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
}

func makeCommonProps(title string, ctx context.Context) commonProps {
	return commonProps{
		Title:            title,
		IsAuthenticated:  isAuthenticated(ctx),
		IsAdmin:          isAdmin(ctx),
		LoggedInUsername: usernameFromContext(ctx),
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
