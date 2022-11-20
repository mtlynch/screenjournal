package metadata

import (
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
)

func (f tmdbFinder) Search(query string) (MovieSearchResults, error) {
	tmdbResults, err := f.tmdbAPI.SearchMovie(query, map[string]string{
		"include_adult": "false",
	})
	if err != nil {
		return MovieSearchResults{}, err
	}
	results := MovieSearchResults{
		Matches:      []MovieSearchResult{},
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

		tmdbID, err := parse.TmdbID(match.ID)
		if err != nil {
			return MovieSearchResults{}, err
		}
		results.Matches = append(results.Matches, MovieSearchResult{
			TmdbID:      tmdbID,
			Title:       match.Title,
			ReleaseDate: match.ReleaseDate,
			PosterPath:  match.PosterPath,
		})
	}

	return results, nil
}
