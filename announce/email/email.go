package email

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"net/mail"
	"path"
	"text/template"

	"github.com/mtlynch/screenjournal/v2/email"
	"github.com/mtlynch/screenjournal/v2/markdown"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	NotificationsStore interface {
		ReadReviewSubscribers() ([]screenjournal.EmailSubscriber, error)
		ReadCommentSubscribers(reviewID screenjournal.ReviewID, commentAuthor screenjournal.Username) ([]screenjournal.EmailSubscriber, error)
	}

	Announcer struct {
		baseURL string
		sender  email.Sender
		store   NotificationsStore
	}
)

func New(baseURL string, sender email.Sender, store NotificationsStore) Announcer {
	return Announcer{
		baseURL: baseURL,
		sender:  sender,
		store:   store,
	}
}

func (a Announcer) AnnounceNewReview(r screenjournal.Review) {
	log.Printf("announcing new review from user %s of %s", r.Owner.String(), r.Movie.Title)
	subscribers, err := a.store.ReadReviewSubscribers()
	if err != nil {
		log.Printf("failed to read announcement recipients from store: %v", err)
		return
	}
	log.Printf("%d user(s) subscribed to new review notifications", len(subscribers))
	for _, subscriber := range subscribers {
		// Don't send a notification to the review author.
		if subscriber.Username.Equal(r.Owner) {
			continue
		}

		var title screenjournal.MediaTitle
		var reviewRoute string
		var seasonSuffix string
		if !r.Movie.ID.IsZero() {
			title = r.Movie.Title
			reviewRoute = fmt.Sprintf("/movies/%d#review%d", r.Movie.ID.Int64(), r.ID.UInt64())
		} else {
			title = r.TvShow.Title
			seasonSuffix = fmt.Sprintf(" (Season %d)", r.TvShowSeason.UInt8())
			reviewRoute = fmt.Sprintf("/tv-shows/%d?season=%d#review%d", r.TvShow.ID.Int64(), r.TvShowSeason.UInt8(), r.ID.UInt64())
		}

		bodyMarkdown := mustRenderTemplate("new-review.tmpl.txt", struct {
			Recipient    string
			Title        string
			SeasonSuffix string
			Author       string
			BaseURL      string
			ReviewRoute  string
		}{
			Recipient:    subscriber.Username.String(),
			Title:        title.String(),
			SeasonSuffix: seasonSuffix,
			Author:       r.Owner.String(),
			BaseURL:      a.baseURL,
			ReviewRoute:  reviewRoute,
		})
		bodyHtml := markdown.RenderEmail(bodyMarkdown)
		msg := email.Message{
			From: mail.Address{
				Name:    "ScreenJournal",
				Address: "activity@thescreenjournal.com",
			},
			To: []mail.Address{
				{
					Name:    subscriber.Username.String(),
					Address: subscriber.Email.String(),
				},
			},
			Subject:  fmt.Sprintf("%s posted a new review: %s%s", r.Owner.String(), title, seasonSuffix),
			TextBody: bodyMarkdown.String(),
			HtmlBody: bodyHtml,
		}
		if err := a.sender.Send(msg); err != nil {
			log.Printf("failed to send message [%s] to recipient [%s]", msg.Subject, msg.To[0].String())
			continue
		}
	}
}

func (a Announcer) AnnounceNewComment(rc screenjournal.ReviewComment) {
	log.Printf("announcing new comment from %s about %s's review of %s", rc.Owner, rc.Review.Owner, rc.Review.Movie.Title)
	users, err := a.store.ReadCommentSubscribers(rc.Review.ID, rc.Owner)
	if err != nil {
		log.Printf("failed to read announcement recipients from store: %v", err)
		return
	}
	log.Printf("%d user(s) are on this thread and accept new comment notifications", len(users))
	for _, u := range users {
		var title screenjournal.MediaTitle
		var commentRoute string
		var seasonSuffix string
		if !rc.Review.Movie.ID.IsZero() {
			title = rc.Review.Movie.Title
			commentRoute = fmt.Sprintf("/movies/%d#comment%d", rc.Review.Movie.ID.Int64(), rc.ID.UInt64())
		} else {
			title = rc.Review.TvShow.Title
			seasonSuffix = fmt.Sprintf(" (Season %d)", rc.Review.TvShowSeason.UInt8())
			commentRoute = fmt.Sprintf("/tv-shows/%d?season=%d#comment%d", rc.Review.TvShow.ID.Int64(), rc.Review.TvShowSeason.UInt8(), rc.ID.UInt64())
		}
		bodyMarkdown := mustRenderTemplate("new-comment.tmpl.txt", struct {
			Recipient     string
			Title         string
			SeasonSuffix  string
			CommentAuthor string
			ReviewAuthor  string
			BaseURL       string
			CommentRoute  string
		}{
			Recipient:     u.Username.String(),
			Title:         title.String(),
			SeasonSuffix:  seasonSuffix,
			CommentAuthor: rc.Owner.String(),
			ReviewAuthor:  rc.Review.Owner.String(),
			BaseURL:       a.baseURL,
			CommentRoute:  commentRoute,
		})
		bodyHtml := markdown.RenderEmail(bodyMarkdown)
		msg := email.Message{
			From: mail.Address{
				Name:    "ScreenJournal",
				Address: "activity@thescreenjournal.com",
			},
			To: []mail.Address{
				{
					Name:    u.Username.String(),
					Address: u.Email.String(),
				},
			},
			Subject:  fmt.Sprintf("%s commented on %s's review of %s%s", rc.Owner.String(), rc.Review.Owner, title, seasonSuffix),
			TextBody: bodyMarkdown.String(),
			HtmlBody: bodyHtml,
		}
		if err := a.sender.Send(msg); err != nil {
			log.Printf("failed to send message [%s] to recipient [%s]", msg.Subject, msg.To[0].String())
			continue
		}
	}
}

//go:embed templates
var templatesFS embed.FS

func mustRenderTemplate(templateFilename string, templateVars any) screenjournal.EmailBodyMarkdown {
	t := template.New(templateFilename)
	t = template.Must(
		t.ParseFS(
			templatesFS,
			path.Join("templates", templateFilename)))

	buf := bytes.NewBuffer([]byte{})
	if err := t.ExecuteTemplate(buf, templateFilename, templateVars); err != nil {
		panic(err)
	}
	return screenjournal.EmailBodyMarkdown(buf.String())
}
