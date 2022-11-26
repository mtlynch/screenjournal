package screenjournal

type (
	// MovieID represents the ID for a movie in the local datastore.
	MovieID int64

	TmdbID int32
	ImdbID string
)

func (mid MovieID) Int64() int64 {
	return int64(mid)
}

func (m TmdbID) Int32() int32 {
	return int32(m)
}

func (id ImdbID) String() string {
	return string(id)
}
