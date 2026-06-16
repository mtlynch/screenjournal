package screenjournal

import (
	"net/url"
	"strconv"
)

type (
	// TvShowID represents the ID for a TV show in the local datastore.
	TvShowID int64

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

func (tvs TvShowSeason) UInt8() uint8 {
	return uint8(tvs)
}

func (tvs TvShowSeason) Equal(o TvShowSeason) bool {
	return tvs.UInt8() == o.UInt8()
}
