package screenjournal

type TmdbID int

func (m TmdbID) Int() int {
	return int(m)
}
