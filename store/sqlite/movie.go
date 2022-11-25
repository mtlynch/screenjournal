package sqlite

import (
	"log"

	"github.com/mtlynch/screenjournal/v2"
)

func (d DB) InsertMovie(m screenjournal.Movie) (screenjournal.MovieID, error) {
	log.Printf("inserting new movie %s", m.Title)

	res, err := d.ctx.Exec(`
	INSERT INTO
		movies
	(
		tmdb_id,
		title
	)
	VALUES (
		?, ?
	)`,
		m.TmdbID,
		m.Title,
	)
	if err != nil {
		return screenjournal.MovieID(0), err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return screenjournal.MovieID(0), err
	}

	return screenjournal.MovieID(lastID), nil
}
