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
		description          string
		priorResetsForUser   int
		priorResetsForOthers int
		timeSincePriorResets time.Duration
		allowExpected        bool
	}{
		{
			description:   "first request is allowed",
			allowExpected: true,
		},
		{
			description:        "second request for same user is allowed",
			priorResetsForUser: 1,
			allowExpected:      true,
		},
		{
			description:        "third request for same user is blocked",
			priorResetsForUser: 2,
			allowExpected:      false,
		},
		{
			description:          "per-user limit resets after 24 hours",
			priorResetsForUser:   2,
			timeSincePriorResets: 24*time.Hour + time.Second,
			allowExpected:        true,
		},
		{
			description:          "global limit of 8 blocks new user",
			priorResetsForOthers: 8,
			allowExpected:        false,
		},
		{
			description:          "global limit resets after 24 hours",
			priorResetsForOthers: 8,
			timeSincePriorResets: 24*time.Hour + time.Second,
			allowExpected:        true,
		},
		{
			description:          "per-user blocked but other users still allowed",
			priorResetsForOthers: 2,
			allowExpected:        true,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			now := baseTime
			limiter := ratelimit.NewPasswordResetLimiter(func() time.Time { return now })

			queryUser := screenjournal.Username("alice")
			for range tt.priorResetsForUser {
				limiter.Record(queryUser)
			}
			for i := range tt.priorResetsForOthers {
				limiter.Record(screenjournal.Username(fmt.Sprintf("other-user-%d", i)))
			}

			now = baseTime.Add(tt.timeSincePriorResets)

			if got, want := limiter.Allow(queryUser), tt.allowExpected; got != want {
				t.Errorf("Allow(%s)=%v, want=%v", queryUser, got, want)
			}
		})
	}
}

func TestTokenAttemptLimiter(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	for _, tt := range []struct {
		description            string
		priorAttemptsForUser   int
		timeSincePriorAttempts time.Duration
		allowExpected          bool
	}{
		{
			description:   "first attempt is allowed",
			allowExpected: true,
		},
		{
			description:          "fifth attempt for same user is allowed",
			priorAttemptsForUser: 4,
			allowExpected:        true,
		},
		{
			description:          "sixth attempt for same user is blocked",
			priorAttemptsForUser: 5,
			allowExpected:        false,
		},
		{
			description:            "per-user limit resets after 24 hours",
			priorAttemptsForUser:   5,
			timeSincePriorAttempts: 24*time.Hour + time.Second,
			allowExpected:          true,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			now := baseTime
			limiter := ratelimit.NewTokenAttemptLimiter(func() time.Time { return now })

			queryUser := screenjournal.Username("alice")
			for range tt.priorAttemptsForUser {
				limiter.Record(queryUser)
			}

			now = baseTime.Add(tt.timeSincePriorAttempts)

			if got, want := limiter.Allow(queryUser), tt.allowExpected; got != want {
				t.Errorf("Allow(%s)=%v, want=%v", queryUser, got, want)
			}
		})
	}
}
