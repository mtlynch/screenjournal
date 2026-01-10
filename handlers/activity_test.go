package handlers

import (
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestBuildActivityGroupsOrdersItemsAndGroupsByDate(t *testing.T) {
	loc := time.Local
	reviewTime := time.Date(2025, 1, 1, 10, 0, 0, 0, loc)
	commentTime := time.Date(2025, 1, 1, 11, 0, 0, 0, loc)
	reactionTime := time.Date(2025, 1, 1, 12, 0, 0, 0, loc)

	review := screenjournal.Review{
		ID:      screenjournal.ReviewID(12),
		Owner:   screenjournal.Username("mike"),
		Rating:  screenjournal.NewRating(7),
		Movie:   screenjournal.Movie{ID: screenjournal.MovieID(3), Title: screenjournal.MediaTitle("Poker Face")},
		Created: reviewTime,
		Comments: []screenjournal.ReviewComment{
			{
				ID:      screenjournal.CommentID(44),
				Owner:   screenjournal.Username("jamie"),
				Created: commentTime,
			},
		},
		Reactions: []screenjournal.ReviewReaction{
			{
				ID:      screenjournal.ReactionID(55),
				Owner:   screenjournal.Username("joe"),
				Emoji:   screenjournal.NewReactionEmoji("🥞"),
				Created: reactionTime,
			},
		},
	}

	groups := buildActivityGroups([]screenjournal.Review{review})
	if got, want := len(groups), 1; got != want {
		t.Fatalf("expected 1 group, got %d", got)
	}
	if got, want := groups[0].DateLabel, "Jan 1, 2025"; got != want {
		t.Fatalf("unexpected date label: got %q want %q", got, want)
	}

	if got, want := len(groups[0].Items), 3; got != want {
		t.Fatalf("expected 3 activity items, got %d", got)
	}

	if groups[0].Items[0].Kind != activityKindReaction {
		t.Errorf("expected first item to be reaction, got %s", groups[0].Items[0].Kind)
	}
	if groups[0].Items[1].Kind != activityKindComment {
		t.Errorf("expected second item to be comment, got %s", groups[0].Items[1].Kind)
	}
	if groups[0].Items[2].Kind != activityKindReview {
		t.Errorf("expected third item to be review, got %s", groups[0].Items[2].Kind)
	}

	if got, want := groups[0].Items[2].RatingLabel, "3.5"; got != want {
		t.Errorf("unexpected rating label: got %q want %q", got, want)
	}
}

func TestBuildActivityGroupsUsesTvShowSeasonLinks(t *testing.T) {
	loc := time.Local
	commentTime := time.Date(2025, 1, 2, 9, 0, 0, 0, loc)
	reviewTime := time.Date(2025, 1, 1, 9, 0, 0, 0, loc)

	review := screenjournal.Review{
		ID:           screenjournal.ReviewID(7),
		Owner:        screenjournal.Username("dave"),
		Rating:       screenjournal.NewRating(5),
		TvShow:       screenjournal.TvShow{ID: screenjournal.TvShowID(21), Title: screenjournal.MediaTitle("Batman Forever")},
		TvShowSeason: screenjournal.TvShowSeason(2),
		Created:      reviewTime,
		Comments: []screenjournal.ReviewComment{
			{
				ID:      screenjournal.CommentID(99),
				Owner:   screenjournal.Username("jamie"),
				Created: commentTime,
			},
		},
	}

	groups := buildActivityGroups([]screenjournal.Review{review})
	if got, want := len(groups), 2; got != want {
		t.Fatalf("expected 2 groups, got %d", got)
	}

	commentItem := groups[0].Items[0]
	if got, want := commentItem.TargetURL, "/tv-shows/21?season=2#comment99"; got != want {
		t.Errorf("unexpected comment target url: got %q want %q", got, want)
	}

	reviewItem := groups[1].Items[0]
	if got, want := reviewItem.TargetURL, "/tv-shows/21?season=2#review7"; got != want {
		t.Errorf("unexpected review target url: got %q want %q", got, want)
	}
}
