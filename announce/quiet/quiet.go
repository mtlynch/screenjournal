package quiet

import (
	"log"

	"github.com/mtlynch/screenjournal/v2/announce"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type announcer struct {
}

func New() announce.Announcer {
	return announcer{}
}

func (a announcer) AnnounceNewReview(r screenjournal.Review) {
	log.Printf("skipping announcement of review for %s because no announcer is configured", r.Movie.Title)
}

func (a announcer) AnnounceNewComment(rc screenjournal.ReviewComment) {
	log.Printf("skipping announcement of new comment from %s about %s's review of %s because no announcer is configured", rc.Owner, rc.Review.Owner, rc.Review.Movie.Title)
}
