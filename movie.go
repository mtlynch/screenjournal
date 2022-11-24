package screenjournal

import "time"

type ImagePath string

type ReleaseDate time.Time

type Movie struct {
	MediaID      MediaID
	TmdbID       TmdbID
	ImdbID       ImdbID
	Title        MediaTitle
	ReleaseDate  ReleaseDate
	PosterPath   ImagePath
	BackdropPath ImagePath
}

func (p ImagePath) String() string {
	return string(p)
}

func (rd ReleaseDate) Time() time.Time {
	return time.Time(rd)
}
