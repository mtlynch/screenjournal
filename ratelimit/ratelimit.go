package ratelimit

import (
	"sync"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

const (
	perUserLimit = 2
	globalLimit  = 8
	window       = 24 * time.Hour
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

// Allow reports whether a password reset email may be sent for the given user
// without exceeding the per-user (2/24h) or global (8/24h) limits.
func (l *PasswordResetLimiter) Allow(username screenjournal.Username) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.prune()

	var userCount, globalCount int
	for _, e := range l.events {
		globalCount++
		if e.username.Equal(username) {
			userCount++
		}
	}

	if userCount >= perUserLimit {
		return false
	}
	if globalCount >= globalLimit {
		return false
	}
	return true
}

// Record logs that a password reset email was sent for the given user.
func (l *PasswordResetLimiter) Record(username screenjournal.Username) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.events = append(l.events, event{
		username:  username,
		timestamp: l.now(),
	})
}

func (l *PasswordResetLimiter) prune() {
	cutoff := l.now().Add(-window)
	kept := l.events[:0]
	for _, e := range l.events {
		if !e.timestamp.Before(cutoff) {
			kept = append(kept, e)
		}
	}
	l.events = kept
}
