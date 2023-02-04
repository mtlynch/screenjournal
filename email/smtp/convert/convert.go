package convert

import (
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/mtlynch/screenjournal/v2/email"
)

type header struct {
	Name  string
	Value string
}

// Boundary to use in generating multipart messages. Really only useful in
// testing.
var MultipartBoundary = ""

func FromEmail(msg email.Message) (string, error) {
	var sb strings.Builder

	mpw := multipart.NewWriter(&sb)

	if MultipartBoundary != "" {
		if err := mpw.SetBoundary(MultipartBoundary); err != nil {
			panic(err)
		}
	}

	headers := []header{}
	headers = append(headers, makeHeader("From", msg.From.String()))
	headers = append(headers, makeHeader("To", msg.To[0].String()))
	headers = append(headers, makeHeader("Subject", msg.Subject))
	headers = append(headers, makeHeader("MIME-Version", "1.0"))
	headers = append(headers, makeHeader("Content-Type", fmt.Sprintf("multipart/alternative; boundary=\"%s\"", mpw.Boundary())))
	for _, hdr := range headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", hdr.Name, hdr.Value))
	}

	part, err := mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/plain; charset=\"iso-8859-1\""}, "Content-Transfer-Encoding": {"quoted-printable"}})
	if err != nil {
		panic(err)
	}
	if _, err := part.Write([]byte(msg.TextBody)); err != nil {
		panic(err)
	}

	part, err = mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/html; charset=\"iso-8859-1\""}, "Content-Transfer-Encoding": {"quoted-printable"}})
	if err != nil {
		panic(err)
	}
	if _, err := part.Write([]byte(msg.HtmlBody)); err != nil {
		panic(err)
	}

	return sb.String(), nil
}

func makeHeader(key, value string) header {
	return header{
		Name:  key,
		Value: value,
	}
}
