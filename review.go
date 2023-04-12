package screenjournal

import (
	"strconv"
	"time"
)

type (
	ReviewID   uint64
	MediaTitle string
	Rating     uint8
	Blurb      string
	WatchDate  time.Time

	Review struct {
		ID       ReviewID
		Owner    Username
		Rating   Rating
		Blurb    Blurb
		Watched  WatchDate
		Created  time.Time
		Modified time.Time
		Movie    Movie
	}
)

func (id ReviewID) UInt64() uint64 {
	return uint64(id)
}

func (id ReviewID) String() string {
	return strconv.FormatUint(id.UInt64(), 10)
}

func (id ReviewID) IsZero() bool {
	return id == ReviewID(0)
}

func (mt MediaTitle) String() string {
	return string(mt)
}

func (r Rating) UInt8() uint8 {
	return uint8(r)
}

func (r Rating) LessThan(o Rating) bool {
	return r.UInt8() < o.UInt8()
}

func (wd WatchDate) Time() time.Time {
	return time.Time(wd)
}

func (b Blurb) String() string {
	return string(b)
}
