package passwordreset

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2/ratelimit"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

const passwordResetTokenExpiry = 7 * 24 * time.Hour

var (
	ErrTooManyResetAttempts = errors.New("too many password reset attempts")
	ErrInvalidResetToken    = errors.New("invalid or expired password reset token")
	ErrExpiredResetToken    = errors.New("password reset token has expired")
)

type (
	Store interface {
		ReadUserByEmail(screenjournal.Email) (screenjournal.User, error)
		ReadPasswordResetEntry(screenjournal.PasswordResetToken) (screenjournal.PasswordResetEntry, error)
		InsertPasswordResetEntry(screenjournal.PasswordResetEntry) error
		UpdateUserPassword(screenjournal.Username, screenjournal.PasswordHash) error
		DeletePasswordResetEntry(screenjournal.PasswordResetToken) error
	}

	emailSender interface {
		Send(screenjournal.User, screenjournal.PasswordResetEntry) error
	}

	Resetter struct {
		store                Store
		emailSender          emailSender
		passwordResetLimiter *ratelimit.PasswordResetLimiter
		tokenAttemptLimiter  *ratelimit.TokenAttemptLimiter
		now                  func() time.Time
	}
)

func New(store Store, sender emailSender, now func() time.Time) Resetter {
	if now == nil {
		panic("password resetter requires a clock")
	}
	return Resetter{
		store:                store,
		emailSender:          sender,
		passwordResetLimiter: ratelimit.NewPasswordResetLimiter(now),
		tokenAttemptLimiter:  ratelimit.NewTokenAttemptLimiter(now),
		now:                  now,
	}
}

func NewNoEmail(store Store, now func() time.Time) Resetter {
	return New(store, noopEmailSender{}, now)
}

func (r Resetter) WithStore(store Store) Resetter {
	r.store = store
	return r
}

func (r Resetter) SendEmail(emailAddr screenjournal.Email) error {
	user, err := r.store.ReadUserByEmail(emailAddr)
	if err != nil {
		if err == store.ErrUserNotFound {
			log.Printf("password reset requested for unregistered email")
			return nil
		}
		return fmt.Errorf("look up user by email: %w", err)
	}

	if !r.passwordResetLimiter.Allow(user.Username) {
		log.Printf("password reset rate limited for user %s", user.Username)
		return nil
	}

	entry := screenjournal.PasswordResetEntry{
		Username:  user.Username,
		Token:     screenjournal.NewPasswordResetToken(),
		ExpiresAt: r.now().Add(passwordResetTokenExpiry),
	}
	if err := r.store.InsertPasswordResetEntry(entry); err != nil {
		return fmt.Errorf("insert password reset entry: %w", err)
	}

	if err := r.emailSender.Send(user, entry); err != nil {
		return fmt.Errorf("send password reset email for user %s: %w", user.Username, err)
	}

	r.passwordResetLimiter.Record(user.Username)
	return nil
}

func (r Resetter) Reset(username screenjournal.Username, token screenjournal.PasswordResetToken, newPasswordHash screenjournal.PasswordHash) error {
	if !r.tokenAttemptLimiter.Allow(username) {
		log.Printf("password reset token attempt rate limited for user %s", username)
		return ErrTooManyResetAttempts
	}

	entry, err := r.store.ReadPasswordResetEntry(token)
	if err != nil {
		r.tokenAttemptLimiter.Record(username)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidResetToken
		}
		return fmt.Errorf("read password reset entry for token %s: %w", token, err)
	}

	if !entry.Username.Equal(username) {
		r.tokenAttemptLimiter.Record(username)
		return ErrInvalidResetToken
	}

	if r.now().After(entry.ExpiresAt) {
		if err := r.store.DeletePasswordResetEntry(token); err != nil {
			log.Printf("failed to delete expired password reset token %s: %v", token, err)
		}
		r.tokenAttemptLimiter.Record(username)
		return ErrExpiredResetToken
	}

	if err := r.store.UpdateUserPassword(entry.Username, newPasswordHash); err != nil {
		return fmt.Errorf("update password for user %s: %w", entry.Username, err)
	}

	if err := r.store.DeletePasswordResetEntry(token); err != nil {
		log.Printf("failed to delete used password reset token %s: %v", token, err)
	}

	return nil
}

type noopEmailSender struct{}

func (noopEmailSender) Send(user screenjournal.User, entry screenjournal.PasswordResetEntry) error {
	log.Printf("password reset email skipped for user %s (token %s)", user.Username, entry.Token)
	return nil
}
