package metadata

import (
	"net/url"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	MovieSearchResult struct {
		TmdbID      screenjournal.TmdbID
		Title       string
		ReleaseDate string
		PosterPath  string
	}

	MovieSearchResults struct {
		Matches      []MovieSearchResult
		Page         int
		TotalPages   int
		TotalResults int
	}

	MovieInfo struct {
		TmdbID      screenjournal.TmdbID
		ImdbID      screenjournal.ImdbID
		Title       screenjournal.MediaTitle
		ReleaseDate screenjournal.ReleaseDate
		PosterPath  url.URL
	}

	TvShowInfo struct {
		TmdbID      screenjournal.TmdbID
		ImdbID      screenjournal.ImdbID
		Title       screenjournal.MediaTitle
		ReleaseDate screenjournal.ReleaseDate
		PosterPath  url.URL
	}
)

func MovieFromMovieInfo(mi MovieInfo) screenjournal.Movie {
	return screenjournal.Movie{
		TmdbID:      mi.TmdbID,
		ImdbID:      mi.ImdbID,
		Title:       mi.Title,
		ReleaseDate: mi.ReleaseDate,
		PosterPath:  mi.PosterPath,
	}
}
