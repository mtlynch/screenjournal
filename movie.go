package screenjournal

import "time"

type (
	ReleaseDate time.Time

	Movie struct {
		ID          MovieID
		TmdbID      TmdbID
		ImdbID      ImdbID
		Title       MediaTitle
		ReleaseDate ReleaseDate
	}
)

func (rd ReleaseDate) Year() int {
	return rd.Time().Year()
}

func (rd ReleaseDate) Time() time.Time {
	return time.Time(rd)
}
