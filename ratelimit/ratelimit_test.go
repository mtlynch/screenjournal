package ratelimit_test

import (
	"testing"
	"time"

	"github.com/mtlynch/screenjournal/v2/ratelimit"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestPasswordResetLimiter(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	type recordedEvent struct {
		username screenjournal.Username
		offset   time.Duration
	}

	for _, tt := range []struct {
		description    string
		recordedEvents []recordedEvent
		queryUser      screenjournal.Username
		queryOffset    time.Duration
		want           bool
	}{
		{
			description:    "first request is allowed",
			recordedEvents: []recordedEvent{},
			queryUser:      screenjournal.Username("alice"),
			want:           true,
		},
		{
			description: "second request for same user is allowed",
			recordedEvents: []recordedEvent{
				{username: screenjournal.Username("alice")},
			},
			queryUser: screenjournal.Username("alice"),
			want:      true,
		},
		{
			description: "third request for same user is blocked",
			recordedEvents: []recordedEvent{
				{username: screenjournal.Username("alice")},
				{username: screenjournal.Username("alice")},
			},
			queryUser: screenjournal.Username("alice"),
			want:      false,
		},
		{
			description: "per-user limit resets after 24 hours",
			recordedEvents: []recordedEvent{
				{username: screenjournal.Username("alice")},
				{username: screenjournal.Username("alice")},
			},
			queryUser:   screenjournal.Username("alice"),
			queryOffset: 24*time.Hour + time.Second,
			want:        true,
		},
		{
			description: "global limit of 8 blocks ninth user",
			recordedEvents: []recordedEvent{
				{username: screenjournal.Username("user0")},
				{username: screenjournal.Username("user1")},
				{username: screenjournal.Username("user2")},
				{username: screenjournal.Username("user3")},
				{username: screenjournal.Username("user4")},
				{username: screenjournal.Username("user5")},
				{username: screenjournal.Username("user6")},
				{username: screenjournal.Username("user7")},
			},
			queryUser: screenjournal.Username("newuser"),
			want:      false,
		},
		{
			description: "global limit resets after 24 hours",
			recordedEvents: []recordedEvent{
				{username: screenjournal.Username("user0")},
				{username: screenjournal.Username("user1")},
				{username: screenjournal.Username("user2")},
				{username: screenjournal.Username("user3")},
				{username: screenjournal.Username("user4")},
				{username: screenjournal.Username("user5")},
				{username: screenjournal.Username("user6")},
				{username: screenjournal.Username("user7")},
			},
			queryUser:   screenjournal.Username("newuser"),
			queryOffset: 24*time.Hour + time.Second,
			want:        true,
		},
		{
			description: "per-user blocked but other users still allowed",
			recordedEvents: []recordedEvent{
				{username: screenjournal.Username("alice")},
				{username: screenjournal.Username("alice")},
			},
			queryUser: screenjournal.Username("bob"),
			want:      true,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			now := baseTime
			limiter := ratelimit.NewPasswordResetLimiter(func() time.Time { return now })

			for _, e := range tt.recordedEvents {
				now = baseTime.Add(e.offset)
				limiter.Record(e.username)
			}

			now = baseTime.Add(tt.queryOffset)
			if got, want := limiter.Allow(tt.queryUser), tt.want; got != want {
				t.Errorf("Allow(%s)=%v, want=%v", tt.queryUser, got, want)
			}
		})
	}
}
