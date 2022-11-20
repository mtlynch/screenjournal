package screenjournal

type (
	TmdbID int32
)

func (m TmdbID) Int32() int32 {
	return int32(m)
}
