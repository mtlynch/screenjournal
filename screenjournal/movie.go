package screenjournal

import (
	"net/url"
	"strconv"
)

type (
	// MovieID represents the ID for a movie in the local datastore.
	MovieID int64

	Movie struct {
		ID          MovieID
		TmdbID      TmdbID
		ImdbID      ImdbID
		Title       MediaTitle
		ReleaseDate ReleaseDate
		PosterPath  url.URL
	}
)

func (mid MovieID) IsZero() bool {
	return mid.Equal(MovieID(0))
}

func (mid MovieID) Equal(o MovieID) bool {
	return mid.Int64() == o.Int64()
}

func (mid MovieID) Int64() int64 {
	return int64(mid)
}

func (mid MovieID) String() string {
	return strconv.FormatInt(mid.Int64(), 10)
}
