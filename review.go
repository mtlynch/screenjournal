package screenjournal

import "time"

type (
	ReviewID   int
	MediaTitle string
	Rating     uint8
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

func (r Rating) UInt8() uint8 {
	return uint8(r)
}

func (wd WatchDate) Time() time.Time {
	return time.Time(wd)
}
