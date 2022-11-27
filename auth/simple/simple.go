package simple

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/auth"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

type (
	SimpleAuthenticator struct {
		username screenjournal.Username
		password string
	}
)

func New(username, password string) (SimpleAuthenticator, error) {
	u, err := parse.Username(username)
	if err != nil {
		return SimpleAuthenticator{}, err
	}

	return SimpleAuthenticator{
		username: u,
		password: password,
	}, nil
}

func (sa SimpleAuthenticator) StartSession(w http.ResponseWriter, r *http.Request) {
	ss, err := sharedSecretFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid shared secret", http.StatusBadRequest)
		return
	}

	if !sharedSecretsEqual(ss, sa.sharedSecret) {
		http.Error(w, "Incorrect shared secret", http.StatusUnauthorized)
		return
	}

	sa.createCookie(w)
}

func sharedSecretFromRequest(r *http.Request) (sharedSecret, error) {
	body := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		return sharedSecret{}, err
	}

	return sharedSecretFromBytes([]byte(body.Username + body.Password))
}

func (sa SimpleAuthenticator) Authenticate(r *http.Request) (screenjournal.UserAuth, error) {
	authCookie, err := r.Cookie(authCookieName)
	if err != nil {
		return screenjournal.UserAuth{}, auth.ErrNotAuthenticated
	}

	ss, err := sharedSecretFromBase64(authCookie.Value)
	if err != nil {
		return screenjournal.UserAuth{}, errors.New("invalid shared secret")
	}

	if !sharedSecretsEqual(ss, sa.sharedSecret) {
		return screenjournal.UserAuth{}, errors.New("invalid shared secret")
	}

	return screenjournal.UserAuth{
		// SimpleAuthenticator only has a single, admin user.
		IsAdmin:  true,
		Username: sa.username,
	}, nil
}
