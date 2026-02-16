package email_test

import (
	"errors"
	"net/mail"
	"testing"
	"time"

	"github.com/kylelemons/godebug/diff"

	"github.com/mtlynch/screenjournal/v2/email"
	passwordreset_email "github.com/mtlynch/screenjournal/v2/passwordreset/email"
	"github.com/mtlynch/screenjournal/v2/ratelimit"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type mockStore struct {
	entry    screenjournal.PasswordResetEntry
	insertFn func(screenjournal.PasswordResetEntry) error
}

func (s *mockStore) InsertPasswordResetEntry(entry screenjournal.PasswordResetEntry) error {
	s.entry = entry
	if s.insertFn != nil {
		return s.insertFn(entry)
	}
	return nil
}

type mockEmailSender struct {
	emailsSent []email.Message
	sendFn     func(email.Message) error
}

func (s *mockEmailSender) Send(msg email.Message) error {
	s.emailsSent = append(s.emailsSent, msg)
	if s.sendFn != nil {
		return s.sendFn(msg)
	}
	return nil
}

var dummyToken = screenjournal.NewPasswordResetTokenFromString("abc123tokenXYZ")

func newDummyToken() screenjournal.PasswordResetToken {
	return dummyToken
}

func TestRequestReset(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	for _, tt := range []struct {
		description    string
		user           screenjournal.User
		recordedResets []screenjournal.Username
		storeErr       error
		sendErr        error
		errExpected    error
		expectedEmails []email.Message
	}{
		{
			description: "sends password reset email to user",
			user: screenjournal.User{
				Username: screenjournal.Username("alice"),
				Email:    screenjournal.Email("alice@example.com"),
			},
			expectedEmails: []email.Message{
				{
					From: mail.Address{
						Name:    "ScreenJournal",
						Address: "password-resets@thescreenjournal.com",
					},
					To: []mail.Address{
						{
							Name:    "alice",
							Address: "alice@example.com",
						},
					},
					Subject: "Reset your ScreenJournal password",
					TextBody: `Hi alice,

We received a request to reset your password. Click the link below to choose a new password:

https://dev.thescreenjournal.com/account/password-reset?token=abc123tokenXYZ

This link will expire in 7 days.

If you didn't request a password reset, you can safely ignore this email.

-ScreenJournal Bot
`,
					HtmlBody: `<p>Hi alice,</p>

<p>We received a request to reset your password. Click the link below to choose a new password:</p>

<p><a href="https://dev.thescreenjournal.com/account/password-reset?token=abc123tokenXYZ">https://dev.thescreenjournal.com/account/password-reset?token=abc123tokenXYZ</a></p>

<p>This link will expire in 7 days.</p>

<p>If you didn't request a password reset, you can safely ignore this email.</p>

<p>-ScreenJournal Bot</p>`,
				},
			},
		},
		{
			description: "returns error when store insert fails",
			user: screenjournal.User{
				Username: screenjournal.Username("alice"),
				Email:    screenjournal.Email("alice@example.com"),
			},
			storeErr:       errors.New("database error"),
			errExpected:    errors.New("inserting password reset entry: database error"),
			expectedEmails: []email.Message{},
		},
		{
			description: "returns error when email send fails",
			user: screenjournal.User{
				Username: screenjournal.Username("alice"),
				Email:    screenjournal.Email("alice@example.com"),
			},
			sendErr:     errors.New("SMTP error"),
			errExpected: errors.New("sending password reset email for user alice: SMTP error"),
			expectedEmails: []email.Message{
				{
					From: mail.Address{
						Name:    "ScreenJournal",
						Address: "password-resets@thescreenjournal.com",
					},
					To: []mail.Address{
						{
							Name:    "alice",
							Address: "alice@example.com",
						},
					},
					Subject: "Reset your ScreenJournal password",
					TextBody: `Hi alice,

We received a request to reset your password. Click the link below to choose a new password:

https://dev.thescreenjournal.com/account/password-reset?token=abc123tokenXYZ

This link will expire in 7 days.

If you didn't request a password reset, you can safely ignore this email.

-ScreenJournal Bot
`,
					HtmlBody: `<p>Hi alice,</p>

<p>We received a request to reset your password. Click the link below to choose a new password:</p>

<p><a href="https://dev.thescreenjournal.com/account/password-reset?token=abc123tokenXYZ">https://dev.thescreenjournal.com/account/password-reset?token=abc123tokenXYZ</a></p>

<p>This link will expire in 7 days.</p>

<p>If you didn't request a password reset, you can safely ignore this email.</p>

<p>-ScreenJournal Bot</p>`,
				},
			},
		},
		{
			description: "silently skips when rate limited",
			user: screenjournal.User{
				Username: screenjournal.Username("alice"),
				Email:    screenjournal.Email("alice@example.com"),
			},
			recordedResets: []screenjournal.Username{
				screenjournal.Username("alice"),
				screenjournal.Username("alice"),
			},
			expectedEmails: []email.Message{},
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			store := &mockStore{}
			if tt.storeErr != nil {
				store.insertFn = func(screenjournal.PasswordResetEntry) error {
					return tt.storeErr
				}
			}

			sender := &mockEmailSender{
				emailsSent: []email.Message{},
			}
			if tt.sendErr != nil {
				sender.sendFn = func(email.Message) error {
					return tt.sendErr
				}
			}

			now := baseTime
			limiter := ratelimit.NewPasswordResetLimiter(func() time.Time { return now })
			for _, username := range tt.recordedResets {
				limiter.Record(username)
			}

			resetter := passwordreset_email.New("https://dev.thescreenjournal.com", sender, store, limiter, newDummyToken, func() time.Time { return now })

			err := resetter.RequestReset(tt.user)

			if got, want := errToString(err), errToString(tt.errExpected); got != want {
				t.Fatalf("err=%s, want=%s", got, want)
			}

			if len(sender.emailsSent) == len(tt.expectedEmails) {
				for i, emailGot := range sender.emailsSent {
					emailWant := tt.expectedEmails[i]
					if d := diff.Diff(emailWant.TextBody, emailGot.TextBody); d != "" {
						t.Errorf("email #%d (plaintext): %s", i, d)
					}
					if d := diff.Diff(emailWant.HtmlBody, emailGot.HtmlBody); d != "" {
						t.Errorf("email #%d (html): %s", i, d)
					}
				}
			}

			if got, want := len(sender.emailsSent), len(tt.expectedEmails); got != want {
				t.Fatalf("email count=%d, want=%d", got, want)
			}
		})
	}
}

func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
