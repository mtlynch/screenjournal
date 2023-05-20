package simple_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/auth/simple"
)

type (
	mockPasswordHash struct {
		data []byte
	}

	mockAuthEntry struct {
		Username     string
		PasswordHash auth.PasswordHash
	}

	mockAuthStore struct {
		entries []mockAuthEntry
	}
)

func newMockPasswordHash(password string) auth.PasswordHash {
	return mockPasswordHash{
		// We're not really hashing the password, but it's okay because this is just
		// mock data for testing.
		data: []byte(password),
	}
}

func (h mockPasswordHash) MatchesPlaintext(plaintext string) bool {
	other := newMockPasswordHash(plaintext)

	return bytes.Equal(h.Bytes(), other.Bytes())
}

func (h mockPasswordHash) String() string {
	return string(h.data)
}

func (h mockPasswordHash) Bytes() []byte {
	return h.data
}

func (s mockAuthStore) InsertUser(username, password string) error {
	return errors.New("not implemented")
}

func (s mockAuthStore) ReadPasswordHash(username string) (auth.PasswordHash, error) {
	for _, entry := range s.entries {
		if entry.Username == username {
			return entry.PasswordHash, nil
		}
	}
	return nil, screenjournal.ErrUserNotFound
}

func TestAuthenticate(t *testing.T) {
	for _, tt := range []struct {
		description string
		store       simple.AuthStore
		username    string
		password    string
		err         error
	}{
		{
			"authenticates when password is valid",
			mockAuthStore{
				entries: []mockAuthEntry{
					{
						Username:     "dummyuser",
						PasswordHash: newMockPasswordHash("dummy-p@ssword"),
					},
				},
			},
			"dummyuser",
			"dummy-p@ssword",
			nil,
		},
		{
			"returns ErrIncorrectPassword when password is invalid",
			mockAuthStore{
				entries: []mockAuthEntry{
					{
						Username:     "dummyuser",
						PasswordHash: newMockPasswordHash("dummy-p@ssword"),
					},
				},
			},
			"dummyuser",
			"wrongpass",
			auth.ErrIncorrectPassword,
		},
		{
			"returns ErrUserNotFound when user is not found",
			mockAuthStore{
				entries: []mockAuthEntry{
					{
						Username:     "dummyuser",
						PasswordHash: newMockPasswordHash("dummy-p@ssword"),
					},
				},
			},
			"madeupuser",
			"dummy-p@ssword",
			screenjournal.ErrUserNotFound,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			authenticator := simple.New(tt.store)
			err := authenticator.Authenticate(tt.username, tt.password)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
		})
	}
}
