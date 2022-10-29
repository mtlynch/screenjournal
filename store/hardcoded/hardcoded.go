package hardcoded

import (
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/store"
)

type hardcodedStore struct {
	Reviews []screenjournal.Review
}

func New() store.Store {
	return hardcodedStore{
		Reviews: []screenjournal.Review{
			{
				ID:      screenjournal.ReviewID(3),
				Owner:   screenjournal.Username("mike"),
				Title:   screenjournal.MediaTitle("The Little Hours"),
				Rating:  screenjournal.Rating(5),
				Blurb:   screenjournal.Blurb("Great cast. It had a lot of non sequiturs, so I'm wondering if it would have made more sense if I'd read The Decameron. Mostly enjoyed it, but it kind of goes off the rails by the end."),
				Watched: mustParseWatchDate("2022-10-27T00:00:00-05:00"),
			},
			{
				ID:      screenjournal.ReviewID(2),
				Owner:   screenjournal.Username("mike"),
				Title:   screenjournal.MediaTitle("Joshy"),
				Rating:  screenjournal.Rating(7),
				Blurb:   screenjournal.Blurb("It had its ups and downs, but there were a lot of good jokes, good cast, original story."),
				Watched: mustParseWatchDate("2022-10-23T00:00:00-05:00"),
			},
			{
				ID:      screenjournal.ReviewID(1),
				Owner:   screenjournal.Username("mike"),
				Title:   screenjournal.MediaTitle("Hello, My Name is Doris"),
				Rating:  screenjournal.Rating(5),
				Blurb:   screenjournal.Blurb("Great first half. Second half fell apart."),
				Watched: mustParseWatchDate("2022-08-21T00:00:00-05:00"),
			},
		},
	}
}

func (hs hardcodedStore) ReadReviews() ([]screenjournal.Review, error) {
	return hs.Reviews, nil
}

func mustParseWatchDate(s string) screenjournal.WatchDate {
	t, err := time.Parse("2006-01-02T15:04:05-07:00", s)
	if err != nil {
		panic(err)
	}
	return screenjournal.WatchDate(t)
}
