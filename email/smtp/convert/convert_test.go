package convert_test

import (
	"net/mail"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/diff"

	"github.com/mtlynch/screenjournal/v2/email"
	"github.com/mtlynch/screenjournal/v2/email/smtp/convert"
)

func TestFromEmail(t *testing.T) {
	var tests = []struct {
		input    email.Message
		expected string
	}{
		{
			input: email.Message{
				From: mail.Address{
					Name:    "ScreenJournal Bot",
					Address: "bot@sj.example.com",
				},
				To: []mail.Address{
					{
						Name:    "Alice User",
						Address: "alice@user.example.com",
					},
				},
				Subject: "Frank posted a review of The Room",
				TextBody: `Hi Alice,

Frank has posted a new review of *The Room*:

https://sj.example.com/reviews/25

Sincerely,
ScreenJournal Bot`,
				HtmlBody: `<p>Hi Alice,</p>

<p>Frank has posted a new review of <em>The Room</em>:</p>

<p><a href="https://sj.example.com/reviews/25">https://sj.example.com/reviews/25</a></p>

<p>-ScreenJournal Bot</p>`,
			},
			expected: normalizeLineEndings(`From: "ScreenJournal Bot" <bot@sj.example.com>
To: "Alice User" <alice@user.example.com>
Subject: Frank posted a review of The Room
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="boundary-type-1234567892-alt"
--boundary-type-1234567892-alt
Content-Transfer-Encoding: quoted-printable
Content-Type: text/plain; charset="iso-8859-1"

Hi Alice,

Frank has posted a new review of *The Room*:

https://sj.example.com/reviews/25

Sincerely,
ScreenJournal Bot
--boundary-type-1234567892-alt
Content-Transfer-Encoding: quoted-printable
Content-Type: text/html; charset="iso-8859-1"

<p>Hi Alice,</p>

<p>Frank has posted a new review of <em>The Room</em>:</p>

<p><a href="https://sj.example.com/reviews/25">https://sj.example.com/reviews/25</a></p>

<p>-ScreenJournal Bot</p>`),
		},
	}

	for _, tt := range tests {
		actual, err := convert.FromEmail(tt.input)
		if err != nil {
			t.Fatalf("failed to generate email: %v", err)
		}

		if diff := diff.Diff(actual, tt.expected); diff != "" {
			t.Fatalf("unexpected smtp message for email: %s\n%s", tt.input.Subject, diff)
		}
	}
}

func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\n", "\r\n")
}
