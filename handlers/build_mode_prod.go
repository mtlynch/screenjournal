//go:build !dev

package handlers

// staticCacheControl returns the Cache-Control header for static assets. Prod
// builds let the browser cache assets for a bounded window.
func staticCacheControl() string {
	return "public, max-age=1800"
}
