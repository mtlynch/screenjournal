package handlers

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/markdown"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

type commonProps struct {
	IsAuthenticated  bool
	IsAdmin          bool
	LoggedInUsername screenjournal.Username
	CspNonce         string
}

//go:embed templates
var templatesFS embed.FS

var baseTemplates = []string{
	"templates/layouts/base.html",
	"templates/partials/footer.html",
	"templates/partials/navbar.html",
}

type ratingOption struct {
	Value uint8
	Label string
}

var ratingOptions = []ratingOption{
	{
		Value: 1,
		Label: "0.5",
	},
	{
		Value: 2,
		Label: "1.0",
	},
	{
		Value: 3,
		Label: "1.5",
	},
	{
		Value: 4,
		Label: "2.0",
	},
	{
		Value: 5,
		Label: "2.5",
	},
	{
		Value: 6,
		Label: "3.0",
	},
	{
		Value: 7,
		Label: "3.5",
	},
	{
		Value: 8,
		Label: "4.0",
	},
	{
		Value: 9,
		Label: "4.5",
	},
	{
		Value: 10,
		Label: "5.0",
	},
}

var moviePageFns = template.FuncMap{
	"dict": func(values ...interface{}) map[string]interface{} {
		if len(values)%2 != 0 {
			panic("dict must have an even number of arguments")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			k, ok := values[i].(string)
			if !ok {
				panic("dict keys must be strings")
			}
			dict[k] = values[i+1]
		}
		return dict
	},
	"relativeCommentDate": relativeCommentDate,
	"relativeWatchDate":   relativeWatchDate,
	"formatReleaseDate": func(t screenjournal.ReleaseDate) string {
		return t.Time().Format("1/2/2006")
	},
	"formatWatchDate":   formatWatchDate,
	"formatCommentTime": formatIso8601Datetime,
	"ratingToStars":     ratingToStars,
	"renderBlurb": func(blurb screenjournal.Blurb) template.HTML {
		return template.HTML(markdown.RenderBlurb(blurb))
	},
	"renderCommentText": func(comment screenjournal.CommentText) template.HTML {
		return template.HTML(markdown.RenderComment(comment))
	},
	"posterPathToURL": posterPathToURL,
}

var reviewPageFns = template.FuncMap{
	"formatDate": func(t time.Time) string {
		return t.Format(time.DateOnly)
	},
}

func (s Server) indexGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/index.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		// Redirect logged in users to the reviews index instead of the landing
		// page.
		if isAuthenticated(r.Context()) {
			http.Redirect(w, r, "/reviews", http.StatusTemporaryRedirect)
			return
		}

		if err := t.Execute(w, struct {
			commonProps
		}{
			commonProps: makeCommonProps(r.Context()),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) aboutGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/about.html")...))
	return func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, struct {
			commonProps
		}{
			commonProps: makeCommonProps(r.Context()),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) logInGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/login.html")...))
	return func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, struct {
			commonProps
		}{
			commonProps: makeCommonProps(r.Context()),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) signUpGet() http.HandlerFunc {
	noInviteTemplate := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/sign-up.html")...))
	byInviteTemplate := template.Must(
		template.New("base.html").
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/sign-up-by-invitation.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		inviteCode, err := inviteCodeFromQueryParams(r)
		if err != nil {
			log.Printf("invalid invite code: %v", err)
			http.Error(w, "Invalid invite code", http.StatusBadRequest)
			return
		}

		var invite screenjournal.SignupInvitation
		if !inviteCode.Empty() {
			invite, err = s.getDB(r).ReadSignupInvitation(inviteCode)
			if err != nil {
				log.Printf("invalid invite code: %v", err)
				http.Error(w, "Invalid invite code", http.StatusUnauthorized)
				return
			}
		}

		uc, err := s.getDB(r).CountUsers()
		if err != nil {
			log.Printf("failed to count users: %v", err)
			http.Error(w, "Failed to load signup template", http.StatusInternalServerError)
			return
		}

		var t *template.Template
		if uc > 0 && invite.Empty() {
			t = byInviteTemplate
		} else {
			t = noInviteTemplate
		}

		var suggestedUsername string
		if !invite.Empty() {
			nonSuggestedCharsPattern := regexp.MustCompile(`(?i)[^a-z0-9]`)
			firstPart := strings.SplitN(invite.Invitee.String(), " ", 2)[0]
			suggestedUsername = nonSuggestedCharsPattern.ReplaceAllString(strings.ToLower(firstPart), "")
		}

		if err := t.Execute(w, struct {
			commonProps
			Invitee           screenjournal.Invitee
			SuggestedUsername string
		}{
			commonProps:       makeCommonProps(r.Context()),
			Invitee:           invite.Invitee,
			SuggestedUsername: suggestedUsername,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) reviewsGet() http.HandlerFunc {
	fns := template.FuncMap{
		"relativeWatchDate": relativeWatchDate,
		"formatWatchDate":   formatWatchDate,
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
		"ratingToStars":   ratingToStars,
		"posterPathToURL": posterPathToURL,
		"renderBlurb": func(blurb screenjournal.Blurb) template.HTML {
			return template.HTML(markdown.RenderBlurb(blurb))
		},
		"splitByNewline": func(s string) []string {
			return strings.Split(s, "\n")
		},
	}

	t := template.Must(
		template.New("base.html").
			Funcs(fns).
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/reviews-index.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		var collectionOwner *screenjournal.Username
		queryOptions := []store.ReadReviewsOption{}
		if username, err := usernameFromRequestPath(r); err == nil {
			collectionOwner = &username
			queryOptions = append(queryOptions, store.FilterReviewsByUsername(username))
		}

		var sortOrder = screenjournal.ByWatchDate
		if sort, err := sortOrderFromQueryParams(r); err == nil {
			sortOrder = sort
			queryOptions = append(queryOptions, store.SortReviews(sort))
		}

		reviews, err := s.getDB(r).ReadReviews(queryOptions...)
		if err != nil {
			log.Printf("failed to read reviews: %v", err)
			http.Error(w, "Failed to read reviews", http.StatusInternalServerError)
			return
		}

		title := "Ratings"
		if collectionOwner != nil {
			title = fmt.Sprintf("%s's %s", collectionOwner, title)
		}

		if err := t.Execute(w, struct {
			commonProps
			Title            string
			Reviews          []screenjournal.Review
			SortOrder        screenjournal.SortOrder
			CollectionOwner  *screenjournal.Username
			UserCanAddReview bool
		}{
			commonProps:      makeCommonProps(r.Context()),
			Title:            title,
			Reviews:          reviews,
			SortOrder:        sortOrder,
			CollectionOwner:  collectionOwner,
			UserCanAddReview: collectionOwner == nil || collectionOwner.Equal(mustGetUsernameFromContext(r.Context())),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) moviesReadGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(moviePageFns).
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/movies-view.html")...))
	return func(w http.ResponseWriter, r *http.Request) {
		mid, err := movieIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid movie ID", http.StatusBadRequest)
			return
		}

		movie, err := s.getDB(r).ReadMovie(mid)
		if err == store.ErrMovieNotFound {
			http.Error(w, "Invalid movie ID", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to read movie metadata: %v", err)
			http.Error(w, "Failed to retrieve movie information", http.StatusInternalServerError)
			return
		}

		reviews, err := s.getDB(r).ReadReviews(store.FilterReviewsByMovieID(mid))
		if err != nil {
			log.Printf("failed to read movie reviews: %v", err)
			http.Error(w, "Failed to retrieve reviews", http.StatusInternalServerError)
			return
		}

		for i, review := range reviews {
			cc, err := s.getDB(r).ReadComments(review.ID)
			if err != nil {
				log.Printf("failed to read reviews comments: %v", err)
				http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
				return
			}
			reviews[i].Comments = cc
		}

		if err := t.Execute(w, struct {
			commonProps
			Movie   screenjournal.Movie
			Reviews []screenjournal.Review
		}{
			commonProps: makeCommonProps(r.Context()),
			Movie:       movie,
			Reviews:     reviews,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) reviewsEditGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(reviewPageFns).
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/reviews-edit.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		id, err := reviewIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, err := s.getDB(r).ReadReview(id)
		if err == store.ErrReviewNotFound {
			http.Error(w, "Invalid review ID", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to read review: %v", err)
			http.Error(w, "Failed to read review", http.StatusInternalServerError)
			return
		}

		loggedInUsername := mustGetUsernameFromContext(r.Context())
		if !review.Owner.Equal(loggedInUsername) {
			http.Error(w, "You can't edit another user's review", http.StatusForbidden)
			return
		}

		if err := t.Execute(w, struct {
			commonProps
			RatingOptions []ratingOption
			Review        screenjournal.Review
			Today         time.Time
		}{
			commonProps:   makeCommonProps(r.Context()),
			RatingOptions: ratingOptions,
			Review:        review,
			Today:         time.Now(),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) reviewsNewGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(reviewPageFns).
			ParseFS(
				templatesFS,
				append(
					baseTemplates,
					"templates/pages/reviews-edit.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		review := screenjournal.Review{}

		// TODO: track movie vs. TV here

		if mid, err := movieIDFromQueryParams(r); err == nil {
			movie, err := s.getDB(r).ReadMovie(mid)
			if err == store.ErrMovieNotFound {
				http.Error(w, "Invalid movie ID", http.StatusNotFound)
				return
			} else if err != nil {
				log.Printf("failed to read movie metadata: %v", err)
				http.Error(w, "Failed to retrieve movie information", http.StatusInternalServerError)
				return
			}
			review.Movie = movie
		} else if err == ErrMovieIDNotProvided {
			// Movie ID is optional for this view.
		} else {
			http.Error(w, "Invalid movie ID", http.StatusBadRequest)
			return
		}

		if tid, err := tmdbIDFromQueryParams(r); err == nil {
			info, err := s.metadataFinder.GetMovieInfo(tid)
			if err != nil {
				http.Error(w, "Failed to get movie info", http.StatusFailedDependency)
				log.Printf("failed to get movie info for TMDB ID %v: %v", tid, err)
				return
			}

			review.Movie.Title = info.Title
			review.Movie.ReleaseDate = info.ReleaseDate
			review.Movie.TmdbID = info.TmdbID
		}
		review.Watched = screenjournal.WatchDate(time.Now())

		if err := t.Execute(w, struct {
			commonProps
			RatingOptions []ratingOption
			Review        screenjournal.Review
			Today         time.Time
		}{
			commonProps:   makeCommonProps(r.Context()),
			RatingOptions: ratingOptions,
			Review:        review,
			Today:         time.Now(),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) invitesGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").ParseFS(
			templatesFS,
			append(
				baseTemplates,
				"templates/fragments/invite-row.html",
				"templates/pages/invites.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		invites, err := s.getDB(r).ReadSignupInvitations()
		if err != nil {
			log.Printf("failed to read signup invitations: %v", err)
			http.Error(w, "Failed to read signup invitations", http.StatusInternalServerError)
			return
		}
		if err := t.Execute(w, struct {
			commonProps
			Invites []screenjournal.SignupInvitation
		}{
			commonProps: makeCommonProps(r.Context()),
			Invites:     invites,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) accountChangePasswordGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").ParseFS(
			templatesFS,
			append(baseTemplates, "templates/pages/account-change-password.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, struct {
			commonProps
		}{
			commonProps: makeCommonProps(r.Context()),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) accountNotificationsGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").ParseFS(
			templatesFS,
			append(baseTemplates, "templates/pages/account-notifications.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		prefs, err := s.getDB(r).ReadNotificationPreferences(mustGetUsernameFromContext(r.Context()))
		if err != nil {
			log.Printf("failed to read notification preferences: %v", err)
			http.Error(w, fmt.Sprintf("failed to read notification preferences: %v", err), http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, struct {
			commonProps
			ReceivesReviewNotices     bool
			ReceivesAllCommentNotices bool
		}{
			commonProps:               makeCommonProps(r.Context()),
			ReceivesReviewNotices:     prefs.NewReviews,
			ReceivesAllCommentNotices: prefs.AllNewComments,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) accountSecurityGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").ParseFS(
			templatesFS,
			append(baseTemplates, "templates/pages/account-security.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, struct {
			commonProps
		}{
			commonProps: makeCommonProps(r.Context()),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func ratingToStars(rating screenjournal.Rating) []string {
	stars := make([]string, parse.MaxRating/2)
	// Add whole stars.
	for i := uint8(0); i < rating.UInt8()/2; i++ {
		stars = append(stars, "fa-solid fa-star")
	}
	if rating.UInt8()%2 != 0 {
		stars = append(stars, "fa-solid fa-star-half-stroke")
	}
	// Add empty stars.
	emptyStars := (parse.MaxRating / 2) - (rating.UInt8() / 2) - rating.UInt8()%2
	for i := uint8(0); i < emptyStars; i++ {
		stars = append(stars, "fa-regular fa-star")
	}
	return stars
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
	if monthsAgo == 1 {
		return "1 month ago"
	}
	return fmt.Sprintf("%d months ago", monthsAgo)
}

func formatWatchDate(t screenjournal.WatchDate) string {
	return t.Time().Format(time.DateOnly)
}

func relativeCommentDate(t time.Time) string {
	minutesAgo := int(time.Since(t).Minutes())
	if minutesAgo < 1 {
		return "just now"
	}
	if minutesAgo == 1 {
		return "a minute ago"
	}
	hoursAgo := int(time.Since(t).Hours())
	if hoursAgo < 1 {
		return fmt.Sprintf("%d minutes ago", minutesAgo)
	}
	if hoursAgo == 1 {
		return "an hour ago"
	}
	if hoursAgo < 24 {
		return fmt.Sprintf("%d hours ago", hoursAgo)
	}

	daysAgo := int(time.Since(t).Hours() / 24)
	weeksAgo := int(daysAgo / 7)
	if daysAgo == 1 {
		return "yesterday"
	} else if daysAgo <= 14 {
		return fmt.Sprintf("%d days ago", daysAgo)
	} else if weeksAgo < 8 {
		return fmt.Sprintf("%d weeks ago", weeksAgo)
	}
	monthsAgo := int(daysAgo / 30)
	if monthsAgo == 1 {
		return "1 month ago"
	}
	return fmt.Sprintf("%d months ago", monthsAgo)
}

func formatIso8601Datetime(t time.Time) string {
	return t.Format("2006-01-02 3:04 pm")
}

func posterPathToURL(pp url.URL) string {
	pp.Scheme = "https"
	pp.Host = "image.tmdb.org"
	pp.Path = "/t/p/w600_and_h900_bestv2" + pp.Path
	return pp.String()
}

func makeCommonProps(ctx context.Context) commonProps {
	username, ok := usernameFromContext(ctx)
	if !ok {
		username = screenjournal.Username("")
	}
	return commonProps{
		IsAuthenticated:  isAuthenticated(ctx),
		IsAdmin:          isAdmin(ctx),
		LoggedInUsername: username,
		CspNonce:         cspNonce(ctx),
	}
}
