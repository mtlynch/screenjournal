package email

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"net/mail"
	"path"
	"text/template"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/announce"
	"github.com/mtlynch/screenjournal/v2/email"
	"github.com/mtlynch/screenjournal/v2/markdown"
)

type (
	NotificationsStore interface {
		ReadNotificationSubscribers() ([]screenjournal.User, error)
	}

	announcer struct {
		baseURL string
		sender  email.Sender
		store   NotificationsStore
	}
)

func New(baseURL string, sender email.Sender, store NotificationsStore) announce.Announcer {
	return announcer{
		baseURL: baseURL,
		sender:  sender,
		store:   store,
	}
}

func (a announcer) AnnounceNewReview(r screenjournal.Review) {
	log.Printf("announcing new review from user %s of %s", r.Owner.String(), r.Movie.Title)
	users, err := a.store.ReadNotificationSubscribers()
	if err != nil {
		log.Printf("failed to read announcement recipients from store: %v", err)
	}
	log.Printf("%d user(s) subscribed to new review notifications", len(users))
	for _, u := range users {
		// Don't send a notification to the review author.
		if u.Username == r.Owner {
			continue
		}
		bodyMarkdown := mustRenderTemplate("new-review.tmpl.txt", struct {
			Recipient string
			Title     string
			Author    string
			BaseURL   string
			MovieID   int64
			ReviewID  uint64
		}{
			Recipient: u.Username.String(),
			Title:     r.Movie.Title.String(),
			Author:    r.Owner.String(),
			BaseURL:   a.baseURL,
			MovieID:   r.Movie.ID.Int64(),
			ReviewID:  r.ID.UInt64(),
		})
		bodyHtml := markdown.Render(bodyMarkdown)
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
			Subject:  fmt.Sprintf("%s posted a new review: %s", r.Owner.String(), r.Movie.Title),
			TextBody: bodyMarkdown,
			HtmlBody: bodyHtml,
		}
		if err := a.sender.Send(msg); err != nil {
			log.Printf("failed to send message [%s] to recipient [%s]", msg.Subject, msg.To[0].String())
		}
	}
}

//go:embed templates
var templatesFS embed.FS

func mustRenderTemplate(templateFilename string, templateVars interface{}) string {
	t := template.New(templateFilename)
	t = template.Must(
		t.ParseFS(
			templatesFS,
			path.Join("templates", templateFilename)))

	buf := bytes.NewBuffer([]byte{})
	if err := t.ExecuteTemplate(buf, templateFilename, templateVars); err != nil {
		panic(err)
	}
	return buf.String()
}
