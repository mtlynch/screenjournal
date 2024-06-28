package tmdb

import (
	"net/url"

	"github.com/mtlynch/screenjournal/v2/metadata"
)

func (f Finder) Search(query string) ([]metadata.MovieInfo, error) {
	tmdbResults, err := f.tmdbAPI.SearchMovie(query, map[string]string{
		"include_adult": "false",
	})
	if err != nil {
		return []metadata.MovieInfo{}, err
	}

	matches := []metadata.MovieInfo{}
	for _, match := range tmdbResults.Results {
		info := metadata.MovieInfo{}

		info.TmdbID, err = ParseTmdbID(match.ID)
		if err != nil {
			return []metadata.MovieInfo{}, err
		}

		if match.ReleaseDate == "" {
			continue
		}
		info.ReleaseDate, err = ParseReleaseDate(match.ReleaseDate)
		if err != nil {
			return []metadata.MovieInfo{}, err
		}

		if match.PosterPath == "" {
			continue
		}
		pp, err := url.Parse(match.PosterPath)
		if err != nil {
			return []metadata.MovieInfo{}, err
		}
		info.PosterPath = *pp

		matches = append(matches, info)
	}

	return matches, nil
}
