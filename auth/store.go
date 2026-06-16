package auth

import (
	simple_auth "codeberg.org/mtlynch/simpleauth/v3/auth"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

type authStore struct {
	userStore UserStore
}

func (s authStore) ReadPasswordHash(usernameRaw string) (simple_auth.PasswordHash, error) {
	username, err := parse.Username(usernameRaw)
	if err != nil {
		return nil, err
	}

	user, err := s.userStore.ReadUser(username)
	if err != nil {
		return nil, err
	}

	return simple_auth.PasswordHashFromBytes(user.PasswordHash.Bytes()), nil
}
