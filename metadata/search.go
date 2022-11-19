package metadata

import "github.com/mtlynch/screenjournal/v2/handlers/parse"

func (f tmdbFinder) Search(query string) (MovieSearchResults, error) {
	tmdbResults, err := f.tmdbAPI.SearchMovie(query, map[string]string{
		"include_adult": "false",
	})
	if err != nil {
		return MovieSearchResults{}, err
	}
	results := MovieSearchResults{
		Matches:      make([]MovieSearchResult, len(tmdbResults.Results)),
		Page:         tmdbResults.Page,
		TotalPages:   tmdbResults.TotalPages,
		TotalResults: tmdbResults.TotalResults,
	}

	for i, match := range tmdbResults.Results {
		tmdbID, err := parse.TmdbID(match.ID)
		if err != nil {
			return MovieSearchResults{}, err
		}
		results.Matches[i] = MovieSearchResult{
			TmdbID:      tmdbID,
			Title:       match.Title,
			ReleaseDate: match.ReleaseDate,
		}

		if match.PosterPath != "" {
			results.Matches[i].PosterURL = "https://image.tmdb.org/t/p/w92" + match.PosterPath
		}

	}
	return results, nil
}
