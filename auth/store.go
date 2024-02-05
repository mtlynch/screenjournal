package auth

import (
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	simple_auth "github.com/mtlynch/simpleauth/v2/auth"
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
