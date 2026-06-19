//go:build dev

package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mtlynch/screenjournal/v2/handlers"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/metadata/tmdb"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var fakeMovies = []screenjournal.Movie{
	newFakeMovie(333287, "Slow Learners", "2015-08-19", "tt2597718", "/slow-learners.jpg"),
	newFakeMovie(745, "The Sixth Sense", "1999-08-06", "tt0167404", "/sixth-sense.jpg"),
	newFakeMovie(928344, "Weird: The Al Yankovic Story", "2022-09-08", "tt17076046", "/qcj2z13G0KjaIgc01ifiUKu7W07.jpg"),
	newFakeMovie(38, "Eternal Sunshine of the Spotless Mind", "2004-03-19", "tt0338013", "/eternal-sunshine.jpg"),
	newFakeMovie(544, "There's Something About Mary", "1998-07-15", "tt0129387", "/something-about-mary.jpg"),
	newFakeMovie(409, "The English Patient", "1996-11-14", "tt0116209", "/english-patient.jpg"),
	newFakeMovie(238, "The Godfather", "1972-03-14", "tt0068646", "/godfather.jpg"),
	newFakeMovie(10663, "The Waterboy", "1998-11-06", "tt0120484", "/miT42qWYC4D0n2mXNzJ9VfhheWW.jpg"),
	newFakeMovie(11017, "Billy Madison", "1995-02-10", "tt0112508", "/iwk9pWR6MwTInEQc8Vw5vGHjeQ0.jpg"),
}

var fakeTvShows = []screenjournal.TvShow{
	newFakeTvShow(4608, "30 Rock", "2006-10-11", "tt0496424", "/30-rock.jpg", 7),
	newFakeTvShow(1400, "Seinfeld", "1989-07-05", "tt0098904", "/aCw8ONfyz3AhngVQa1E2Ss4KSUQ.jpg", 9),
}

type fakeMetadataFinder struct{}

func newMetadataFinder() (handlers.MetadataFinder, error) {
	if os.Getenv("SJ_TMDB_API") == "dummy" {
		return fakeMetadataFinder{}, nil
	}
	return tmdb.New(requireEnv("SJ_TMDB_API"))
}

func (fakeMetadataFinder) SearchMovies(query screenjournal.SearchQuery) ([]metadata.SearchResult, error) {
	results := []metadata.SearchResult{}
	for _, movie := range fakeMovies {
		if !matchesQuery(movie.Title.String(), query.String()) {
			continue
		}
		results = append(results, metadata.SearchResult{
			TmdbID:      movie.TmdbID,
			Title:       movie.Title,
			ReleaseDate: movie.ReleaseDate,
			PosterPath:  movie.PosterPath,
		})
	}
	return results, nil
}

func (fakeMetadataFinder) SearchTvShows(query screenjournal.SearchQuery) ([]metadata.SearchResult, error) {
	results := []metadata.SearchResult{}
	for _, tvShow := range fakeTvShows {
		if !matchesQuery(tvShow.Title.String(), query.String()) {
			continue
		}
		results = append(results, metadata.SearchResult{
			TmdbID:      tvShow.TmdbID,
			Title:       tvShow.Title,
			ReleaseDate: tvShow.AirDate,
			PosterPath:  tvShow.PosterPath,
		})
	}
	return results, nil
}

func (fakeMetadataFinder) GetMovie(id screenjournal.TmdbID) (screenjournal.Movie, error) {
	for _, movie := range fakeMovies {
		if movie.TmdbID.Equal(id) {
			return movie, nil
		}
	}
	return screenjournal.Movie{}, fmt.Errorf("failed to find fake movie metadata for TMDB ID %s", id)
}

func (fakeMetadataFinder) GetTvShow(id screenjournal.TmdbID) (screenjournal.TvShow, error) {
	for _, tvShow := range fakeTvShows {
		if tvShow.TmdbID.Equal(id) {
			return tvShow, nil
		}
	}
	return screenjournal.TvShow{}, fmt.Errorf("failed to find fake TV show metadata for TMDB ID %s", id)
}

func matchesQuery(title string, query string) bool {
	return strings.Contains(strings.ToLower(title), strings.ToLower(query))
}

func newFakeMovie(
	tmdbID int32,
	title string,
	releaseDate string,
	imdbID string,
	posterPath string,
) screenjournal.Movie {
	return screenjournal.Movie{
		TmdbID:      screenjournal.TmdbID(tmdbID),
		Title:       screenjournal.MediaTitle(title),
		ReleaseDate: mustFakeReleaseDate(releaseDate),
		ImdbID:      screenjournal.ImdbID(imdbID),
		PosterPath:  mustFakePosterPath(posterPath),
	}
}

func newFakeTvShow(
	tmdbID int32,
	title string,
	airDate string,
	imdbID string,
	posterPath string,
	seasonCount uint8,
) screenjournal.TvShow {
	return screenjournal.TvShow{
		TmdbID:      screenjournal.TmdbID(tmdbID),
		Title:       screenjournal.MediaTitle(title),
		AirDate:     mustFakeReleaseDate(airDate),
		ImdbID:      screenjournal.ImdbID(imdbID),
		PosterPath:  mustFakePosterPath(posterPath),
		SeasonCount: seasonCount,
	}
}

func mustFakeReleaseDate(raw string) screenjournal.ReleaseDate {
	releaseDate, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		panic(err)
	}
	return screenjournal.ReleaseDate(releaseDate)
}

func mustFakePosterPath(raw string) url.URL {
	posterPath, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return *posterPath
}
