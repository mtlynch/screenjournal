package screenjournal

type (
	// MovieID represents the ID for a TV/movie in the local datastore.
	MovieID int64

	TmdbID int32
)

func (m TmdbID) Int32() int32 {
	return int32(m)
}
