package email

import (
	"net/mail"
	"time"
)

type Message struct {
	From     mail.Address
	To       []mail.Address
	Bcc      []mail.Address
	Subject  string
	Date     time.Time
	TextBody string
	HtmlBody string
}

type Sender interface {
	Send(message Message) error
}
