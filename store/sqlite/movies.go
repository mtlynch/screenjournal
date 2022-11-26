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

	return movieFromRow(row)
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

func (db DB) UpdateMovie(m screenjournal.Movie) error {
	log.Printf("updating movie information for %s (id=%v)", m.Title, m.ID)

	if _, err := db.ctx.Exec(`
	UPDATE movies
	SET
		title = ?
	WHERE
		id = ?`,
		m.Title,
		m.ID.Int64()); err != nil {
		return err
	}

	return nil
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