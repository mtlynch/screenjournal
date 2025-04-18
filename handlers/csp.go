package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/mtlynch/screenjournal/v2/random"
)

var contextKeyCSPNonce = &contextKey{"csp-nonce"}

func enforceContentSecurityPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce := base64.StdEncoding.EncodeToString(random.Bytes(16))

		type cspDirective struct {
			name   string
			values []string
		}
		directives := []cspDirective{
			{
				name: "default-src",
				values: []string{
					"'self'",
				},
			},
			{
				name: "script-src-elem",
				values: []string{
					"'self'",
					"'nonce-" + nonce + "'",
				},
			},
			{
				name: "style-src-elem",
				values: []string{
					"'self'",
					"'nonce-" + nonce + "'",
					// for htmx 2.0.4 inline style
					"'sha256-bsV5JivYxvGywDAZ22EZJKBFip65Ng9xoJVLbBg7bdo='",
				},
			},
			{
				name: "img-src",
				values: []string{
					"'self'",
					"data:",
					"image.tmdb.org",
				},
			},
			{
				name: "media-src",
				values: []string{
					"'self'",
					"data:",
				},
			},
		}
		policyParts := []string{}
		for _, directive := range directives {
			policyParts = append(policyParts, fmt.Sprintf("%s %s", directive.name, strings.Join(directive.values, " ")))
		}
		policy := strings.Join(policyParts, "; ") + ";"

		w.Header().Set("Content-Security-Policy", policy)

		ctx := context.WithValue(r.Context(), contextKeyCSPNonce, nonce)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func cspNonce(ctx context.Context) string {
	key, ok := ctx.Value(contextKeyCSPNonce).(string)
	if !ok {
		panic("CSP nonce is missing from request context")
	}
	return key
}
