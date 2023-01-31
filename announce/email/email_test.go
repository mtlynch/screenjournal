package email_test

import (
	"net/mail"
	"reflect"
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	email_announce "github.com/mtlynch/screenjournal/v2/announce/email"
	"github.com/mtlynch/screenjournal/v2/email"
)

type mockUserStore struct {
	users []screenjournal.User
}

func (us mockUserStore) ReadUsers() ([]screenjournal.User, error) {
	return us.users, nil
}

type mockEmailSender struct {
	emailsSent []email.Message
}

func (s *mockEmailSender) Send(message email.Message) error {
	s.emailsSent = append(s.emailsSent, message)
	return nil
}

func TestAnnounceNewReview(t *testing.T) {
	for _, tt := range []struct {
		description    string
		sender         mockEmailSender
		store          mockUserStore
		review         screenjournal.Review
		expectedEmails []email.Message
	}{
		{
			description: "announces new review to everyne except the author",
			store: mockUserStore{
				users: []screenjournal.User{
					{
						Username: screenjournal.Username("alice"),
						Email:    screenjournal.Email("alice.amberson@example.com"),
					},
					{
						Username: screenjournal.Username("bob"),
						Email:    screenjournal.Email("bob.bobberton@example.com"),
					},
					{
						Username: screenjournal.Username("charlie"),
						Email:    screenjournal.Email("charlie.barley@example.com"),
					},
				},
			},
			review: screenjournal.Review{
				ID:    screenjournal.ReviewID(123),
				Owner: screenjournal.Username("bob"),
				Movie: screenjournal.Movie{
					Title: screenjournal.MediaTitle("The Matrix"),
				},
			},
			expectedEmails: []email.Message{
				{
					From: mail.Address{
						Name:    "ScreenJournal",
						Address: "activity@thescreenjournal.com",
					},
					To: []mail.Address{
						{
							Name:    "alice",
							Address: "alice.amberson@example.com",
						},
					},
					Subject: "bob posted a new review: The Matrix",
					TextBody: `Hey alice,

bob just posted a new review for The Matrix! Check it out:

https://dev.thescreenjournal.com/reviews/123

-ScreenJournal Bot
`,
				},
				{
					From: mail.Address{
						Name:    "ScreenJournal",
						Address: "activity@thescreenjournal.com",
					},
					To: []mail.Address{
						{
							Name:    "charlie",
							Address: "charlie.barley@example.com",
						},
					},
					Subject: "bob posted a new review: The Matrix",
					TextBody: `Hey charlie,

bob just posted a new review for The Matrix! Check it out:

https://dev.thescreenjournal.com/reviews/123

-ScreenJournal Bot
`,
				},
			},
		},
		{
			description: "sends no emails when no users exist except the author",
			store: mockUserStore{
				users: []screenjournal.User{
					{
						Username: screenjournal.Username("bob"),
						Email:    screenjournal.Email("bob.bobberton@example.com"),
					},
				},
			},
			review: screenjournal.Review{
				ID:    screenjournal.ReviewID(146),
				Owner: screenjournal.Username("bob"),
				Movie: screenjournal.Movie{
					Title: screenjournal.MediaTitle("Big"),
				},
			},
			expectedEmails: []email.Message{},
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			sender := mockEmailSender{
				emailsSent: []email.Message{},
			}

			announcer := email_announce.New("https://dev.thescreenjournal.com", &sender, tt.store)
			announcer.AnnounceNewReview(tt.review)

			// Clear timestamps for easier comparisons.
			for i := range sender.emailsSent {
				sender.emailsSent[i].Date = time.Time{}
			}

			if got, want := sender.emailsSent, tt.expectedEmails; !reflect.DeepEqual(got, want) {
				t.Fatalf("unexpected announcement emails, got=%+v, want=%+v", got, want)
			}
		})
	}
}
