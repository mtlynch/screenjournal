package email_test

import (
	"net/mail"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	email_announce "github.com/mtlynch/screenjournal/v2/announce/email"
	"github.com/mtlynch/screenjournal/v2/email"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type mockNotificationsStore struct {
	subscribers []screenjournal.EmailSubscriber
}

func (ns mockNotificationsStore) ReadReviewSubscribers() ([]screenjournal.EmailSubscriber, error) {
	return ns.subscribers, nil
}

func (ns mockNotificationsStore) ReadCommentSubscribers() ([]screenjournal.EmailSubscriber, error) {
	return ns.subscribers, nil
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
		store          mockNotificationsStore
		review         screenjournal.Review
		expectedEmails []email.Message
	}{
		{
			description: "announces new movie review to everyone except the author",
			store: mockNotificationsStore{
				subscribers: []screenjournal.EmailSubscriber{
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
				ID:    screenjournal.ReviewID(456),
				Owner: screenjournal.Username("bob"),
				Movie: screenjournal.Movie{
					ID:    screenjournal.MovieID(123),
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

bob just posted a new review of *The Matrix*! Check it out:

https://dev.thescreenjournal.com/movies/123#review456

-ScreenJournal Bot

To manage your notifications, visit https://dev.thescreenjournal.com/account/notifications
`,
					HtmlBody: `<p>Hey alice,</p>

<p>bob just posted a new review of <em>The Matrix</em>! Check it out:</p>

<p><a href="https://dev.thescreenjournal.com/movies/123#review456">https://dev.thescreenjournal.com/movies/123#review456</a></p>

<p>-ScreenJournal Bot</p>

<p>To manage your notifications, visit <a href="https://dev.thescreenjournal.com/account/notifications">https://dev.thescreenjournal.com/account/notifications</a></p>`,
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

bob just posted a new review of *The Matrix*! Check it out:

https://dev.thescreenjournal.com/movies/123#review456

-ScreenJournal Bot

To manage your notifications, visit https://dev.thescreenjournal.com/account/notifications
`,
					HtmlBody: `<p>Hey charlie,</p>

<p>bob just posted a new review of <em>The Matrix</em>! Check it out:</p>

<p><a href="https://dev.thescreenjournal.com/movies/123#review456">https://dev.thescreenjournal.com/movies/123#review456</a></p>

<p>-ScreenJournal Bot</p>

<p>To manage your notifications, visit <a href="https://dev.thescreenjournal.com/account/notifications">https://dev.thescreenjournal.com/account/notifications</a></p>`,
				},
			},
		},

		{
			description: "announces new TV show review to other user",
			store: mockNotificationsStore{
				subscribers: []screenjournal.EmailSubscriber{
					{
						Username: screenjournal.Username("alice"),
						Email:    screenjournal.Email("alice.amberson@example.com"),
					},
					{
						Username: screenjournal.Username("bob"),
						Email:    screenjournal.Email("bob.bobberton@example.com"),
					},
				},
			},
			review: screenjournal.Review{
				ID:    screenjournal.ReviewID(456),
				Owner: screenjournal.Username("bob"),
				TvShow: screenjournal.TvShow{
					ID:    screenjournal.TvShowID(789),
					Title: screenjournal.MediaTitle("30 Rock"),
				},
				TvShowSeason: screenjournal.TvShowSeason(3),
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
					Subject: "bob posted a new review: 30 Rock (Season 3)",
					TextBody: `Hey alice,

bob just posted a new review of *30 Rock* (Season 3)! Check it out:

https://dev.thescreenjournal.com/tv-shows/789?season=3#review456

-ScreenJournal Bot

To manage your notifications, visit https://dev.thescreenjournal.com/account/notifications
`,
					HtmlBody: `<p>Hey alice,</p>

<p>bob just posted a new review of <em>30 Rock</em> (Season 3)! Check it out:</p>

<p><a href="https://dev.thescreenjournal.com/tv-shows/789?season=3#review456">https://dev.thescreenjournal.com/tv-shows/789?season=3#review456</a></p>

<p>-ScreenJournal Bot</p>

<p>To manage your notifications, visit <a href="https://dev.thescreenjournal.com/account/notifications">https://dev.thescreenjournal.com/account/notifications</a></p>`,
				},
			},
		},
		{
			description: "sends no emails when no users exist except the author",
			store: mockNotificationsStore{
				subscribers: []screenjournal.EmailSubscriber{
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

			if diff := cmp.Diff(tt.expectedEmails, sender.emailsSent, cmpopts.EquateComparable(time.Time{})); diff != "" {
				t.Errorf("announcement emails mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAnnounceNewComment(t *testing.T) {
	for _, tt := range []struct {
		description    string
		sender         mockEmailSender
		store          mockNotificationsStore
		comment        screenjournal.ReviewComment
		expectedEmails []email.Message
	}{
		{
			description: "announces new movie review comment to everyone except the author",
			store: mockNotificationsStore{
				subscribers: []screenjournal.EmailSubscriber{
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
			comment: screenjournal.ReviewComment{
				ID:    screenjournal.CommentID(707),
				Owner: screenjournal.Username("alice"),
				Review: screenjournal.Review{
					ID:    screenjournal.ReviewID(456),
					Owner: screenjournal.Username("bob"),
					Movie: screenjournal.Movie{
						ID:    screenjournal.MovieID(123),
						Title: screenjournal.MediaTitle("The Matrix"),
					},
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
							Name:    "bob",
							Address: "bob.bobberton@example.com",
						},
					},
					Subject: "alice commented on bob's review of The Matrix",
					TextBody: `Hey bob,

alice just commented on bob's review of *The Matrix*! Check it out:

https://dev.thescreenjournal.com/movies/123#comment707

-ScreenJournal Bot

To manage your notifications, visit https://dev.thescreenjournal.com/account/notifications
`,
					HtmlBody: `<p>Hey bob,</p>

<p>alice just commented on bob's review of <em>The Matrix</em>! Check it out:</p>

<p><a href="https://dev.thescreenjournal.com/movies/123#comment707">https://dev.thescreenjournal.com/movies/123#comment707</a></p>

<p>-ScreenJournal Bot</p>

<p>To manage your notifications, visit <a href="https://dev.thescreenjournal.com/account/notifications">https://dev.thescreenjournal.com/account/notifications</a></p>`,
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
					Subject: "alice commented on bob's review of The Matrix",
					TextBody: `Hey charlie,

alice just commented on bob's review of *The Matrix*! Check it out:

https://dev.thescreenjournal.com/movies/123#comment707

-ScreenJournal Bot

To manage your notifications, visit https://dev.thescreenjournal.com/account/notifications
`,
					HtmlBody: `<p>Hey charlie,</p>

<p>alice just commented on bob's review of <em>The Matrix</em>! Check it out:</p>

<p><a href="https://dev.thescreenjournal.com/movies/123#comment707">https://dev.thescreenjournal.com/movies/123#comment707</a></p>

<p>-ScreenJournal Bot</p>

<p>To manage your notifications, visit <a href="https://dev.thescreenjournal.com/account/notifications">https://dev.thescreenjournal.com/account/notifications</a></p>`,
				},
			},
		},
		{
			description: "announces new TV show review comment to other user",
			store: mockNotificationsStore{
				subscribers: []screenjournal.EmailSubscriber{
					{
						Username: screenjournal.Username("alice"),
						Email:    screenjournal.Email("alice.amberson@example.com"),
					},
					{
						Username: screenjournal.Username("bob"),
						Email:    screenjournal.Email("bob.bobberton@example.com"),
					},
				},
			},
			comment: screenjournal.ReviewComment{
				ID:    screenjournal.CommentID(707),
				Owner: screenjournal.Username("alice"),
				Review: screenjournal.Review{
					ID:    screenjournal.ReviewID(456),
					Owner: screenjournal.Username("bob"),
					TvShow: screenjournal.TvShow{
						ID:    screenjournal.TvShowID(789),
						Title: screenjournal.MediaTitle("30 Rock"),
					},
					TvShowSeason: screenjournal.TvShowSeason(3),
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
							Name:    "bob",
							Address: "bob.bobberton@example.com",
						},
					},
					Subject: "alice commented on bob's review of 30 Rock (Season 3)",
					TextBody: `Hey bob,

alice just commented on bob's review of *30 Rock* (Season 3)! Check it out:

https://dev.thescreenjournal.com/tv-shows/789?season=3#comment707

-ScreenJournal Bot

To manage your notifications, visit https://dev.thescreenjournal.com/account/notifications
`,
					HtmlBody: `<p>Hey bob,</p>

<p>alice just commented on bob's review of <em>30 Rock</em> (Season 3)! Check it out:</p>

<p><a href="https://dev.thescreenjournal.com/tv-shows/789?season=3#comment707">https://dev.thescreenjournal.com/tv-shows/789?season=3#comment707</a></p>

<p>-ScreenJournal Bot</p>

<p>To manage your notifications, visit <a href="https://dev.thescreenjournal.com/account/notifications">https://dev.thescreenjournal.com/account/notifications</a></p>`,
				},
			},
		},
		{
			description: "sends no emails when no users exist except the author",
			store: mockNotificationsStore{
				subscribers: []screenjournal.EmailSubscriber{
					{
						Username: screenjournal.Username("bob"),
						Email:    screenjournal.Email("bob.bobberton@example.com"),
					},
				},
			},
			comment: screenjournal.ReviewComment{
				ID:    screenjournal.CommentID(641),
				Owner: screenjournal.Username("bob"),
				Review: screenjournal.Review{
					ID:    screenjournal.ReviewID(146),
					Owner: screenjournal.Username("alice"),
					Movie: screenjournal.Movie{
						Title: screenjournal.MediaTitle("Big"),
					},
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
			announcer.AnnounceNewComment(tt.comment)

			if diff := cmp.Diff(tt.expectedEmails, sender.emailsSent, cmpopts.EquateComparable(time.Time{})); diff != "" {
				t.Errorf("comment announcement emails mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
