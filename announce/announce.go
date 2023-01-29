package announce

import (
	"github.com/mtlynch/screenjournal/v2"
)

type Announcer interface {
	AnnounceNewReview(screenjournal.Review)
}
