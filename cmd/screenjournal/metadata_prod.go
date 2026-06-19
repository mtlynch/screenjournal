//go:build !dev

package main

import (
	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
)

func newMetadataFinder() (handlers.MetadataFinder, error) {
	return tmdb.New(requireEnv("SJ_TMDB_API"))
}
