package screenjournal

type (
	// MediaID represents the ID for a TV/movie in the local datastore.
	MediaID int64

	TmdbID int
	ImdbID string
)

func (m TmdbID) Int() int {
	return int(m)
}
