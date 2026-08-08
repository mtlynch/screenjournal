//go:build dev

package handlers

// staticCacheControl returns the Cache-Control header for static assets. Dev
// builds disable caching so freshly rebuilt JS/CSS is always served, rather than
// a stale copy the browser holds for max-age.
func staticCacheControl() string {
	return "no-store"
}
