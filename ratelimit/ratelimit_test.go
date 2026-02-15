package ratelimit_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2/ratelimit"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestPasswordResetLimiter(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	for _, tt := range []struct {
		description string
		setup       func(*ratelimit.PasswordResetLimiter, *time.Time)
		queryUser   screenjournal.Username
		queryTime   time.Time
		want        bool
	}{
		{
			description: "first request is allowed",
			setup:       func(l *ratelimit.PasswordResetLimiter, now *time.Time) {},
			queryUser:   screenjournal.Username("alice"),
			queryTime:   baseTime,
			want:        true,
		},
		{
			description: "second request for same user is allowed",
			setup: func(l *ratelimit.PasswordResetLimiter, now *time.Time) {
				l.Record(screenjournal.Username("alice"))
			},
			queryUser: screenjournal.Username("alice"),
			queryTime: baseTime,
			want:      true,
		},
		{
			description: "third request for same user is blocked",
			setup: func(l *ratelimit.PasswordResetLimiter, now *time.Time) {
				l.Record(screenjournal.Username("alice"))
				l.Record(screenjournal.Username("alice"))
			},
			queryUser: screenjournal.Username("alice"),
			queryTime: baseTime,
			want:      false,
		},
		{
			description: "per-user limit resets after 24 hours",
			setup: func(l *ratelimit.PasswordResetLimiter, now *time.Time) {
				l.Record(screenjournal.Username("alice"))
				l.Record(screenjournal.Username("alice"))
			},
			queryUser: screenjournal.Username("alice"),
			queryTime: baseTime.Add(24*time.Hour + time.Second),
			want:      true,
		},
		{
			description: "global limit of 8 blocks ninth user",
			setup: func(l *ratelimit.PasswordResetLimiter, now *time.Time) {
				for i := 0; i < 8; i++ {
					l.Record(screenjournal.Username(fmt.Sprintf("user%d", i)))
				}
			},
			queryUser: screenjournal.Username("newuser"),
			queryTime: baseTime,
			want:      false,
		},
		{
			description: "global limit resets after 24 hours",
			setup: func(l *ratelimit.PasswordResetLimiter, now *time.Time) {
				for i := 0; i < 8; i++ {
					l.Record(screenjournal.Username(fmt.Sprintf("user%d", i)))
				}
			},
			queryUser: screenjournal.Username("newuser"),
			queryTime: baseTime.Add(24*time.Hour + time.Second),
			want:      true,
		},
		{
			description: "per-user blocked but other users still allowed",
			setup: func(l *ratelimit.PasswordResetLimiter, now *time.Time) {
				l.Record(screenjournal.Username("alice"))
				l.Record(screenjournal.Username("alice"))
			},
			queryUser: screenjournal.Username("bob"),
			queryTime: baseTime,
			want:      true,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			now := baseTime
			limiter := ratelimit.NewPasswordResetLimiter(func() time.Time { return now })
			tt.setup(limiter, &now)
			now = tt.queryTime
			if got, want := limiter.Allow(tt.queryUser), tt.want; got != want {
				t.Errorf("Allow(%s)=%v, want=%v", tt.queryUser, got, want)
			}
		})
	}
}
