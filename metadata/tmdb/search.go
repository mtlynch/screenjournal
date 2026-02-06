package tmdb

import (
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func (f Finder) SearchMovies(query screenjournal.SearchQuery) ([]metadata.SearchResult, error) {
	tmdbResults, err := f.tmdbAPI.SearchMovie(query.String())
	if err != nil {
		return []metadata.SearchResult{}, err
	}

	matches := []metadata.SearchResult{}
	for _, match := range tmdbResults.Results {
		info := metadata.SearchResult{}

		info.TmdbID, err = parse.TmdbID(match.ID)
		if err != nil {
			return []metadata.SearchResult{}, err
		}

		info.Title, err = parse.MediaTitle(match.Title)
		if err != nil {
			return []metadata.SearchResult{}, err
		}

		if match.ReleaseDate == "" {
			continue
		}
		info.ReleaseDate, err = ParseReleaseDate(match.ReleaseDate)
		if err != nil {
			return []metadata.SearchResult{}, err
		}

		if match.PosterPath == "" {
			continue
		}
		info.PosterPath, err = parse.PosterPath(match.PosterPath)
		if err != nil {
			return []metadata.SearchResult{}, err
		}

		matches = append(matches, info)
	}

	return matches, nil
}

func (f Finder) SearchTvShows(query screenjournal.SearchQuery) ([]metadata.SearchResult, error) {
	tmdbResults, err := f.tmdbAPI.SearchTv(query.String())
	if err != nil {
		return []metadata.SearchResult{}, err
	}

	matches := []metadata.SearchResult{}
	for _, match := range tmdbResults.Results {
		info := metadata.SearchResult{}

		info.TmdbID, err = parse.TmdbID(match.ID)
		if err != nil {
			return []metadata.SearchResult{}, err
		}

		info.Title, err = parse.MediaTitle(match.Name)
		if err != nil {
			return []metadata.SearchResult{}, err
		}

		if match.FirstAirDate == "" {
			continue
		}
		info.ReleaseDate, err = ParseReleaseDate(match.FirstAirDate)
		if err != nil {
			return []metadata.SearchResult{}, err
		}

		if match.PosterPath == "" {
			continue
		}
		info.PosterPath, err = parse.PosterPath(match.PosterPath)
		if err != nil {
			return []metadata.SearchResult{}, err
		}

		matches = append(matches, info)
	}

	return matches, nil
}
