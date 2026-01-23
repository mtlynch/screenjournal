package parse

import (
	"errors"
	"fmt"
	"sync"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	ErrPasswordTooShort             = fmt.Errorf("invalid password: must be at least %d characters", minPasswordLength)
	ErrPasswordTooLong              = fmt.Errorf("invalid password: must be %d characters or fewer", maxPasswordLength)
	ErrPasswordHasInvalidCharacters = errors.New("invalid password: must only contain letters A-Z, a-z, 0-9, or special characters")

	minPasswordLength = 8
	maxPasswordLength = 40

	allowedCharacters = sync.OnceValue(func() map[rune]bool {
		const allowedPasswordCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`~!@#$%^&*()-_=+[]{}\\|;:'\",<.>/? "
		chars := make(map[rune]bool, len(allowedPasswordCharacters))
		for _, c := range allowedPasswordCharacters {
			chars[c] = true
		}
		return chars
	})
)

func Password(plaintext string) (screenjournal.Password, error) {
	if len(plaintext) < minPasswordLength {
		return screenjournal.Password(""), ErrPasswordTooShort
	}
	if len(plaintext) > maxPasswordLength {
		return screenjournal.Password(""), ErrPasswordTooLong
	}

	allowed := allowedCharacters()
	for _, c := range plaintext {
		if !allowed[c] {
			return screenjournal.Password(""), ErrPasswordHasInvalidCharacters
		}
	}
	return screenjournal.Password(plaintext), nil
}
