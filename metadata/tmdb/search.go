package tmdb

import (
	"github.com/mtlynch/screenjournal/v2/metadata"
)

func (f Finder) Search(query string) (metadata.MovieSearchResults, error) {
	tmdbResults, err := f.tmdbAPI.SearchMovie(query, map[string]string{
		"include_adult": "false",
	})
	if err != nil {
		return metadata.MovieSearchResults{}, err
	}
	results := metadata.MovieSearchResults{
		Matches:      []metadata.MovieSearchResult{},
		Page:         tmdbResults.Page,
		TotalPages:   tmdbResults.TotalPages,
		TotalResults: tmdbResults.TotalResults,
	}

	for _, match := range tmdbResults.Results {
		// Exclude results that are not sufficiently populated.
		if match.ReleaseDate == "" {
			continue
		}
		if match.PosterPath == "" {
			continue
		}

		tmdbID, err := ParseTmdbID(match.ID)
		if err != nil {
			return metadata.MovieSearchResults{}, err
		}
		results.Matches = append(results.Matches, metadata.MovieSearchResult{
			TmdbID:      tmdbID,
			Title:       match.Title,
			ReleaseDate: match.ReleaseDate,
			PosterPath:  match.PosterPath,
		})
	}

	return results, nil
}
