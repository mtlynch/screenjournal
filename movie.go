package screenjournal

type ImagePath string

type Movie struct {
	MediaID      MediaID
	TmdbID       TmdbID
	ImdbID       ImdbID
	Title        MediaTitle
	PosterPath   ImagePath
	BackdropPath ImagePath
}

func (p ImagePath) String() string {
	return string(p)
}
