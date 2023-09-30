package announce

import (
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type Announcer interface {
	AnnounceNewReview(screenjournal.Review)
	AnnounceNewComment(screenjournal.ReviewComment)
}
