package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

const (
	activityKindReview   = "review"
	activityKindComment  = "comment"
	activityKindReaction = "reaction"
)

type activityItem struct {
	Kind           string
	Created        time.Time
	ActorName      screenjournal.Username
	ActorURL       string
	TargetUserName screenjournal.Username
	TargetUserURL  string
	TargetText     string
	TargetURL      string
	Rating         screenjournal.Rating
	Emoji          string
}

type activityGroup struct {
	DateLabel string
	Items     []activityItem
}

func (s Server) activityGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			Funcs(template.FuncMap{
				"ratingToStars": ratingToStars,
			}).
			ParseFS(
				templatesFS,
				append(baseTemplates, "templates/pages/activity.html")...))

	return func(w http.ResponseWriter, r *http.Request) {
		reviews, err := s.getDB(r).ReadReviews()
		if err != nil {
			log.Printf("failed to read reviews: %v", err)
			http.Error(w, "Failed to load activity", http.StatusInternalServerError)
			return
		}

		for i, review := range reviews {
			rr, err := s.getDB(r).ReadReactions(review.ID)
			if err != nil {
				log.Printf("failed to read reactions for review %s: %v", review.ID, err)
				http.Error(w, "Failed to load activity", http.StatusInternalServerError)
				return
			}
			reviews[i].Reactions = rr
		}

		if err := t.Execute(w, struct {
			commonProps
			Groups []activityGroup
		}{
			commonProps: makeCommonProps(r.Context()),
			Groups:      buildActivityGroups(reviews),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func buildActivityGroups(reviews []screenjournal.Review) []activityGroup {
	items := []activityItem{}
	for _, review := range reviews {
		reviewTargetText := reviewMediaTitle(review)
		reviewURL := reviewTargetURL(review, review.ID)
		items = append(items, activityItem{
			Kind:       activityKindReview,
			Created:    review.Created,
			ActorName:  review.Owner,
			ActorURL:   userReviewsURL(review.Owner),
			TargetText: reviewTargetText,
			TargetURL:  reviewURL,
			Rating:     review.Rating,
		})

		for _, comment := range review.Comments {
			items = append(items, activityItem{
				Kind:           activityKindComment,
				Created:        comment.Created,
				ActorName:      comment.Owner,
				ActorURL:       userReviewsURL(comment.Owner),
				TargetUserName: review.Owner,
				TargetUserURL:  userReviewsURL(review.Owner),
				TargetText:     fmt.Sprintf("review of %s", reviewTargetText),
				TargetURL:      reviewCommentURL(review, comment.ID),
			})
		}

		for _, reaction := range review.Reactions {
			items = append(items, activityItem{
				Kind:           activityKindReaction,
				Created:        reaction.Created,
				ActorName:      reaction.Owner,
				ActorURL:       userReviewsURL(reaction.Owner),
				TargetUserName: review.Owner,
				TargetUserURL:  userReviewsURL(review.Owner),
				TargetText:     reviewTargetText,
				TargetURL:      reviewTargetURL(review, review.ID),
				Emoji:          reaction.Emoji.String(),
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Created.After(items[j].Created)
	})

	groups := []activityGroup{}
	var currentGroup *activityGroup
	var currentDateKey string

	for _, item := range items {
		dateKey := activityDateKey(item.Created)
		if currentGroup == nil || dateKey != currentDateKey {
			group := activityGroup{
				DateLabel: formatActivityDate(item.Created),
				Items:     []activityItem{},
			}
			groups = append(groups, group)
			currentGroup = &groups[len(groups)-1]
			currentDateKey = dateKey
		}
		currentGroup.Items = append(currentGroup.Items, item)
	}

	return groups
}

func activityDateKey(t time.Time) string {
	local := t.In(time.Local)
	return fmt.Sprintf("%04d-%02d-%02d", local.Year(), local.Month(), local.Day())
}

func formatActivityDate(t time.Time) string {
	return t.In(time.Local).Format("Jan 2, 2006")
}

func userReviewsURL(username screenjournal.Username) string {
	return fmt.Sprintf("/reviews/by/%s", username)
}

func reviewMediaTitle(review screenjournal.Review) string {
	if !review.Movie.ID.IsZero() {
		return review.Movie.Title.String()
	}
	title := review.TvShow.Title.String()
	if review.TvShowSeason.UInt8() == 0 {
		return title
	}
	return fmt.Sprintf("%s season %d", title, review.TvShowSeason.UInt8())
}

func reviewPageURL(review screenjournal.Review) string {
	if !review.Movie.ID.IsZero() {
		return fmt.Sprintf("/movies/%s", review.Movie.ID.String())
	}
	base := fmt.Sprintf("/tv-shows/%s", review.TvShow.ID.String())
	if review.TvShowSeason.UInt8() == 0 {
		return base
	}
	return fmt.Sprintf("%s?season=%d", base, review.TvShowSeason.UInt8())
}

func reviewTargetURL(review screenjournal.Review, id screenjournal.ReviewID) string {
	return fmt.Sprintf("%s#review%s", reviewPageURL(review), id.String())
}

func reviewCommentURL(review screenjournal.Review, id screenjournal.CommentID) string {
	return fmt.Sprintf("%s#comment%d", reviewPageURL(review), id.UInt64())
}
