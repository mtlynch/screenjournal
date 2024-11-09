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
		ID          TvShowID
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

func (tvs TvShowSeason) UInt8() uint8 {
	return uint8(tvs)
}

func (tvs TvShowSeason) Equal(o TvShowSeason) bool {
	return tvs.UInt8() == o.UInt8()
}
