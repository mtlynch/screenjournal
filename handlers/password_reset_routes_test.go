package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mtlynch/screenjournal/v2/auth"
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

type noopPasswordResetter struct{}

func (noopPasswordResetter) SendEmail(screenjournal.Email) error {
	return nil
}

func (noopPasswordResetter) Reset(screenjournal.Username, screenjournal.PasswordResetToken, screenjournal.PasswordHash) error {
	return nil
}

func TestResetPasswordGet(t *testing.T) {
	for _, tt := range []struct {
		description        string
		passwordResetter   handlers.PasswordResetter
		bodyMustContain    string
		bodyMustNotContain string
		expectedStatusCode int
	}{
		{
			description:        "shows unavailable message when server does not support resets",
			passwordResetter:   nil,
			bodyMustContain:    "Password resets are not available on this server.",
			bodyMustNotContain: `id="reset-password-form"`,
			expectedStatusCode: http.StatusOK,
		},
		{
			description:        "shows reset form when server supports resets",
			passwordResetter:   noopPasswordResetter{},
			bodyMustContain:    `id="reset-password-form"`,
			bodyMustNotContain: "Password resets are not available on this server.",
			expectedStatusCode: http.StatusOK,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()
			authenticator := auth.New(dataStore)
			sessionManager := newMockSessionManager([]mockSessionEntry{})
			s := handlers.New(
				authenticator,
				nilAnnouncer,
				&sessionManager,
				dataStore,
				nilMetadataFinder,
				tt.passwordResetter,
			)

			req := httptest.NewRequest(http.MethodGet, "/reset-password", nil)
			rec := httptest.NewRecorder()
			s.Router().ServeHTTP(rec, req)
			res := rec.Result()

			if got, want := res.StatusCode, tt.expectedStatusCode; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			body := rec.Body.String()
			if got, want := strings.Contains(body, tt.bodyMustContain), true; got != want {
				t.Errorf("bodyContains(%q)=%v, want=%v", tt.bodyMustContain, got, want)
			}
			if got, want := strings.Contains(body, tt.bodyMustNotContain), false; got != want {
				t.Errorf("bodyContains(%q)=%v, want=%v", tt.bodyMustNotContain, got, want)
			}
		})
	}
}

func TestResetPasswordPost(t *testing.T) {
	for _, tt := range []struct {
		description        string
		passwordResetter   handlers.PasswordResetter
		expectedStatusCode int
		bodyMustContain    string
	}{
		{
			description:        "returns service unavailable when server does not support resets",
			passwordResetter:   nil,
			expectedStatusCode: http.StatusServiceUnavailable,
			bodyMustContain:    "Password resets are not available on this server",
		},
		{
			description:        "renders success page when server supports resets",
			passwordResetter:   noopPasswordResetter{},
			expectedStatusCode: http.StatusOK,
			bodyMustContain:    "Reset successful!",
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()
			authenticator := auth.New(dataStore)
			sessionManager := newMockSessionManager([]mockSessionEntry{})
			s := handlers.New(
				authenticator,
				nilAnnouncer,
				&sessionManager,
				dataStore,
				nilMetadataFinder,
				tt.passwordResetter,
			)

			req := httptest.NewRequest(
				http.MethodPost,
				"/reset-password",
				strings.NewReader("email=user@example.com"),
			)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rec := httptest.NewRecorder()
			s.Router().ServeHTTP(rec, req)
			res := rec.Result()

			if got, want := res.StatusCode, tt.expectedStatusCode; got != want {
				t.Fatalf("httpStatus=%v, want=%v", got, want)
			}

			body := rec.Body.String()
			if got, want := strings.Contains(body, tt.bodyMustContain), true; got != want {
				t.Errorf("bodyContains(%q)=%v, want=%v", tt.bodyMustContain, got, want)
			}
		})
	}
}
