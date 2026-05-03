package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mtlynch/screenjournal/v2/handlers/sessions"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/sqlite"
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
		CreateSession(http.ResponseWriter, context.Context, screenjournal.Username, bool) error
		SessionFromContext(context.Context) (sessions.Session, error)
		EndSession(context.Context, http.ResponseWriter)
		// WrapRequest wraps the given handler, adding the Session object
		// (if there's an active session) to the request context before
		// passing control to the next handler.
		WrapRequest(http.Handler) http.Handler
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

	ServerParams struct {
		Authenticator    Authenticator
		Announcer        Announcer
		SessionManager   SessionManager
		Store            sqlite.Store
		MetadataFinder   MetadataFinder
		PasswordResetter PasswordResetter
	}

	Server struct {
		router           *mux.Router
		authenticator    Authenticator
		announcer        Announcer
		sessionManager   SessionManager
		store            sqlite.Store
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
func New(params ServerParams) Server {
	s := Server{
		router:           mux.NewRouter(),
		authenticator:    params.Authenticator,
		announcer:        params.Announcer,
		sessionManager:   params.SessionManager,
		store:            params.Store,
		metadataFinder:   params.MetadataFinder,
		passwordResetter: params.PasswordResetter,
	}

	s.routes()
	return s
}
