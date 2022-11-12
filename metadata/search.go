package metadata

func (f tmdbFinder) Search(query string) (MovieSearchResults, error) {
	tmdbResults, err := f.tmdbAPI.SearchMovie(query, map[string]string{
		"include_adult": "false",
	})
	if err != nil {
		return MovieSearchResults{}, err
	}
	results := MovieSearchResults{
		Matches:      make([]MovieStub, len(tmdbResults.Results)),
		Page:         tmdbResults.Page,
		TotalPages:   tmdbResults.TotalPages,
		TotalResults: tmdbResults.TotalResults,
	}

	for i, match := range tmdbResults.Results {
		results.Matches[i] = MovieStub{
			ID:          match.ID,
			Title:       match.Title,
			ReleaseDate: match.ReleaseDate,
			PosterURL:   "https://image.tmdb.org/t/p/w92" + match.PosterPath,
		}
	}
	return results, nil
}
