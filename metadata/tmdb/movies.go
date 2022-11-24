package tmdb

import (
	"log"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/handlers/parse"
	"github.com/mtlynch/screenjournal/v2/metadata"
)

func (f tmdbFinder) GetMovieInfo(id screenjournal.TmdbID) (metadata.MovieInfo, error) {
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

	info.ReleaseDate, err = ParseReleaseDate(m.ReleaseDate)
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

	if len(m.PosterPath) > 0 {
		posterPath, err := ParseImagePath(m.PosterPath)
		if err != nil {
			log.Printf("failed to parse poster path (%s) from TMDB ID %v: %v", m.PosterPath, id, err)
		} else {
			info.PosterPath = posterPath
		}
	}

	if len(m.BackdropPath) > 0 {
		backdropPath, err := ParseImagePath(m.BackdropPath)
		if err != nil {
			log.Printf("failed to parse backdrop path (%s) from TMDB ID %v: %v", m.BackdropPath, id, err)
		} else {
			info.BackdropPath = backdropPath
		}
	}

	return info, nil
}
