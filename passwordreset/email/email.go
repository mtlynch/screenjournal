package email

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"net/mail"
	"text/template"
	"time"

	"github.com/mtlynch/screenjournal/v2/email"
	"github.com/mtlynch/screenjournal/v2/markdown"
	"github.com/mtlynch/screenjournal/v2/ratelimit"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	Store interface {
		InsertPasswordResetEntry(screenjournal.PasswordResetEntry) error
	}

	Resetter struct {
		baseURL  string
		sender   email.Sender
		store    Store
		limiter  *ratelimit.PasswordResetLimiter
		newToken func() screenjournal.PasswordResetToken
		now      func() time.Time
	}
)

func New(baseURL string, sender email.Sender, store Store, limiter *ratelimit.PasswordResetLimiter, newToken func() screenjournal.PasswordResetToken, now func() time.Time) Resetter {
	return Resetter{
		baseURL:  baseURL,
		sender:   sender,
		store:    store,
		limiter:  limiter,
		newToken: newToken,
		now:      now,
	}
}

//go:embed templates
var templatesFS embed.FS

var emailTemplate = template.Must(
	template.New("password-reset.tmpl.txt").
		ParseFS(templatesFS, "templates/password-reset.tmpl.txt"))

func (r Resetter) Request(user screenjournal.User) error {
	if !r.limiter.Allow(user.Username) {
		log.Printf("password reset rate limited for user %s", user.Username)
		return nil
	}

	entry := screenjournal.PasswordResetEntry{
		Username:  user.Username,
		Token:     r.newToken(),
		ExpiresAt: r.now().Add(7 * 24 * time.Hour),
	}

	if err := r.store.InsertPasswordResetEntry(entry); err != nil {
		return fmt.Errorf("inserting password reset entry: %w", err)
	}

	resetURL := fmt.Sprintf("%s/account/password-reset?token=%s", r.baseURL, entry.Token)

	var bodyBuf bytes.Buffer
	if err := emailTemplate.Execute(&bodyBuf, struct {
		Username string
		ResetURL string
	}{
		Username: user.Username.String(),
		ResetURL: resetURL,
	}); err != nil {
		return fmt.Errorf("rendering password reset email: %w", err)
	}

	bodyMarkdown := screenjournal.EmailBodyMarkdown(bodyBuf.String())
	msg := email.Message{
		From: mail.Address{
			Name:    "ScreenJournal",
			Address: "password-resets@thescreenjournal.com",
		},
		To: []mail.Address{
			{
				Name:    user.Username.String(),
				Address: user.Email.String(),
			},
		},
		Subject:  "Reset your ScreenJournal password",
		TextBody: bodyMarkdown.String(),
		HtmlBody: markdown.RenderEmail(bodyMarkdown),
	}

	if err := r.sender.Send(msg); err != nil {
		return fmt.Errorf("sending password reset email for user %s: %w", user.Username, err)
	}

	r.limiter.Record(user.Username)

	log.Printf("sent password reset email for user %s", user.Username)
	return nil
}
