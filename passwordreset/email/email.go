package email

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"net/mail"
	"text/template"

	"github.com/mtlynch/screenjournal/v2/email"
	"github.com/mtlynch/screenjournal/v2/markdown"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type Resetter struct {
	baseURL string
	sender  email.Sender
}

func New(baseURL string, sender email.Sender) Resetter {
	return Resetter{
		baseURL: baseURL,
		sender:  sender,
	}
}

//go:embed templates
var templatesFS embed.FS

var emailTemplate = template.Must(
	template.New("password-reset.tmpl.txt").
		ParseFS(templatesFS, "templates/password-reset.tmpl.txt"))

func (r Resetter) Send(user screenjournal.User, entry screenjournal.PasswordResetEntry) error {
	resetURL := fmt.Sprintf("%s/account/password-reset?username=%s&token=%s", r.baseURL, user.Username, entry.Token)

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

	log.Printf("sent password reset email for user %s", user.Username)
	return nil
}
