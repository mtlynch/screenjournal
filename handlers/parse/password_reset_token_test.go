package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestPasswordResetToken(t *testing.T) {
	for _, tt := range []struct {
		description string
		input       string
		token       screenjournal.PasswordResetToken
		err         error
	}{
		{
			"valid token with all uppercase letters is valid",
			"ABCDEFGHJKLMNPQRSTUVWXYZABCDEFGH",
			screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZABCDEFGH"),
			nil,
		},
		{
			"valid token with all lowercase letters is valid",
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			screenjournal.NewPasswordResetTokenFromString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			nil,
		},
		{
			"valid token with all numbers is valid",
			"23456789234567892345678923456789",
			screenjournal.NewPasswordResetTokenFromString("23456789234567892345678923456789"),
			nil,
		},
		{
			"valid token with mixed characters is valid",
			"ABCDEFGHJKLMNPQRSTUVWXYZabcdef23",
			screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef23"),
			nil,
		},
		{
			"valid token with exactly 32 characters is valid",
			strings.Repeat("A", 32),
			screenjournal.NewPasswordResetTokenFromString(strings.Repeat("A", 32)),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.PasswordResetToken{},
			parse.ErrInvalidPasswordResetToken,
		},
		{
			"token with 31 characters is invalid",
			strings.Repeat("A", 31),
			screenjournal.PasswordResetToken{},
			parse.ErrInvalidPasswordResetToken,
		},
		{
			"token with 33 characters is invalid",
			strings.Repeat("A", 33),
			screenjournal.PasswordResetToken{},
			parse.ErrInvalidPasswordResetToken,
		},
		{
			"token with tab character is invalid",
			"ABCDEFGHJKLMNPQRSTUVWXYZABCDEF\tH",
			screenjournal.PasswordResetToken{},
			parse.ErrInvalidPasswordResetToken,
		},
		{
			"single character token is invalid",
			"A",
			screenjournal.PasswordResetToken{},
			parse.ErrInvalidPasswordResetToken,
		},
		{
			"very long token is invalid",
			strings.Repeat("A", 100),
			screenjournal.PasswordResetToken{},
			parse.ErrInvalidPasswordResetToken,
		},
		{
			"token with unicode character is invalid",
			"ABCDEFGHJKLMNPQRSTUVWXYZABCDEF🔑",
			screenjournal.PasswordResetToken{},
			parse.ErrInvalidPasswordResetToken,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.input), func(t *testing.T) {
			token, err := parse.PasswordResetToken(tt.input)

			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := token, tt.token; got != want {
				t.Errorf("token=%v, want=%v", got, want)
			}
		})
	}
}
