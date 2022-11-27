package simple

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/sessions"
	"golang.org/x/crypto/pbkdf2"
)

const authCookieName = "sharedSecret"

type (
	sharedSecret []byte

	manager struct {
		username     screenjournal.Username
		sharedSecret sharedSecret
	}
)

func New(username screenjournal.Username, password string) (sessions.Manager, error) {
	// Simply concatenating the username and password together for a shared secret
	// is not very secure, but this is just a temporary, placeholder
	// implementation.
	ss, err := sharedSecretFromBytes([]byte(username.String() + password))
	if err != nil {
		return manager{}, err
	}

	return manager{
		username:     username,
		sharedSecret: ss,
	}, nil
}

func (m manager) Create(w http.ResponseWriter, r *http.Request, _ screenjournal.Username) error {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    base64.StdEncoding.EncodeToString(m.sharedSecret),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
	})
	return nil
}

func (m manager) End(ctx context.Context, w http.ResponseWriter) {
	// The simple manager can't really invalidate sessions because the credentials
	// are hard-coded and the session token is static, so all we can do is ask the
	// client to delete their cookie.
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
