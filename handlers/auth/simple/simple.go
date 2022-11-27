package simple

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/auth"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"golang.org/x/crypto/pbkdf2"
)

const authCookieName = "sharedSecret"

type (
	sharedSecret []byte

	SimpleAuthenticator struct {
		username     screenjournal.Username
		sharedSecret sharedSecret
	}
)

func New(username, password string) (SimpleAuthenticator, error) {
	u, err := parse.Username(username)
	if err != nil {
		return SimpleAuthenticator{}, err
	}

	pw, err := parse.Password(password)
	if err != nil {
		return SimpleAuthenticator{}, err
	}

	// Simply concatenating the username and password together for a shared secret
	// is not very secure, but this is just a temporary, placeholder
	// implementation.
	ss, err := sharedSecretFromBytes([]byte(u.String() + pw.String()))
	if err != nil {
		return SimpleAuthenticator{}, err
	}

	return SimpleAuthenticator{
		username:     u,
		sharedSecret: ss,
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

func (sa SimpleAuthenticator) createCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    base64.StdEncoding.EncodeToString(sa.sharedSecret),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
	})
}

func (sa SimpleAuthenticator) ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
	})
}

func sharedSecretFromBytes(b []byte) (sharedSecret, error) {
	if len(b) == 0 {
		return sharedSecret{}, errors.New("invalid shared secret key")
	}

	// These would be insecure values for storing a database of user credentials,
	// but we're only storing a single password, so it's not important to have
	// random salt or high iteration rounds.
	staticSalt := []byte{1, 2, 3, 4}
	iter := 100

	dk := pbkdf2.Key(b, staticSalt, iter, 32, sha256.New)

	return sharedSecret(dk), nil
}

func sharedSecretFromBase64(b64encoded string) (sharedSecret, error) {
	if len(b64encoded) == 0 {
		return sharedSecret{}, errors.New("invalid shared secret")
	}

	decoded, err := base64.StdEncoding.DecodeString(b64encoded)
	if err != nil {
		return sharedSecret{}, err
	}

	return sharedSecret(decoded), nil
}

func sharedSecretsEqual(a, b sharedSecret) bool {
	return subtle.ConstantTimeCompare(a, b) != 0
}
