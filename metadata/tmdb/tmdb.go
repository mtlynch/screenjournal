package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Response types for TMDB API.

type MovieResponse struct {
	Title       string `json:"title"`
	ImdbID      string `json:"imdb_id"`
	ReleaseDate string `json:"release_date"`
	PosterPath  string `json:"poster_path"`
}

type TvResponse struct {
	Name         string     `json:"name"`
	FirstAirDate string     `json:"first_air_date"`
	PosterPath   string     `json:"poster_path"`
	Seasons      []TvSeason `json:"seasons"`
}

type TvSeason struct {
	SeasonNumber int    `json:"season_number"`
	Name         string `json:"name"`
	EpisodeCount int    `json:"episode_count"`
	AirDate      string `json:"air_date"`
}

type TvExternalIDs struct {
	ImdbID string `json:"imdb_id"`
}

type MovieSearchResult struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	PosterPath  string `json:"poster_path"`
}

type MovieSearchResults struct {
	Results []MovieSearchResult `json:"results"`
}

type TvSearchResult struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	FirstAirDate string `json:"first_air_date"`
	PosterPath   string `json:"poster_path"`
}

type TvSearchResults struct {
	Results []TvSearchResult `json:"results"`
}

type tmdbAPI interface {
	GetMovieInfo(id int) (*MovieResponse, error)
	GetTvInfo(id int) (*TvResponse, error)
	GetTvExternalIds(id int) (*TvExternalIDs, error)
	SearchMovie(query string) (*MovieSearchResults, error)
	SearchTv(query string) (*TvSearchResults, error)
}

type apiClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

func newAPIClient(apiKey string) *apiClient {
	return &apiClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.themoviedb.org/3",
	}
}

func (c *apiClient) get(rawURL string, result interface{}) error {
	resp, err := c.httpClient.Get(rawURL)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *apiClient) GetMovieInfo(id int) (*MovieResponse, error) {
	u := fmt.Sprintf("%s/movie/%d?api_key=%s", c.baseURL, id, c.apiKey)
	var result MovieResponse
	if err := c.get(u, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *apiClient) GetTvInfo(id int) (*TvResponse, error) {
	u := fmt.Sprintf("%s/tv/%d?api_key=%s", c.baseURL, id, c.apiKey)
	var result TvResponse
	if err := c.get(u, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *apiClient) GetTvExternalIds(id int) (*TvExternalIDs, error) {
	u := fmt.Sprintf("%s/tv/%d/external_ids?api_key=%s", c.baseURL, id, c.apiKey)
	var result TvExternalIDs
	if err := c.get(u, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *apiClient) SearchMovie(query string) (*MovieSearchResults, error) {
	u := fmt.Sprintf("%s/search/movie?api_key=%s&query=%s&include_adult=false", c.baseURL, c.apiKey, url.QueryEscape(query))
	var result MovieSearchResults
	if err := c.get(u, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *apiClient) SearchTv(query string) (*TvSearchResults, error) {
	u := fmt.Sprintf("%s/search/tv?api_key=%s&query=%s&include_adult=false", c.baseURL, c.apiKey, url.QueryEscape(query))
	var result TvSearchResults
	if err := c.get(u, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type Finder struct {
	tmdbAPI tmdbAPI
}

func New(apiKey string) (Finder, error) {
	return NewWithAPI(newAPIClient(apiKey)), nil
}

func NewWithAPI(api tmdbAPI) Finder {
	return Finder{
		tmdbAPI: api,
	}
}
