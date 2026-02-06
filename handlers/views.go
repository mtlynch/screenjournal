package handlers

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"time"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/markdown"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type commonProps struct {
	IsAuthenticated  bool
	IsAdmin          bool
	LoggedInUsername screenjournal.Username
	CspNonce         string
}

// reviewViewModel is a template model for displaying a review in
// reviews-for-single-media-entry.html.
// It intentionally exposes only what the template needs.
type reviewViewModel struct {
	ID        screenjournal.ReviewID
	Owner     screenjournal.Username
	Rating    screenjournal.Rating
	Blurb     screenjournal.Blurb
	Watched   screenjournal.WatchDate
	Comments  []screenjournal.ReviewComment
	Reactions []reactionForTemplate
	// CanEdit is true when the logged-in user is allowed to edit the review.
	CanEdit bool
}

func makeReviewViewModel(
	r screenjournal.Review,
	loggedInUsername screenjournal.Username,
	isAdminUser bool,
) reviewViewModel {
	return reviewViewModel{
		ID:        r.ID,
		Owner:     r.Owner,
		Rating:    r.Rating,
		Blurb:     r.Blurb,
		Watched:   r.Watched,
		Comments:  r.Comments,
		Reactions: convertReactionsForTemplate(r.Reactions, loggedInUsername, isAdminUser),
		CanEdit:   r.Owner.Equal(loggedInUsername),
	}
}

func makeReviewViewModels(
	reviews []screenjournal.Review,
	loggedInUsername screenjournal.Username,
	isAdminUser bool,
) []reviewViewModel {
	vms := make([]reviewViewModel, len(reviews))
	for i, r := range reviews {
		vms[i] = makeReviewViewModel(r, loggedInUsername, isAdminUser)
	}
	return vms
}

//go:embed templates
var templatesFS embed.FS

var baseTemplates = []string{
	"templates/layouts/base.html",
	"templates/partials/footer.html",
	"templates/partials/navbar.html",
}

type ratingOption struct {
	Value uint8
	Label string
}

var ratingOptions = []ratingOption{
	{
		Value: 1,
		Label: "0.5",
	},
	{
		Value: 2,
		Label: "1.0",
	},
	{
		Value: 3,
		Label: "1.5",
	},
	{
		Value: 4,
		Label: "2.0",
	},
	{
		Value: 5,
		Label: "2.5",
	},
	{
		Value: 6,
		Label: "3.0",
	},
	{
		Value: 7,
		Label: "3.5",
	},
	{
		Value: 8,
		Label: "4.0",
	},
	{
		Value: 9,
		Label: "4.5",
	},
	{
		Value: 10,
		Label: "5.0",
	},
}

var moviePageFns = template.FuncMap{
	"dict": func(values ...any) map[string]any {
		if len(values)%2 != 0 {
			panic("dict must have an even number of arguments")
		}
		dict := make(map[string]any, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			k, ok := values[i].(string)
			if !ok {
				panic("dict keys must be strings")
			}
			dict[k] = values[i+1]
		}
		return dict
	},
	"relativeCommentDate": relativeCommentDate,
	"relativeWatchDate":   relativeWatchDate,
	"formatReleaseDate": func(t screenjournal.ReleaseDate) string {
		return t.Time().Format("1/2/2006")
	},
	"formatWatchDate":   formatWatchDate,
	"formatCommentTime": formatIso8601Datetime,
	"ratingToStars":     ratingToStars,
	"renderBlurb": func(blurb screenjournal.Blurb) template.HTML {
		return template.HTML(markdown.RenderBlurb(blurb))
	},
	"renderCommentText": func(comment screenjournal.CommentText) template.HTML {
		return template.HTML(markdown.RenderComment(comment))
	},
	"posterPathToURL": posterPathToURL,
}

var reviewPageFns = template.FuncMap{
	"formatDate": func(t time.Time) string {
		return t.Format(time.DateOnly)
	},
}

func ratingToStars(rating screenjournal.Rating) []string {
	if rating.IsNil() {
		return []string{}
	}

	ratingVal := rating.UInt8()
	stars := make([]string, parse.MaxRating/2)
	// Add whole stars.
	for i := uint8(0); i < ratingVal/2; i++ {
		stars = append(stars, "fa-solid fa-star")
	}
	if ratingVal%2 != 0 {
		stars = append(stars, "fa-solid fa-star-half-stroke")
	}
	// Add empty stars.
	emptyStars := (parse.MaxRating / 2) - (ratingVal / 2) - ratingVal%2
	for i := uint8(0); i < emptyStars; i++ {
		stars = append(stars, "fa-regular fa-star")
	}
	return stars
}

func relativeWatchDate(t screenjournal.WatchDate) string {
	daysAgo := int(time.Since(t.Time()).Hours() / 24)
	weeksAgo := int(daysAgo / 7)
	if daysAgo < 1 {
		return "today"
	} else if daysAgo == 1 {
		return "yesterday"
	} else if daysAgo <= 14 {
		return fmt.Sprintf("%d days ago", daysAgo)
	} else if weeksAgo < 8 {
		return fmt.Sprintf("%d weeks ago", weeksAgo)
	}
	monthsAgo := int(daysAgo / 30)
	if monthsAgo == 1 {
		return "1 month ago"
	} else if monthsAgo <= 23 {
		return fmt.Sprintf("%d months ago", monthsAgo)
	}

	yearsAgo := int(math.Round(float64(daysAgo) / 365.0))
	return fmt.Sprintf("%d years ago", yearsAgo)
}

func formatWatchDate(t screenjournal.WatchDate) string {
	return t.Time().Format(time.DateOnly)
}

func relativeCommentDate(t time.Time) string {
	minutesAgo := int(time.Since(t).Minutes())
	if minutesAgo < 1 {
		return "just now"
	}
	if minutesAgo == 1 {
		return "a minute ago"
	}
	hoursAgo := int(time.Since(t).Hours())
	if hoursAgo < 1 {
		return fmt.Sprintf("%d minutes ago", minutesAgo)
	}
	if hoursAgo == 1 {
		return "an hour ago"
	}
	if hoursAgo < 24 {
		return fmt.Sprintf("%d hours ago", hoursAgo)
	}

	daysAgo := int(time.Since(t).Hours() / 24)
	weeksAgo := int(daysAgo / 7)
	if daysAgo == 1 {
		return "yesterday"
	} else if daysAgo <= 14 {
		return fmt.Sprintf("%d days ago", daysAgo)
	} else if weeksAgo < 8 {
		return fmt.Sprintf("%d weeks ago", weeksAgo)
	}

	monthsAgo := int(daysAgo / 30)
	if monthsAgo == 1 {
		return "1 month ago"
	} else if monthsAgo <= 23 {
		return fmt.Sprintf("%d months ago", monthsAgo)
	}

	yearsAgo := int(math.Round(float64(daysAgo) / 365.0))
	return fmt.Sprintf("%d years ago", yearsAgo)
}

func formatIso8601Datetime(t time.Time) string {
	return t.Format("2006-01-02 3:04 pm")
}

func posterPathToURL(pp url.URL) string {
	pp.Scheme = "https"
	pp.Host = "image.tmdb.org"
	pp.Path = "/t/p/w600_and_h900_bestv2" + pp.Path
	return pp.String()
}

func makeCommonProps(ctx context.Context) commonProps {
	username, ok := usernameFromContext(ctx)
	if !ok {
		username = screenjournal.Username("")
	}
	return commonProps{
		IsAuthenticated:  isAuthenticated(ctx),
		IsAdmin:          isAdmin(ctx),
		LoggedInUsername: username,
		CspNonce:         cspNonce(ctx),
	}
}
