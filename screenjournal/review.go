package screenjournal

import (
	"strconv"
	"time"
)

type (
	ReviewID   uint64
	MediaType  string
	MediaTitle string
	Blurb      string
	WatchDate  time.Time

	Rating struct {
		Value *uint8
	}

	Review struct {
		ID           ReviewID
		Owner        Username
		Rating       Rating
		Blurb        Blurb
		Watched      WatchDate
		Created      time.Time
		Modified     time.Time
		Movie        Movie
		TvShow       TvShow
		TvShowSeason TvShowSeason
		Comments     []ReviewComment
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

func (mt MediaType) IsEmpty() bool {
	return mt.String() == ""
}

func (mt MediaType) Equal(o MediaType) bool {
	return mt.String() == o.String()
}

func (mt MediaType) String() string {
	return string(mt)
}

func (mt MediaTitle) String() string {
	return string(mt)
}

func (r Rating) UInt8() *uint8 {
	return r.Value
}

func NewRating(val uint8) Rating {
	return Rating{Value: &val}
}

func (r Rating) IsNil() bool {
	return r.Value == nil
}

func (r Rating) Equal(other Rating) bool {
	if r.IsNil() && other.IsNil() {
		return true
	}
	if r.IsNil() || other.IsNil() {
		return false
	}
	return *r.Value == *other.Value
}

func (wd WatchDate) Time() time.Time {
	return time.Time(wd)
}

func (b Blurb) String() string {
	return string(b)
}

func (r Review) MediaType() MediaType {
	if !r.Movie.ID.IsZero() {
		return MediaTypeMovie
	}
	return MediaTypeTvShow
}
