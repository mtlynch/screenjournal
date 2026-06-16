package passwordreset

import (
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
		InsertPasswordResetEntry(screenjournal.PasswordResetEntry) error
		UsePasswordResetEntry(
			screenjournal.Username,
			screenjournal.PasswordResetToken,
			screenjournal.PasswordHash,
			time.Time,
		) error
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

	if !r.passwordResetLimiter.HasAttemptsRemaining(user.Username) {
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

	r.passwordResetLimiter.RecordAttempt(user.Username)
	return nil
}

func (r Resetter) Reset(username screenjournal.Username, token screenjournal.PasswordResetToken, newPasswordHash screenjournal.PasswordHash) error {
	if !r.tokenAttemptLimiter.HasAttemptsRemaining(username) {
		log.Printf("password reset token attempt rate limited for user %s", username)
		return ErrTooManyResetAttempts
	}

	err := r.store.UsePasswordResetEntry(username, token, newPasswordHash, r.now())
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidPasswordResetToken):
			r.tokenAttemptLimiter.RecordAttempt(username)
			return ErrInvalidResetToken
		case errors.Is(err, store.ErrExpiredPasswordResetToken):
			r.tokenAttemptLimiter.RecordAttempt(username)
			return ErrExpiredResetToken
		}

		r.tokenAttemptLimiter.RecordAttempt(username)
		return fmt.Errorf("use password reset token %s: %w", tokenPrefix(token), err)
	}

	return nil
}

type noopEmailSender struct{}

func (noopEmailSender) Send(user screenjournal.User, entry screenjournal.PasswordResetEntry) error {
	log.Printf(
		"password reset email skipped for user %s (token %s)",
		user.Username,
		tokenPrefix(entry.Token),
	)
	return nil
}

func tokenPrefix(token screenjournal.PasswordResetToken) string {
	tokenRaw := token.String()
	const tokenPreviewLength = 6
	if len(tokenRaw) <= tokenPreviewLength {
		return tokenRaw
	}
	return tokenRaw[:tokenPreviewLength] + "..."
}
