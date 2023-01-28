package quiet

import (
	"log"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/announce"
)

type announcer struct {
}

func New() announce.Announcer {
	return announcer{}
}

func (a announcer) AnnounceNewReview(r screenjournal.Review) {
	log.Printf("skipping announcement of review for %s because no announcer is configured", r.Movie.Title)
}
