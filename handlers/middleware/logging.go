package middleware

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// responseRecorder captures the status code and bytes written.
type responseRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

// LoggingHandler returns a handler that logs requests in Common Log Format.
func LoggingHandler(out io.Writer, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		h.ServeHTTP(rec, r)
		_, _ = fmt.Fprintf(out, "%s - - [%s] \"%s %s %s\" %d %d\n",
			r.RemoteAddr,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.RequestURI,
			r.Proto,
			rec.status,
			rec.size,
		)
	})
}
