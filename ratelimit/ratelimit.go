package ratelimit

import (
	"sync"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

const (
	// perUserLimit is the maximum number of password reset emails a single
	// user can receive within the rate limit window.
	perUserLimit = 2
	// globalLimit is the maximum number of password reset emails that can
	// be sent to all users combined within the rate limit window.
	globalLimit = 8
	// window is the sliding time window over which rate limits are enforced.
	window = 24 * time.Hour
)

type event struct {
	username  screenjournal.Username
	timestamp time.Time
}

// PasswordResetLimiter enforces rate limits on password reset emails.
type PasswordResetLimiter struct {
	mu     sync.Mutex
	events []event
	now    func() time.Time
}

// NewPasswordResetLimiter creates a limiter that uses the given function to
// determine the current time.
func NewPasswordResetLimiter(now func() time.Time) *PasswordResetLimiter {
	return &PasswordResetLimiter{
		now: now,
	}
}

// HasAttemptsRemaining reports whether a password reset email may be sent for
// the given user without exceeding the per-user (2/24h) or global (8/24h)
// limits.
func (l *PasswordResetLimiter) HasAttemptsRemaining(username screenjournal.Username) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.removeExpiredEvents()

	var userCount int
	for _, e := range l.events {
		if e.username.Equal(username) {
			userCount++
		}
	}

	if userCount >= perUserLimit {
		return false
	}
	if len(l.events) >= globalLimit {
		return false
	}
	return true
}

// RecordAttempt logs that a password reset email was sent for the given user.
func (l *PasswordResetLimiter) RecordAttempt(username screenjournal.Username) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.events = append(l.events, event{
		username:  username,
		timestamp: l.now(),
	})
}

func (l *PasswordResetLimiter) removeExpiredEvents() {
	cutoff := l.now().Add(-window)
	kept := l.events[:0]
	for _, e := range l.events {
		if !e.timestamp.Before(cutoff) {
			kept = append(kept, e)
		}
	}
	l.events = kept
}

const (
	// tokenAttemptPerUserLimit is the maximum number of password reset
	// token attempts a single user can make within the rate limit window.
	tokenAttemptPerUserLimit = 5
	// tokenAttemptWindow is the sliding time window over which token
	// attempt rate limits are enforced.
	tokenAttemptWindow = 24 * time.Hour
)

// TokenAttemptLimiter enforces rate limits on password reset token attempts.
type TokenAttemptLimiter struct {
	mu     sync.Mutex
	events []event
	now    func() time.Time
}

// NewTokenAttemptLimiter creates a limiter that uses the given function to
// determine the current time.
func NewTokenAttemptLimiter(now func() time.Time) *TokenAttemptLimiter {
	return &TokenAttemptLimiter{
		now: now,
	}
}

// HasAttemptsRemaining reports whether a password reset token attempt may
// proceed for the given user without exceeding the per-user (5/24h) limit.
func (l *TokenAttemptLimiter) HasAttemptsRemaining(username screenjournal.Username) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.removeExpiredEvents()

	var userCount int
	for _, e := range l.events {
		if e.username.Equal(username) {
			userCount++
		}
	}

	return userCount < tokenAttemptPerUserLimit
}

// RecordAttempt logs that a password reset token attempt was made for the
// given user.
func (l *TokenAttemptLimiter) RecordAttempt(username screenjournal.Username) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.events = append(l.events, event{
		username:  username,
		timestamp: l.now(),
	})
}

func (l *TokenAttemptLimiter) removeExpiredEvents() {
	cutoff := l.now().Add(-tokenAttemptWindow)
	kept := l.events[:0]
	for _, e := range l.events {
		if !e.timestamp.Before(cutoff) {
			kept = append(kept, e)
		}
	}
	l.events = kept
}
