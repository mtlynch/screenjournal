package sqlite

import (
	"log"

	"github.com/mtlynch/screenjournal/v2"
)

func (d DB) InsertMovie(m screenjournal.Movie) (screenjournal.MediaID, error) {
	log.Printf("inserting new movie %s", m.Title)

	res, err := d.ctx.Exec(`
	INSERT INTO
		movies
	(
		tmdb_id,
    imdb_id,
    title,
    poster_path
	)
	VALUES (
		?, ?, ?, ?
	)
	`,
		m.TmdbID,
		m.ImdbID,
		m.Title,
		m.PosterPath,
	)
	if err != nil {
		return screenjournal.MediaID(0), err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return screenjournal.MediaID(0), err
	}

	log.Printf("inserted! movie ID=%v", lastID) // DEBUG

	return screenjournal.MediaID(lastID), nil
}
