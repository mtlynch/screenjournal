package tmdb

import (
	"log"
	"net/url"

	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/metadata"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func (f Finder) GetMovieInfo(id screenjournal.TmdbID) (metadata.MovieInfo, error) {
	m, err := f.tmdbAPI.GetMovieInfo(int(id.Int32()), map[string]string{})
	if err != nil {
		return metadata.MovieInfo{}, err
	}

	info := metadata.MovieInfo{
		TmdbID: id,
	}

	info.Title, err = parse.MediaTitle(m.Title)
	if err != nil {
		return metadata.MovieInfo{}, err
	}

	if len(m.ImdbID) > 0 {
		imdbID, err := ParseImdbID(m.ImdbID)
		if err != nil {
			log.Printf("failed to parse IMDB ID (%s) from TMDB ID %v: %v", m.ImdbID, id, err)
		} else {
			info.ImdbID = imdbID
		}
	}

	if len(m.ReleaseDate) > 0 {
		rd, err := ParseReleaseDate(m.ReleaseDate)
		if err != nil {
			log.Printf("failed to parse release date (%s) from TMDB ID %v: %v", m.ReleaseDate, id, err)
		} else {
			info.ReleaseDate = rd
		}
	}

	if len(m.PosterPath) > 0 {
		pp, err := url.Parse(m.PosterPath)
		if err != nil {
			log.Printf("failed to parse poster path (%s) from TMDB ID %v: %v", m.PosterPath, id, err)
		} else {
			info.PosterPath = *pp
		}
	}

	return info, nil
}
