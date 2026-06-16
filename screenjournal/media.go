package screenjournal

import (
	"fmt"
	"time"
)

type (
	TmdbID      int32
	ImdbID      string
	ReleaseDate time.Time
)

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

func (rd ReleaseDate) Year() int {
	if rd.Time().IsZero() {
		return 0
	}
	return rd.Time().Year()
}

func (rd ReleaseDate) Time() time.Time {
	return time.Time(rd)
}
