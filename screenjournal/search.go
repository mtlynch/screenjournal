package screenjournal

type SearchQuery string

func (q SearchQuery) String() string {
	return string(q)
}
