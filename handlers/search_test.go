package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/store/test_sqlite"
)

func TestSearch(t *testing.T) {
	for _, tt := range []struct {
		description string
	}{
		{
			description: "valid request with all fields populated and movie information is in local DB",
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			// TODO: Finish test.
			dataStore := test_sqlite.New()

			s := handlers.New(mockAuthenticator{}, nil, nil, dataStore, nil)

			req, err := http.NewRequest("GET", "/api/search?query=foo", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Accept", "text/json")
			req.AddCookie(&http.Cookie{
				Name: mockSessionTokenName,
			})

			w := httptest.NewRecorder()
			s.Router().ServeHTTP(w, req)

			if status := w.Code; status != http.StatusOK {
				t.Fatalf("%s: handler returned wrong status code: got %v want %v",
					tt.description, status, http.StatusOK)
			}
		})
	}
}
