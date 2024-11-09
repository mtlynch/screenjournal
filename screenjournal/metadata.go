package screenjournal

import (
	"fmt"
	"strconv"
)

type (
	// MovieID represents the ID for a movie in the local datastore.
	MovieID int64

	// TvShowID represents the ID for a TV show in the local datastore.
	TvShowID int64

	TmdbID int32
	ImdbID string
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

func (tvID TvShowID) IsZero() bool {
	return tvID.Equal(TvShowID(0))
}

func (tvID TvShowID) Equal(o TvShowID) bool {
	return tvID.Int64() == o.Int64()
}

func (tvID TvShowID) Int64() int64 {
	return int64(tvID)
}

func (tvID TvShowID) String() string {
	return strconv.FormatInt(tvID.Int64(), 10)
}

func (m TmdbID) Equal(o TmdbID) bool {
	return m.Int32() == o.Int32()
}

func (m TmdbID) Int32() int32 {
	return int32(m)
}

func (m TmdbID) String() string {
	return fmt.Sprintf("%d", m)
}

func (id ImdbID) String() string {
	return string(id)
}
