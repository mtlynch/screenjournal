package middleware

import (
	"net"
	"net/http"
	"strings"
)

// ProxyIPHeaders returns a handler that sets r.RemoteAddr from proxy headers.
func ProxyIPHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ip := realIP(r); ip != "" {
			r.RemoteAddr = ip
		}
		h.ServeHTTP(w, r)
	})
}

func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the comma-separated list.
		if i := strings.IndexByte(xff, ','); i > 0 {
			xff = xff[:i]
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	// Parse the Forwarded header (RFC 7239).
	if fwd := r.Header.Get("Forwarded"); fwd != "" {
		for _, part := range strings.Split(fwd, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(strings.ToLower(part), "for=") {
				addr := part[4:]
				addr = strings.Trim(addr, `"`)
				// Handle IPv6 with brackets.
				host, _, err := net.SplitHostPort(addr)
				if err == nil {
					return host
				}
				return addr
			}
		}
	}
	return ""
}
