package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mtlynch/screenjournal/v2/markdown"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

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

		// Convert reviews to view models for templates.
		reviewsForTemplate := makeReviewViewModels(reviews, loggedInUsername, isAdminUser)

		type mediaStub struct {
			IsTvShow     bool
			Type         screenjournal.MediaType
			ID           int64
			Title        screenjournal.MediaTitle
			SeasonNumber screenjournal.TvShowSeason
			PosterPath   url.URL
			ImdbID       screenjournal.ImdbID
			TmdbID       screenjournal.TmdbID
			ReleaseDate  screenjournal.ReleaseDate
		}
		if err := t.Execute(w, struct {
			commonProps
			Media           mediaStub
			Reviews         []reviewViewModel
			AvailableEmojis []screenjournal.ReactionEmoji
		}{
			commonProps: makeCommonProps(r.Context()),
			Media: mediaStub{
				Type:        screenjournal.MediaTypeMovie,
				IsTvShow:    false,
				ID:          movie.ID.Int64(),
				Title:       movie.Title,
				PosterPath:  movie.PosterPath,
				ImdbID:      movie.ImdbID,
				TmdbID:      movie.TmdbID,
				ReleaseDate: movie.ReleaseDate,
			},
			Reviews:         reviewsForTemplate,
			AvailableEmojis: screenjournal.AllowedReactionEmojis(),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

		// Convert reviews to view models for templates.
		reviewsForTemplate := makeReviewViewModels(reviews, loggedInUsername, isAdminUser)

		type mediaStub struct {
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

		if err := t.Execute(w, struct {
			commonProps
			Media           mediaStub
			Reviews         []reviewViewModel
			AvailableEmojis []screenjournal.ReactionEmoji
		}{
			commonProps: makeCommonProps(r.Context()),
			Media: mediaStub{
				Type:         screenjournal.MediaTypeTvShow,
				IsTvShow:     true,
				ID:           tvShow.ID.Int64(),
				Title:        tvShow.Title,
				SeasonNumber: seasonNumber,
				PosterPath:   tvShow.PosterPath,
				ImdbID:       tvShow.ImdbID,
				TmdbID:       tvShow.TmdbID,
				ReleaseDate:  tvShow.AirDate,
			},
			Reviews:         reviewsForTemplate,
			AvailableEmojis: screenjournal.AllowedReactionEmojis(),
		}); err != nil {
			log.Printf("failed to render template: %v", err)
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
