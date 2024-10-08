package metadata

import (
	"net/url"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	SearchResult struct {
		TmdbID      screenjournal.TmdbID
		ImdbID      screenjournal.ImdbID
		Title       screenjournal.MediaTitle
		ReleaseDate screenjournal.ReleaseDate
		PosterPath  url.URL
	}

	MovieInfo struct {
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
