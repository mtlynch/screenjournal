package metadata

import (
	"net/url"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	SearchResult struct {
		TmdbID      screenjournal.TmdbID
		Title       screenjournal.MediaTitle
		ReleaseDate screenjournal.ReleaseDate
		PosterPath  url.URL
	}
)
