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
