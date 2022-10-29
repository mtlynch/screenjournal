package screenjournal

import "time"

type (
	ReviewID   int
	MediaTitle string
	Rating     int8
	Blurb      string
	WatchDate  time.Time

	Review struct {
		ID       ReviewID
		Owner    Username
		Title    MediaTitle
		Rating   Rating
		Blurb    Blurb
		Watched  WatchDate
		Created  time.Time
		Modified time.Time
	}
)

func (id ReviewID) Int() int {
	return int(id)
}

func (r Rating) Int8() int8 {
	return int8(r)
}

func (wd WatchDate) Time() time.Time {
	return time.Time(wd)
}
