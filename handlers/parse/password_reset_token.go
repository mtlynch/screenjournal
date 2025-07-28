package parse

import (
	"errors"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	ErrInvalidPasswordResetToken = errors.New("invalid password reset token")
)

func isValidPasswordResetToken(token string) bool {
	if len(token) != screenjournal.PasswordResetTokenLength {
		return false
	}

	charsetMap := make(map[rune]bool)
	for _, r := range screenjournal.PasswordResetTokenCharset {
		charsetMap[r] = true
	}

	for _, r := range token {
		if !charsetMap[r] {
			return false
		}
	}
	return true
}

func PasswordResetToken(token string) (screenjournal.PasswordResetToken, error) {
	if token == "" {
		return screenjournal.PasswordResetToken{}, ErrInvalidPasswordResetToken
	}

	if !isValidPasswordResetToken(token) {
		return screenjournal.PasswordResetToken{}, ErrInvalidPasswordResetToken
	}

	return screenjournal.NewPasswordResetTokenFromString(token), nil
}
