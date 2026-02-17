package passwordreset_test

import (
	"bytes"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2/passwordreset"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

type mockStore struct {
	usersByEmail    map[screenjournal.Email]screenjournal.User
	usersByUsername map[screenjournal.Username]screenjournal.User
	entriesByToken  map[string]screenjournal.PasswordResetEntry

	readUserByEmailErr error
	readEntryErr       error
	insertEntryErr     error
	updatePasswordErr  error
	deleteEntryErr     error

	updatedPasswords map[screenjournal.Username]screenjournal.PasswordHash
}

func newMockStore() *mockStore {
	return &mockStore{
		usersByEmail:     map[screenjournal.Email]screenjournal.User{},
		usersByUsername:  map[screenjournal.Username]screenjournal.User{},
		entriesByToken:   map[string]screenjournal.PasswordResetEntry{},
		updatedPasswords: map[screenjournal.Username]screenjournal.PasswordHash{},
	}
}

func (s *mockStore) ReadUserByEmail(email screenjournal.Email) (screenjournal.User, error) {
	if s.readUserByEmailErr != nil {
		return screenjournal.User{}, s.readUserByEmailErr
	}
	if user, ok := s.usersByEmail[email]; ok {
		return user, nil
	}
	return screenjournal.User{}, store.ErrUserNotFound
}

func (s *mockStore) ReadPasswordResetEntry(token screenjournal.PasswordResetToken) (screenjournal.PasswordResetEntry, error) {
	if s.readEntryErr != nil {
		return screenjournal.PasswordResetEntry{}, s.readEntryErr
	}
	if entry, ok := s.entriesByToken[token.String()]; ok {
		return entry, nil
	}
	return screenjournal.PasswordResetEntry{}, sql.ErrNoRows
}

func (s *mockStore) InsertPasswordResetEntry(entry screenjournal.PasswordResetEntry) error {
	if s.insertEntryErr != nil {
		return s.insertEntryErr
	}
	s.entriesByToken[entry.Token.String()] = entry
	return nil
}

func (s *mockStore) UpdateUserPassword(username screenjournal.Username, newPasswordHash screenjournal.PasswordHash) error {
	if s.updatePasswordErr != nil {
		return s.updatePasswordErr
	}
	s.updatedPasswords[username] = newPasswordHash
	return nil
}

func (s *mockStore) DeletePasswordResetEntry(token screenjournal.PasswordResetToken) error {
	if s.deleteEntryErr != nil {
		return s.deleteEntryErr
	}
	delete(s.entriesByToken, token.String())
	return nil
}

type mockEmailSender struct {
	sentMessages []emailMessage
	err          error
}

type emailMessage struct {
	user  screenjournal.User
	entry screenjournal.PasswordResetEntry
}

func (s *mockEmailSender) Send(user screenjournal.User, entry screenjournal.PasswordResetEntry) error {
	s.sentMessages = append(s.sentMessages, emailMessage{
		user:  user,
		entry: entry,
	})
	if s.err != nil {
		return s.err
	}
	return nil
}

func TestSendEmail(t *testing.T) {
	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("sends email and stores reset token for known user", func(t *testing.T) {
		dataStore := newMockStore()
		sender := &mockEmailSender{}
		user := screenjournal.User{
			Username: screenjournal.Username("alice"),
			Email:    screenjournal.Email("alice@example.com"),
		}
		dataStore.usersByEmail[user.Email] = user

		resetter := passwordreset.New(dataStore, sender, func() time.Time { return fixedNow })

		err := resetter.SendEmail(user.Email)
		if got, want := err, error(nil); got != want {
			t.Fatalf("err=%v, want=%v", got, want)
		}

		if got, want := len(sender.sentMessages), 1; got != want {
			t.Fatalf("emailsSent=%d, want=%d", got, want)
		}

		msg := sender.sentMessages[0]
		if got, want := msg.user.Username, user.Username; !got.Equal(want) {
			t.Errorf("username=%s, want=%s", got, want)
		}
		if got, want := msg.entry.ExpiresAt, fixedNow.Add(7*24*time.Hour); !got.Equal(want) {
			t.Errorf("expiresAt=%s, want=%s", got, want)
		}
		if got, want := msg.entry.Token.Empty(), false; got != want {
			t.Errorf("tokenEmpty=%v, want=%v", got, want)
		}
	})

	t.Run("returns success for unknown user email", func(t *testing.T) {
		dataStore := newMockStore()
		sender := &mockEmailSender{}
		resetter := passwordreset.New(dataStore, sender, func() time.Time { return fixedNow })

		err := resetter.SendEmail(screenjournal.Email("nobody@example.com"))
		if got, want := err, error(nil); got != want {
			t.Fatalf("err=%v, want=%v", got, want)
		}
		if got, want := len(sender.sentMessages), 0; got != want {
			t.Fatalf("emailsSent=%d, want=%d", got, want)
		}
	})

	t.Run("rate limits after two sends per user", func(t *testing.T) {
		dataStore := newMockStore()
		sender := &mockEmailSender{}
		user := screenjournal.User{
			Username: screenjournal.Username("alice"),
			Email:    screenjournal.Email("alice@example.com"),
		}
		dataStore.usersByEmail[user.Email] = user

		resetter := passwordreset.New(dataStore, sender, func() time.Time { return fixedNow })

		for range 3 {
			err := resetter.SendEmail(user.Email)
			if got, want := err, error(nil); got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
		}

		if got, want := len(sender.sentMessages), 2; got != want {
			t.Fatalf("emailsSent=%d, want=%d", got, want)
		}
	})
}

func TestReset(t *testing.T) {
	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("resets password and deletes token on success", func(t *testing.T) {
		dataStore := newMockStore()
		sender := &mockEmailSender{}
		username := screenjournal.Username("alice")
		token := screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef23")
		newPasswordHash := screenjournal.PasswordHash("new-password-hash")

		dataStore.entriesByToken[token.String()] = screenjournal.PasswordResetEntry{
			Username:  username,
			Token:     token,
			ExpiresAt: fixedNow.Add(time.Hour),
		}

		resetter := passwordreset.New(dataStore, sender, func() time.Time { return fixedNow })

		err := resetter.Reset(username, token, newPasswordHash)
		if got, want := err, error(nil); got != want {
			t.Fatalf("err=%v, want=%v", got, want)
		}
		if got, want := dataStore.updatedPasswords[username], newPasswordHash; !bytes.Equal(got, want) {
			t.Fatalf("passwordHash=%v, want=%v", got, want)
		}
		if got, want := len(dataStore.entriesByToken), 0; got != want {
			t.Fatalf("remainingTokens=%d, want=%d", got, want)
		}
	})

	t.Run("returns invalid token error for missing token", func(t *testing.T) {
		dataStore := newMockStore()
		sender := &mockEmailSender{}
		username := screenjournal.Username("alice")
		token := screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef23")

		resetter := passwordreset.New(dataStore, sender, func() time.Time { return fixedNow })

		err := resetter.Reset(username, token, screenjournal.PasswordHash("new-password-hash"))
		if got, want := errors.Is(err, passwordreset.ErrInvalidResetToken), true; got != want {
			t.Fatalf("isInvalidTokenErr=%v, want=%v (err=%v)", got, want, err)
		}
	})

	t.Run("returns expired token error for expired token", func(t *testing.T) {
		dataStore := newMockStore()
		sender := &mockEmailSender{}
		username := screenjournal.Username("alice")
		token := screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef23")
		dataStore.entriesByToken[token.String()] = screenjournal.PasswordResetEntry{
			Username:  username,
			Token:     token,
			ExpiresAt: fixedNow.Add(-time.Hour),
		}

		resetter := passwordreset.New(dataStore, sender, func() time.Time { return fixedNow })

		err := resetter.Reset(username, token, screenjournal.PasswordHash("new-password-hash"))
		if got, want := errors.Is(err, passwordreset.ErrExpiredResetToken), true; got != want {
			t.Fatalf("isExpiredTokenErr=%v, want=%v (err=%v)", got, want, err)
		}
		if got, want := len(dataStore.entriesByToken), 0; got != want {
			t.Fatalf("remainingTokens=%d, want=%d", got, want)
		}
	})

	t.Run("rate limits after five failed attempts", func(t *testing.T) {
		dataStore := newMockStore()
		sender := &mockEmailSender{}
		username := screenjournal.Username("alice")
		token := screenjournal.NewPasswordResetTokenFromString("ABCDEFGHJKLMNPQRSTUVWXYZabcdef23")

		resetter := passwordreset.New(dataStore, sender, func() time.Time { return fixedNow })

		for range 5 {
			err := resetter.Reset(username, token, screenjournal.PasswordHash("new-password-hash"))
			if got, want := errors.Is(err, passwordreset.ErrInvalidResetToken), true; got != want {
				t.Fatalf("isInvalidTokenErr=%v, want=%v (err=%v)", got, want, err)
			}
		}

		err := resetter.Reset(username, token, screenjournal.PasswordHash("new-password-hash"))
		if got, want := errors.Is(err, passwordreset.ErrTooManyResetAttempts), true; got != want {
			t.Fatalf("isTooManyAttemptsErr=%v, want=%v (err=%v)", got, want, err)
		}
	})
}
