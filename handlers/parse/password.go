package parse

import (
	"errors"
	"fmt"
	"sync"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type allowedCharactersLookup struct {
	chars map[rune]bool
	once  sync.Once
}

var (
	ErrPasswordTooShort             = fmt.Errorf("invalid password: must be at least %d characters", minPasswordLength)
	ErrPasswordTooLong              = fmt.Errorf("invalid password: must be %d characters or fewer", maxPasswordLength)
	ErrPasswordHasInvalidCharacters = errors.New("invalid password: must only contain letters A-Z, a-z, 0-9, or special characters")

	allowedPasswordCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`~!@#$%^&*()-_=+[]{}\\|;:'\",<.>/? "
	minPasswordLength         = 8
	maxPasswordLength         = 40

	allowedCharacters allowedCharactersLookup
)

func Password(plaintext string) (screenjournal.Password, error) {
	// Initialize the allowed characters lookup.
	allowedCharacters.once.Do(func() {
		allowedCharacters.chars = map[rune]bool{}
		for _, c := range allowedPasswordCharacters {
			allowedCharacters.chars[c] = true
		}
	})

	if len(plaintext) < minPasswordLength {
		return screenjournal.Password(""), ErrPasswordTooShort
	}
	if len(plaintext) > maxPasswordLength {
		return screenjournal.Password(""), ErrPasswordTooLong
	}
	for _, c := range plaintext {
		if !allowedCharacters.chars[c] {
			return screenjournal.Password(""), ErrPasswordHasInvalidCharacters
		}
	}
	return screenjournal.Password(plaintext), nil
}
