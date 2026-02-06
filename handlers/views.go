package handlers

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
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

// mediaViewStub holds the media-specific fields passed to the
// reviews-for-single-media-entry.html template.
type mediaViewStub struct {
	Type         screenjournal.MediaType
	IsTvShow     bool
	ID           int64
	Title        screenjournal.MediaTitle
	SeasonNumber screenjournal.TvShowSeason
	PosterPath   url.URL
	ImdbID       screenjournal.ImdbID
	TmdbID       screenjournal.TmdbID
	ReleaseDate  screenjournal.ReleaseDate
}

// reviewViewModel is a template model for displaying a review in
// reviews-for-single-media-entry.html.
// It intentionally exposes only what the template needs.
type reviewViewModel struct {
	ID        screenjournal.ReviewID
	Owner     screenjournal.Username
	Rating    screenjournal.Rating
	Blurb     screenjournal.Blurb
	Watched   screenjournal.WatchDate
	Comments  []screenjournal.ReviewComment
	Reactions []reactionForTemplate
	// CanEdit is true when the logged-in user is allowed to edit the review.
	CanEdit bool
}

func makeReviewViewModel(
	r screenjournal.Review,
	loggedInUsername screenjournal.Username,
	isAdminUser bool,
) reviewViewModel {
	return reviewViewModel{
		ID:        r.ID,
		Owner:     r.Owner,
		Rating:    r.Rating,
		Blurb:     r.Blurb,
		Watched:   r.Watched,
		Comments:  r.Comments,
		Reactions: convertReactionsForTemplate(r.Reactions, loggedInUsername, isAdminUser),
		CanEdit:   r.Owner.Equal(loggedInUsername),
	}
}

func makeReviewViewModels(
	reviews []screenjournal.Review,
	loggedInUsername screenjournal.Username,
	isAdminUser bool,
) []reviewViewModel {
	vms := make([]reviewViewModel, len(reviews))
	for i, r := range reviews {
		vms[i] = makeReviewViewModel(r, loggedInUsername, isAdminUser)
	}
	return vms
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
			plaintext := markdown.RenderBlurbAsPlaintext(b)
			var elidedChars []rune
			for _, c := range plaintext {
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

// renderMediaReviews is a shared helper that loads comments/reactions for each
// review, converts them to view models, and renders the
// reviews-for-single-media-entry template.
func (s Server) renderMediaReviews(w http.ResponseWriter, r *http.Request, t *template.Template, media mediaViewStub, reviews []screenjournal.Review) {
	loggedInUsername := mustGetUsernameFromContext(r.Context())
	isAdminUser := isAdmin(r.Context())

	for i, review := range reviews {
		cc, err := s.getDB(r).ReadComments(review.ID)
		if err != nil {
			log.Printf("failed to read reviews comments: %v", err)
			http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
			return
		}
		reviews[i].Comments = cc

		rr, err := s.getDB(r).ReadReactions(review.ID)
		if err != nil {
			log.Printf("failed to read reviews reactions: %v", err)
			http.Error(w, "Failed to retrieve reactions", http.StatusInternalServerError)
			return
		}
		reviews[i].Reactions = rr
	}

	reviewsForTemplate := makeReviewViewModels(reviews, loggedInUsername, isAdminUser)

	if err := t.Execute(w, struct {
		commonProps
		Media           mediaViewStub
		Reviews         []reviewViewModel
		AvailableEmojis []screenjournal.ReactionEmoji
	}{
		commonProps:     makeCommonProps(r.Context()),
		Media:           media,
		Reviews:         reviewsForTemplate,
		AvailableEmojis: screenjournal.AllowedReactionEmojis(),
	}); err != nil {
		log.Printf("failed to render template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s Server) moviesReadGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(moviePageFns).
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/reviews-for-single-media-entry.html")...))
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

		s.renderMediaReviews(w, r, t, mediaViewStub{
			Type:        screenjournal.MediaTypeMovie,
			IsTvShow:    false,
			ID:          movie.ID.Int64(),
			Title:       movie.Title,
			PosterPath:  movie.PosterPath,
			ImdbID:      movie.ImdbID,
			TmdbID:      movie.TmdbID,
			ReleaseDate: movie.ReleaseDate,
		}, reviews)
	}
}

func (s Server) tvShowsReadGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(moviePageFns).
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/reviews-for-single-media-entry.html")...))
	return func(w http.ResponseWriter, r *http.Request) {
		tvID, err := tvShowIDFromRequestPath(r)
		if err != nil {
			http.Error(w, "Invalid TV show ID", http.StatusBadRequest)
			return
		}

		seasonNumber, err := tvShowSeasonFromQueryParams(r)
		if err != nil {
			log.Printf("invalid TV show season: %v", err)
			http.Error(w, "Invalid TV show season", http.StatusBadRequest)
			return
		}

		tvShow, err := s.getDB(r).ReadTvShow(tvID)
		if err == store.ErrTvShowNotFound {
			http.Error(w, "Invalid TV show ID", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to read TV show metadata: %v", err)
			http.Error(w, "Failed to retrieve TV show information", http.StatusInternalServerError)
			return
		}

		reviews, err := s.getDB(r).ReadReviews(store.FilterReviewsByTvShowID(tvID), store.FilterReviewsByTvShowSeason(seasonNumber))
		if err != nil {
			log.Printf("failed to read TV show reviews: %v", err)
			http.Error(w, "Failed to retrieve TV show reviews", http.StatusInternalServerError)
			return
		}

		s.renderMediaReviews(w, r, t, mediaViewStub{
			Type:         screenjournal.MediaTypeTvShow,
			IsTvShow:     true,
			ID:           tvShow.ID.Int64(),
			Title:        tvShow.Title,
			SeasonNumber: seasonNumber,
			PosterPath:   tvShow.PosterPath,
			ImdbID:       tvShow.ImdbID,
			TmdbID:       tvShow.TmdbID,
			ReleaseDate:  tvShow.AirDate,
		}, reviews)
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

		var mediaType screenjournal.MediaType
		if review.Movie.ID.Int64() == 0 {
			mediaType = screenjournal.MediaTypeMovie
		} else {
			mediaType = screenjournal.MediaTypeTvShow
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
			MediaType     screenjournal.MediaType
			Today         time.Time
		}{
			commonProps:   makeCommonProps(r.Context()),
			RatingOptions: ratingOptions,
			Review:        review,
			MediaType:     mediaType,
			Today:         time.Now(),
		}); err != nil {
			log.Printf("failed to execute template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) reviewsNewTitleSearchGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(reviewPageFns).
			ParseFS(
				templatesFS,
				append(
					baseTemplates,
					"templates/pages/reviews-new.html")...))

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

func (s Server) reviewsNewPickSeasonGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(reviewPageFns).
			ParseFS(
				templatesFS,
				append(
					baseTemplates,
					"templates/pages/reviews-tv-pick-season.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		tmdbID, err := tmdbIDFromQueryParams(r)
		if err != nil {
			log.Printf("invalid TMDB ID: %v", err)
			http.Error(w, "Invalid TMDB ID", http.StatusBadRequest)
		}

		// Even if the show is in the local datastore, get the latest metadata from
		// TMDB, as there could be new seasons since the last cache.
		var tvShowID *screenjournal.TvShowID = nil
		tvShow, err := s.getTvShow(r, tvShowID, &tmdbID)
		if err != nil {
			http.Error(w, "Failed to get TV show info", http.StatusFailedDependency)
			log.Printf("failed to get TV show info with, TMDB ID=%v: %v", tmdbID, err)
			return
		}

		if tvShow.SeasonCount == 1 {
			http.Redirect(w, r, fmt.Sprintf("/reviews/new/write?mediaType=tv-show&tmdbId=%d&season=1", tvShow.TmdbID.Int32()), http.StatusSeeOther)
			return
		}

		seasonOptions := make([]uint8, tvShow.SeasonCount)
		for i := uint8(0); i < tvShow.SeasonCount; i++ {
			// Add 1 so that Season 1 doesn't show as Season 0, etc.
			seasonOptions[i] = i + 1
		}

		if err := t.Execute(w, struct {
			commonProps
			TmdbID        screenjournal.TmdbID
			TvShowTitle   screenjournal.MediaTitle
			ReleaseYear   int
			SeasonOptions []uint8
		}{
			commonProps:   makeCommonProps(r.Context()),
			TmdbID:        tmdbID,
			TvShowTitle:   tvShow.Title,
			ReleaseYear:   tvShow.AirDate.Year(),
			SeasonOptions: seasonOptions,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) reviewsNewWriteReviewGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(reviewPageFns).
			ParseFS(
				templatesFS,
				append(
					baseTemplates,
					"templates/pages/reviews-edit.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		var mediaType screenjournal.MediaType

		// Review must have either a movieID (cached information in the database),
		// or a TMDB ID (no movie info cached yet).

		var movieID *screenjournal.MovieID
		if mid, err := movieIDFromQueryParams(r); err == ErrMovieIDNotProvided {
			// It's okay for the movie ID to be absent, as it's optional.
		} else if err != nil {
			log.Printf("invalid movie ID: %v", err)
			http.Error(w, "Invalid movie ID", http.StatusBadRequest)
			return
		} else {
			movieID = &mid
			mediaType = screenjournal.MediaTypeMovie
		}

		var tvShowID *screenjournal.TvShowID
		tvid, err := tvShowIDFromQueryParams(r)
		if err == ErrTvShowIDNotProvided {
			// It's okay for the TV show ID to be absent, as it's optional.
		} else if err != nil {
			log.Printf("invalid TV show ID: %v", err)
			http.Error(w, "Invalid TV show ID", http.StatusBadRequest)
			return
		} else {
			tvShowID = &tvid
			mediaType = screenjournal.MediaTypeTvShow
		}

		var tmdbID *screenjournal.TmdbID
		tid, err := tmdbIDFromQueryParams(r)
		if err == ErrTmdbIDNotProvided {
			// It's okay for the TMDB ID to be absent, as it's optional.
		} else if err != nil {
			log.Printf("invalid TMDB ID: %v", err)
			http.Error(w, "Invalid TMDB ID", http.StatusBadRequest)
			return
		} else {
			tmdbID = &tid
		}

		// If we can't infer the media type from other query params, check for an
		// media type query param.
		if mediaType.IsEmpty() {
			if mediaType, err = mediaTypeFromQueryParams(r); err != nil {
				log.Printf("invalid media type: %v", err)
				http.Error(w, "Invalid media type", http.StatusBadRequest)
				return
			}
		}

		var movie screenjournal.Movie
		var tvShow screenjournal.TvShow
		var tvShowSeason screenjournal.TvShowSeason
		if mediaType == screenjournal.MediaTypeMovie {
			m, err := s.getMovie(r, movieID, tmdbID)
			if err != nil {
				http.Error(w, "Failed to get movie info", http.StatusFailedDependency)
				log.Printf("failed to get movie info with movie ID=%v, TMDB ID=%v: %v", movieID, tmdbID, err)
				return
			}
			movie = m
		} else if mediaType == screenjournal.MediaTypeTvShow {
			t, err := s.getTvShow(r, tvShowID, tmdbID)
			if err != nil {
				http.Error(w, "Failed to get TV show info", http.StatusFailedDependency)
				log.Printf("failed to get TV show info with, TMDB ID=%v: %v", tmdbID, err)
				return
			}
			tvShow = t

			season, err := tvShowSeasonFromQueryParams(r)
			if err != nil {
				http.Error(w, "Invalid TV show season", http.StatusBadRequest)
				log.Printf("invalid TV show season: %v", err)
				return
			}
			tvShowSeason = season
		}

		if err := t.Execute(w, struct {
			commonProps
			RatingOptions []ratingOption
			Review        screenjournal.Review
			MediaType     screenjournal.MediaType
			Today         time.Time
		}{
			commonProps:   makeCommonProps(r.Context()),
			RatingOptions: ratingOptions,
			Review: screenjournal.Review{
				Movie:        movie,
				TvShow:       tvShow,
				TvShowSeason: tvShowSeason,
				Watched:      screenjournal.WatchDate(time.Now()),
			},
			MediaType: mediaType,
			Today:     time.Now(),
		}); err != nil {
			log.Printf("failed to execute template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s Server) getMovie(r *http.Request, movieID *screenjournal.MovieID, tmdbID *screenjournal.TmdbID) (screenjournal.Movie, error) {
	// Try to get the movie information from the database.
	if movieID != nil {
		m, err := s.getDB(r).ReadMovie(*movieID)
		if err != nil {
			return screenjournal.Movie{}, err
		}
		return m, nil
	}

	// If we can't read the movie information from the database, use the TMDB ID
	// to get information from TMDB.
	if tmdbID != nil {
		return s.metadataFinder.GetMovie(*tmdbID)
	}

	return screenjournal.Movie{}, errors.New("need movie ID or TMDB ID to retrieve movie metadata")
}

func (s Server) getTvShow(r *http.Request, tvShowID *screenjournal.TvShowID, tmdbID *screenjournal.TmdbID) (screenjournal.TvShow, error) {
	// Try to get the TV show information from the database.
	if tvShowID != nil {
		t, err := s.getDB(r).ReadTvShow(*tvShowID)
		if err != nil {
			return screenjournal.TvShow{}, err
		}
		return t, nil
	}

	if tmdbID != nil {
		return s.metadataFinder.GetTvShow(*tmdbID)
	}

	return screenjournal.TvShow{}, errors.New("need TMDB ID to retrieve TV show metadata")
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

func (s Server) passwordResetAdminGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(template.FuncMap{
				"formatTime": func(t time.Time) string {
					return t.Format("Jan 2, 2006 3:04 PM")
				},
			}).
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/admin-reset-password.html", "templates/fragments/password-reset-row.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		allUsers, err := s.getDB(r).ReadUsersPublicMeta()
		if err != nil {
			log.Printf("failed to read users: %v", err)
			http.Error(w, "Failed to load users", http.StatusInternalServerError)
			return
		}

		// Filter the current user from the list.
		currentUsername := mustGetUsernameFromContext(r.Context())
		var users []screenjournal.UserPublicMeta
		for _, user := range allUsers {
			if !user.Username.Equal(currentUsername) {
				users = append(users, user)
			}
		}

		// Clean up expired tokens before displaying.
		if err := s.getDB(r).DeleteExpiredPasswordResetEntries(); err != nil {
			log.Printf("failed to clean up expired password reset tokens: %v", err)
		}

		passwordResetEntries, err := s.getDB(r).ReadPasswordResetEntries()
		if err != nil {
			log.Printf("failed to read password reset requests: %v", err)
			http.Error(w, "Failed to load password reset requests", http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, struct {
			commonProps
			passwordResetAdminGetRequest
		}{
			commonProps: makeCommonProps(r.Context()),
			passwordResetAdminGetRequest: passwordResetAdminGetRequest{
				Users:                users,
				PasswordResetEntries: passwordResetEntries,
			},
		}); err != nil {
			log.Printf("failed to render admin reset password template: %v", err)
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
	}
}

func (s Server) accountPasswordResetGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").ParseFS(
			templatesFS,
			append(baseTemplates, "templates/pages/account-change-password.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		// If a token is provided, validate it before rendering the page.
		token, err := parse.PasswordResetToken(r.URL.Query().Get("token"))
		if err != nil {
			http.Error(w, "Invalid password reset token", http.StatusUnauthorized)
			return
		}

		// Verify token exists and hasn't expired.
		passwordResetEntry, err := s.getDB(r).ReadPasswordResetEntry(token)
		if err != nil {
			http.Error(w, "Invalid or expired password reset token", http.StatusUnauthorized)
			return
		}

		if passwordResetEntry.IsExpired() {
			// Clean up expired token.
			if err := s.getDB(r).DeletePasswordResetEntry(token); err != nil {
				log.Printf("failed to delete expired password reset token %s: %v", token, err)
			}
			http.Error(w, "Password reset token has expired", http.StatusUnauthorized)
			return
		}

		if err := t.Execute(w, struct {
			commonProps
			Token         string
			FormTargetURL string
			CancelURL     string
		}{
			commonProps:   makeCommonProps(r.Context()),
			Token:         token.String(),
			FormTargetURL: fmt.Sprintf("/account/password-reset?token=%s", token),
			CancelURL:     "/login",
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
			Token         string
			FormTargetURL string
			CancelURL     string
		}{
			commonProps:   makeCommonProps(r.Context()),
			Token:         "",
			FormTargetURL: "/account/password",
			CancelURL:     "/account/security",
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

func (s Server) usersGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(reviewPageFns).
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/users.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.getDB(r).ReadUsersPublicMeta()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, struct {
			commonProps
			Users []screenjournal.UserPublicMeta
		}{
			commonProps: makeCommonProps(r.Context()),
			Users:       users,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func ratingToStars(rating screenjournal.Rating) []string {
	if rating.IsNil() {
		return []string{}
	}

	ratingVal := rating.UInt8()
	stars := make([]string, parse.MaxRating/2)
	// Add whole stars.
	for i := uint8(0); i < ratingVal/2; i++ {
		stars = append(stars, "fa-solid fa-star")
	}
	if ratingVal%2 != 0 {
		stars = append(stars, "fa-solid fa-star-half-stroke")
	}
	// Add empty stars.
	emptyStars := (parse.MaxRating / 2) - (ratingVal / 2) - ratingVal%2
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
	} else if monthsAgo <= 23 {
		return fmt.Sprintf("%d months ago", monthsAgo)
	}

	yearsAgo := int(math.Round(float64(daysAgo) / 365.0))
	return fmt.Sprintf("%d years ago", yearsAgo)
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
	} else if monthsAgo <= 23 {
		return fmt.Sprintf("%d months ago", monthsAgo)
	}

	yearsAgo := int(math.Round(float64(daysAgo) / 365.0))
	return fmt.Sprintf("%d years ago", yearsAgo)
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
