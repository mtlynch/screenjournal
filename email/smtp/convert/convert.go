package convert

import (
	"fmt"
	"strings"

	"github.com/mtlynch/screenjournal/v2/email"
)

type header struct {
	Name  string
	Value string
}

func FromEmail(msg email.Message) (string, error) {
	var sb strings.Builder
	headers := []header{}
	headers = append(headers, makeHeader("From", msg.From.String()))
	headers = append(headers, makeHeader("To", msg.To[0].String()))
	headers = append(headers, makeHeader("Subject", msg.Subject))
	for _, hdr := range headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", hdr.Name, hdr.Value))
	}
	sb.WriteString("\r\n")
	sb.WriteString(msg.TextBody)
	sb.WriteString("\r\n")
	return sb.String(), nil
}

func makeHeader(key, value string) header {
	return header{
		Name:  key,
		Value: value,
	}
}
