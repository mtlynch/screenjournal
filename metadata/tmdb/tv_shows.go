package tmdb

import (
	"errors"
	"log"
	"net/url"
	"time"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func (f Finder) GetTvShow(id screenjournal.TmdbID) (screenjournal.TvShow, error) {
	m, err := f.tmdbAPI.GetTvInfo(int(id.Int32()), map[string]string{})
	if err != nil {
		return screenjournal.TvShow{}, err
	}

	tvShow := screenjournal.TvShow{
		TmdbID: id,
	}

	tvShow.Title, err = parse.MediaTitle(m.Name)
	if err != nil {
		return screenjournal.TvShow{}, err
	}

	imdbID, err := f.readImdbID(id)
	if err != nil {
		log.Printf("failed to read IMDB ID from TMDB ID %v: %v", id, err)
	} else {
		tvShow.ImdbID = imdbID
	}

	log.Printf("imdb id=%v", tvShow.ImdbID.String())

	if len(m.FirstAirDate) > 0 {
		ad, err := ParseReleaseDate(m.FirstAirDate)
		if err != nil {
			log.Printf("failed to parse air date (%s) from TMDB ID %v: %v", m.FirstAirDate, id, err)
		} else {
			tvShow.AirDate = ad
		}
	}

	tvShow.SeasonCount = uint8(1)
	for _, s := range m.Seasons {
		// Sometimes specials are listed as Season 0. (e.g., Friends)
		if s.SeasonNumber == 0 {
			continue
		}
		if s.Name == "Specials" {
			continue
		}
		// Some shows list empty seasons (e.g. Nobody Wants This)
		if s.EpisodeCount == 0 {
			continue
		}
		hasAiredFn := func(airDateRaw string) bool {
			airDate, err := ParseReleaseDate(airDateRaw)
			if err != nil {
				return false
			}
			return time.Now().After(airDate.Time())
		}
		if !hasAiredFn(s.AirDate) {
			continue
		}
		if s.SeasonNumber > int(tvShow.SeasonCount) {
			tvShow.SeasonCount = uint8(s.SeasonNumber)
		}
	}

	if len(m.PosterPath) > 0 {
		pp, err := url.Parse(m.PosterPath)
		if err != nil {
			log.Printf("failed to parse poster path (%s) from TMDB ID %v: %v", m.PosterPath, id, err)
		} else {
			tvShow.PosterPath = *pp
		}
	}

	return tvShow, nil
}

func (f Finder) readImdbID(id screenjournal.TmdbID) (screenjournal.ImdbID, error) {
	externalIDs, err := f.tmdbAPI.GetTvExternalIds(int(id.Int32()), map[string]string{})
	if err != nil {
		return screenjournal.ImdbID(""), errors.New("failed to read external IDs")
	}

	imdbID, err := ParseImdbID(externalIDs.ImdbID)
	if err != nil {
		return screenjournal.ImdbID(""), errors.New("failed to parse IMDB ID")
	}

	return imdbID, nil
}
