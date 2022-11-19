package sqlite

import "github.com/mtlynch/screenjournal/v2"

func (db DB) TmdbIDToLocalID(tmdbID screenjournal.TmdbID) (screenjournal.MediaID, error) {
	var id int
	if err := db.ctx.QueryRow(`
	SELECT
		id
	FROM
		movies
	WHERE
		tmdb_id = ?`, tmdbID.Int()).Scan(&id); err != nil {
		return screenjournal.MediaID(0), err
	}

	return screenjournal.MediaID(id), nil
}
