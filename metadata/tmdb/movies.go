package tmdb

import (
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

	return info, nil
}
