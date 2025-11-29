package handlers

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

// noDirFS wraps an fs.FS to disable directory listing.
type noDirFS struct {
	fs fs.FS
}

func (n noDirFS) Open(name string) (fs.File, error) {
	f, err := n.fs.Open(name)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	if stat.IsDir() {
		_ = f.Close()
		return nil, fs.ErrNotExist
	}

	return f, nil
}

// cachingFileServer wraps an http.Handler to add caching headers.
func cachingFileServer(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add cache control headers.
		w.Header().Set("Cache-Control", "public, max-age=604800") // 1 week.

		// Let the original handler serve the file. The FileServer will handle
		// Last-Modified and If-Modified-Since automatically.
		h.ServeHTTP(w, r)
	})
}

// getStaticFilesHandler returns an http.Handler that serves static files from
// the embedded filesystem.
func getStaticFilesHandler() http.Handler {
	// Get the static subdirectory as a filesystem.
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	// Wrap the file server with our caching middleware.
	return cachingFileServer(http.FileServer(http.FS(noDirFS{staticFS})))
}
