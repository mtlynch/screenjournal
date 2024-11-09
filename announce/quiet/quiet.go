package quiet

import (
	"log"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type Announcer struct {
}

func New() Announcer {
	return Announcer{}
}

func (a Announcer) AnnounceNewReview(r screenjournal.Review) {
	log.Printf("skipping announcement of review for %s because no announcer is configured", readMediaTitle(r))
}

func (a Announcer) AnnounceNewComment(rc screenjournal.ReviewComment) {
	log.Printf("skipping announcement of new comment from %s about %s's review of %s because no announcer is configured", rc.Owner, rc.Review.Owner, readMediaTitle(rc.Review))
}

func readMediaTitle(r screenjournal.Review) screenjournal.MediaTitle {
	if !r.Movie.ID.IsZero() {
		return r.Movie.Title
	}
	return r.TvShow.Title
}
