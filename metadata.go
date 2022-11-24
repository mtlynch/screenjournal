package screenjournal

type (
	// MediaID represents the ID for a TV/movie in the local datastore.
	MediaID int64

	TmdbID int32
	ImdbID string
)

func (m TmdbID) Int32() int32 {
	return int32(m)
}

func (id ImdbID) String() string {
	return string(id)
}
