package auth

import (
	simple_auth "github.com/mtlynch/simpleauth/v2/auth"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func HashPassword(password screenjournal.Password) (screenjournal.PasswordHash, error) {
	h, err := simple_auth.HashPassword(password.String())
	if err != nil {
		return screenjournal.PasswordHash{}, err
	}
	return screenjournal.PasswordHash(h.Bytes()), nil
}
