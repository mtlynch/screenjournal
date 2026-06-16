package smtp

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/smtp"
	"strconv"
	"time"

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
	if host == "" {
		return sender{}, errors.New("invalid SMTP hostname")
	}
	if port == 0 {
		return sender{}, errors.New("invalid SMTP port")
	}
	if username == "" || password == "" {
		return sender{}, errors.New("invalid SMTP credentials")
	}
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

	serverName := net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port))
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.Dial("tcp", serverName)
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer func() {
		if err := c.Quit(); err != nil {
			log.Printf("failed to close SMTP connection: %v", err)
		}
	}()

	if err := c.StartTLS(&tls.Config{
		ServerName: s.config.Host,
	}); err != nil {
		return err
	}

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
	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
