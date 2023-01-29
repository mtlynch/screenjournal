package convert_test

import (
	"net/mail"
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
			},
			expected: "From: \"ScreenJournal Bot\" <bot@sj.example.com>\r\n" +
				"To: \"Alice User\" <alice@user.example.com>\r\n" +
				"Subject: Frank posted a review of The Room\r\n" +
				"\r\n" +
				`Hi Alice,

Frank has posted a new review of *The Room*:

https://sj.example.com/reviews/25

Sincerely,
ScreenJournal Bot` +
				"\r\n",
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
