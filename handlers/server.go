package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	simple_sessions "codeberg.org/mtlynch/simpleauth/v3/sessions"

	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	Announcer interface {
		AnnounceNewReview(screenjournal.Review)
		AnnounceNewComment(screenjournal.ReviewComment)
	}

	PasswordResetter interface {
		SendEmail(screenjournal.Email) error
		Reset(screenjournal.Username, screenjournal.PasswordResetToken, screenjournal.PasswordHash) error
	}

	SessionManager interface {
		LogIn(context.Context, http.ResponseWriter, simple_sessions.UserID) error
		UserIDFromContext(context.Context) (simple_sessions.UserID, error)
		LogOut(context.Context, http.ResponseWriter) error
		// LoadUser wraps the given handler, adding the user ID (if there's an
		// active session) to the request context before passing control to the
		// next handler.
		LoadUser(http.Handler) http.Handler
	}

	Authenticator interface {
		Authenticate(username screenjournal.Username, password screenjournal.Password) error
	}

	MetadataFinder interface {
		SearchMovies(query screenjournal.SearchQuery) ([]metadata.SearchResult, error)
		SearchTvShows(query screenjournal.SearchQuery) ([]metadata.SearchResult, error)
		GetMovie(id screenjournal.TmdbID) (screenjournal.Movie, error)
		GetTvShow(id screenjournal.TmdbID) (screenjournal.TvShow, error)
	}

	Server struct {
		router           *mux.Router
		authenticator    Authenticator
		announcer        Announcer
		sessionManager   SessionManager
		store            Store
		metadataFinder   MetadataFinder
		passwordResetter PasswordResetter
	}
)

// Router returns the underlying router interface for the server.
func (s Server) Router() *mux.Router {
	return s.router
}

// New creates a new server with all the state it needs to satisfy HTTP
// requests.
func New(authenticator Authenticator, announcer Announcer, sessionManager SessionManager, store Store, metadataFinder MetadataFinder, passwordResetter PasswordResetter) Server {
	s := Server{
		router:           mux.NewRouter(),
		authenticator:    authenticator,
		announcer:        announcer,
		sessionManager:   sessionManager,
		store:            store,
		metadataFinder:   metadataFinder,
		passwordResetter: passwordResetter,
	}

	s.initDev()
	s.routes()
	return s
}
