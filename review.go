package screenjournal

import "time"

type (
	ReviewID   int
	MediaTitle string
	Rating     int8
	Blurb      string
	WatchDate  time.Time

	Review struct {
		ID      ReviewID
		Owner   Username
		Title   MediaTitle
		Rating  Rating
		Blurb   Blurb
		Watched WatchDate
	}
)

func (wd WatchDate) Time() time.Time {
	return time.Time(wd)
}
