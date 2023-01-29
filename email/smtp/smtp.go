package smtp

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"

	"github.com/mtlynch/screenjournal/v2/email"
	"github.com/mtlynch/screenjournal/v2/email/smtp/convert"
)

type (
	config struct {
		Host     string
		Port     int
		Username string
		Password string
	}

	sender struct {
		config config
	}
)

func New(host string, port int, username, password string) (email.Sender, error) {
	// TODO: Check for valid config.
	return sender{
		config: config{
			Host:     host,
			Port:     port,
			Username: username,
			Password: password,
		},
	}, nil
}

func (s sender) Send(msg email.Message) error {
	log.Printf("sending email from %s to %s (%s)", msg.From.String(), msg.To[0].String(), msg.Subject)

	serverName := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	c, err := smtp.Dial(serverName)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
	}
	err = c.StartTLS(tlsConfig)
	if err != nil {
		return err
	}

	defer func() {
		if err := c.Quit(); err != nil {
			log.Printf("failed to close TLS connection: %v", err)
		}
	}()

	// Plain auth is okay since we're wrapping it in TLS.
	if err := c.Auth(smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)); err != nil {
		return err
	}

	if err := c.Mail(msg.From.Address); err != nil {
		return err
	}

	rcpts := msg.To
	// TODO: Add cc and bcc recepients
	for _, rcpt := range rcpts {
		if err := c.Rcpt(rcpt.Address); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	rawMsg, err := convert.FromEmail(msg)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(rawMsg))
	if err != nil {
		return err
	}

	return nil
}
