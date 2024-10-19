package screenjournal

import (
	"net/url"
	"time"
)

type (
	ReleaseDate time.Time

	Movie struct {
		ID          MovieID
		TmdbID      TmdbID
		ImdbID      ImdbID
		Title       MediaTitle
		ReleaseDate ReleaseDate
		PosterPath  url.URL
	}

	// TODO: Move to another file

	TvShowSeason uint8

	TvShow struct {
		TmdbID      TmdbID
		ImdbID      ImdbID
		Title       MediaTitle
		AirDate     ReleaseDate
		SeasonCount uint8
		PosterPath  url.URL
	}
)

func (rd ReleaseDate) Year() int {
	if rd.Time().IsZero() {
		return 0
	}
	return rd.Time().Year()
}

func (rd ReleaseDate) Time() time.Time {
	return time.Time(rd)
}
