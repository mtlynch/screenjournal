package handlers

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"
)

//go:embed static/*
var staticFiles embed.FS

// Changing this token on process start invalidates every static asset ETag.
var staticETagVersion = fmt.Sprintf("%d", time.Now().UTC().UnixNano())

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
		etag := staticAssetETag(r.URL.Path)
		w.Header().Set("Cache-Control", "public, max-age=1800")
		w.Header().Set("ETag", etag)
		if requestHasMatchingETag(r, etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// Let the original handler serve the file. The FileServer will handle
		// Last-Modified and If-Modified-Since automatically.
		h.ServeHTTP(w, r)
	})
}

// staticAssetETag returns a stable ETag for one asset during a server run.
func staticAssetETag(path string) string {
	hash := sha256.Sum256([]byte(staticETagVersion + "\n" + path))
	return fmt.Sprintf(`"%s"`, hex.EncodeToString(hash[:]))
}

// requestHasMatchingETag reports whether the client already has this asset.
func requestHasMatchingETag(r *http.Request, etag string) bool {
	for _, rawValue := range strings.Split(r.Header.Get("If-None-Match"), ",") {
		if strings.TrimSpace(rawValue) == etag {
			return true
		}
	}
	return false
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
