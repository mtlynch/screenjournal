package email

import (
	"fmt"
	"log"
	"net/mail"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/announce"
	"github.com/mtlynch/screenjournal/v2/email"
)

type (
	UserStore interface {
		ReadUsers() ([]screenjournal.User, error)
	}

	announcer struct {
		baseURL string
		sender  email.Sender
		store   UserStore
	}
)

func New(baseURL string, sender email.Sender, store UserStore) announce.Announcer {
	return announcer{
		baseURL: baseURL,
		sender:  sender,
		store:   store,
	}
}

func (a announcer) AnnounceNewReview(r screenjournal.Review) {
	log.Printf("announcing new review for %+v", r)
	users, err := a.store.ReadUsers()
	if err != nil {
		log.Printf("failed to read announcement recipients from store: %v", err)
	}
	for _, u := range users {
		log.Printf("checking user %s", u.Username.String())
		if u.Username == r.Owner {
			continue
		}
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
			Subject: fmt.Sprintf("New review of %s", r.Movie.Title),
			Date:    time.Now(),
			TextBody: fmt.Sprintf(`A new review is up for %s!

%s/reviews/%d`, r.Movie.Title, a.baseURL, r.ID),
		}
		if err := a.sender.Send(msg); err != nil {
			log.Printf("failed to send message [%s] to recipient [%s]", msg.Subject, msg.To[0].String())
		}
	}
}
