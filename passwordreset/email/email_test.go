package email_test

import (
	"errors"
	"fmt"
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

func TestRequestReset(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	for _, tt := range []struct {
		description        string
		user               screenjournal.User
		recordedResets     []screenjournal.Username
		storeErr           error
		sendErr            error
		errExpected        error
		emailCountExpected int
	}{
		{
			description: "sends password reset email to user",
			user: screenjournal.User{
				Username: screenjournal.Username("alice"),
				Email:    screenjournal.Email("alice@example.com"),
			},
			emailCountExpected: 1,
		},
		{
			description: "returns error when store insert fails",
			user: screenjournal.User{
				Username: screenjournal.Username("alice"),
				Email:    screenjournal.Email("alice@example.com"),
			},
			storeErr:           errors.New("database error"),
			errExpected:        errors.New("inserting password reset entry: database error"),
			emailCountExpected: 0,
		},
		{
			description: "returns error when email send fails",
			user: screenjournal.User{
				Username: screenjournal.Username("alice"),
				Email:    screenjournal.Email("alice@example.com"),
			},
			sendErr:            errors.New("SMTP error"),
			errExpected:        errors.New("sending password reset email for user alice: SMTP error"),
			emailCountExpected: 1,
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
			emailCountExpected: 0,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			store := &mockStore{}
			if tt.storeErr != nil {
				store.insertFn = func(screenjournal.PasswordResetEntry) error {
					return tt.storeErr
				}
			}

			sender := &mockEmailSender{}
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

			resetter := passwordreset_email.New("https://dev.thescreenjournal.com", sender, store, limiter)

			err := resetter.RequestReset(tt.user)

			if got, want := errToString(err), errToString(tt.errExpected); got != want {
				t.Fatalf("err=%s, want=%s", got, want)
			}
			if got, want := len(sender.emailsSent), tt.emailCountExpected; got != want {
				t.Fatalf("email count=%d, want=%d", got, want)
			}

			if tt.emailCountExpected == 0 {
				return
			}

			msg := sender.emailsSent[0]

			if got, want := msg.From, (mail.Address{Name: "ScreenJournal", Address: "password-resets@thescreenjournal.com"}); got != want {
				t.Errorf("from=%v, want=%v", got, want)
			}
			if got, want := len(msg.To), 1; got != want {
				t.Fatalf("to count=%d, want=%d", got, want)
			}
			if got, want := msg.To[0], (mail.Address{Name: tt.user.Username.String(), Address: tt.user.Email.String()}); got != want {
				t.Errorf("to=%v, want=%v", got, want)
			}
			if got, want := msg.Subject, "Reset your ScreenJournal password"; got != want {
				t.Errorf("subject=%s, want=%s", got, want)
			}

			// Build expected email body using the token captured by the
			// mock store.
			resetURL := fmt.Sprintf("https://dev.thescreenjournal.com/account/password-reset?token=%s", store.entry.Token)
			wantTextBody := fmt.Sprintf(`Hi %s,

We received a request to reset your password. Click the link below to choose a new password:

%s

This link will expire in 7 days.

If you didn't request a password reset, you can safely ignore this email.

-ScreenJournal Bot
`, tt.user.Username, resetURL)

			if d := diff.Diff(wantTextBody, msg.TextBody); d != "" {
				t.Errorf("text body diff:\n%s", d)
			}

			wantHtmlBody := fmt.Sprintf(`<p>Hi %s,</p>

<p>We received a request to reset your password. Click the link below to choose a new password:</p>

<p><a href="%s">%s</a></p>

<p>This link will expire in 7 days.</p>

<p>If you didn't request a password reset, you can safely ignore this email.</p>

<p>-ScreenJournal Bot</p>`, tt.user.Username, resetURL, resetURL)

			if d := diff.Diff(wantHtmlBody, msg.HtmlBody); d != "" {
				t.Errorf("html body diff:\n%s", d)
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
