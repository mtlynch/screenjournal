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

	// Rating represents a review rating that can be nil (no rating)
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

// UInt8 returns the uint8 value of a Rating, or 0 if the rating is nil
func (r Rating) UInt8() uint8 {
	if r.Value == nil {
		return 0
	}
	return *r.Value
}

// NewRating creates a new Rating from a uint8 value
func NewRating(val uint8) Rating {
	return Rating{Value: &val}
}

// IsNil returns true if the rating is nil (not set)
func (r Rating) IsNil() bool {
	return r.Value == nil
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
