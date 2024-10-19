package screenjournal

import (
	"strconv"
	"time"
)

type (
	ReviewID   uint64
	MediaType  string
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
		TvShow   TvShow
		Comments []ReviewComment
	}

	ReviewComment struct {
		ID          CommentID
		Owner       Username
		CommentText CommentText
		Created     time.Time
		Modified    time.Time
		Review      Review
	}
)

const (
	MediaTypeMovie  = MediaType("movie")
	MediaTypeTvShow = MediaType("tv-show")
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

func (mt MediaType) String() string {
	return string(mt)
}

func (mt MediaTitle) String() string {
	return string(mt)
}

func (r Rating) UInt8() uint8 {
	return uint8(r)
}

func (wd WatchDate) Time() time.Time {
	return time.Time(wd)
}

func (b Blurb) String() string {
	return string(b)
}
