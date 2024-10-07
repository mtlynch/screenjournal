package tmdb

import (
	"net/url"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func (f Finder) Search(query screenjournal.SearchQuery) ([]metadata.MovieInfo, error) {
	tmdbResults, err := f.tmdbAPI.SearchMovie(query.String(), map[string]string{
		"include_adult": "false",
	})
	if err != nil {
		return []metadata.MovieInfo{}, err
	}

	matches := []metadata.MovieInfo{}
	for _, match := range tmdbResults.Results {
		info := metadata.MovieInfo{}

		info.TmdbID, err = parse.TmdbID(match.ID)
		if err != nil {
			return []metadata.MovieInfo{}, err
		}

		info.Title, err = parse.MediaTitle(match.Title)
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

func (f Finder) SearchTvShows(query screenjournal.SearchQuery) ([]metadata.TvShowInfo, error) {
	tmdbResults, err := f.tmdbAPI.SearchTv(query.String(), map[string]string{
		"include_adult": "false",
	})
	if err != nil {
		return []metadata.TvShowInfo{}, err
	}

	matches := []metadata.TvShowInfo{}
	for _, match := range tmdbResults.Results {
		info := metadata.TvShowInfo{}

		info.TmdbID, err = parse.TmdbID(match.ID)
		if err != nil {
			return []metadata.TvShowInfo{}, err
		}

		info.Title, err = parse.MediaTitle(match.Name)
		if err != nil {
			return []metadata.TvShowInfo{}, err
		}

		if match.FirstAirDate == "" {
			continue
		}
		info.ReleaseDate, err = ParseReleaseDate(match.FirstAirDate)
		if err != nil {
			return []metadata.TvShowInfo{}, err
		}

		if match.PosterPath == "" {
			continue
		}
		pp, err := url.Parse(match.PosterPath)
		if err != nil {
			return []metadata.TvShowInfo{}, err
		}
		info.PosterPath = *pp

		matches = append(matches, info)
	}

	return matches, nil
}
