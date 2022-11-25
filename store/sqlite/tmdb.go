package sqlite

import (
	"database/sql"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (db DB) TmdbIDToLocalID(tmdbID screenjournal.TmdbID) (screenjournal.MovieID, error) {
	var id int
	if err := db.ctx.QueryRow(`
	SELECT
		id
	FROM
		movies
	WHERE
		tmdb_id = ?`, tmdbID.Int32()).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return screenjournal.MovieID(0), store.ErrTmdbIDNotFound
		}
		return screenjournal.MovieID(0), err
	}

	return screenjournal.MovieID(id), nil
}
