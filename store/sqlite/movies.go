package sqlite

import (
	"database/sql"
	"log"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (db DB) ReadMovie(id screenjournal.MovieID) (screenjournal.Movie, error) {
	row := db.ctx.QueryRow(`
	SELECT
		id,
		tmdb_id,
		title
	FROM
		movies
	WHERE
		id = ?`, id.Int64())

	return movieFromRow(row)
}

func (db DB) ReadMovieByTmdbID(tmdbID screenjournal.TmdbID) (screenjournal.Movie, error) {
	row := db.ctx.QueryRow(`
	SELECT
		id,
		tmdb_id,
		title
	FROM
		movies
	WHERE
		tmdb_id = ?`, tmdbID.Int32())

	movie, err := movieFromRow(row)
	if err == sql.ErrNoRows {
		return screenjournal.Movie{}, ErrTmdbIDNotFound
	}

	return movie, nil
}

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

func movieFromRow(row rowScanner) (screenjournal.Movie, error) {
	var id int
	var tmdb_id int
	var title string

	err := row.Scan(&id, &tmdb_id, &title)
	if err == sql.ErrNoRows {
		return screenjournal.Movie{}, store.ErrMovieNotFound
	} else if err != nil {
		return screenjournal.Movie{}, err
	}

	return screenjournal.Movie{
		ID:     screenjournal.MovieID(id),
		TmdbID: screenjournal.TmdbID(tmdb_id),
		Title:  screenjournal.MediaTitle(title),
	}, nil
}
