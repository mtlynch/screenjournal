package parse

import (
	"errors"
	"regexp"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	ErrInvalidPasswordResetToken = errors.New("invalid password reset token")

	passwordResetTokenPattern = regexp.MustCompile(`^[ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789]{32}$`)
)

func PasswordResetToken(token string) (screenjournal.PasswordResetToken, error) {
	if token == "" {
		return screenjournal.PasswordResetToken{}, ErrInvalidPasswordResetToken
	}

	if !passwordResetTokenPattern.MatchString(token) {
		return screenjournal.PasswordResetToken{}, ErrInvalidPasswordResetToken
	}

	return screenjournal.NewPasswordResetTokenFromString(token), nil
}
