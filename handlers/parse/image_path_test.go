package parse_test

import (
	"fmt"
	"testing"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func TestImagePath(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		id          screenjournal.ImagePath
		err         error
	}{
		{
			"valid image path",
			"/6FfCtAuVAW8XJjZ7eWeLibRLWTw.jpg",
			screenjournal.ImagePath("/6FfCtAuVAW8XJjZ7eWeLibRLWTw.jpg"),
			nil,
		},
		{
			"empty string is invalid",
			"",
			screenjournal.ImagePath(""),
			parse.ErrInvalidImagePath,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
			id, err := parse.ImagePath(tt.in)
			if got, want := err, tt.err; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}
			if got, want := id.String(), tt.id.String(); got != want {
				t.Errorf("imagePath=%s, want=%s", got, want)
			}
		})
	}
}
