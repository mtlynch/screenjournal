package metadata

import (
	"net/url"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	SearchResult struct {
		TmdbID      screenjournal.TmdbID
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

	TvShowInfo struct {
		TmdbID      screenjournal.TmdbID
		ImdbID      screenjournal.ImdbID
		Title       screenjournal.MediaTitle
		ReleaseDate screenjournal.ReleaseDate
		SeasonCount uint8
		PosterPath  url.URL
	}
)

func MovieFromMovieInfo(info MovieInfo) screenjournal.Movie {
	return screenjournal.Movie{
		TmdbID:      info.TmdbID,
		ImdbID:      info.ImdbID,
		Title:       info.Title,
		ReleaseDate: info.ReleaseDate,
		PosterPath:  info.PosterPath,
	}
}

func TvShowFromTvShowInfo(info TvShowInfo) screenjournal.TvShow {
	return screenjournal.TvShow{
		TmdbID:     info.TmdbID,
		ImdbID:     info.ImdbID,
		Title:      info.Title,
		AirDate:    info.ReleaseDate,
		PosterPath: info.PosterPath,
	}
}
