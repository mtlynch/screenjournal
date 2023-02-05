package sqlite

import (
	"database/sql"
	"log"
	"net/url"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (db DB) ReadMovie(id screenjournal.MovieID) (screenjournal.Movie, error) {
	row := db.ctx.QueryRow(`
	SELECT
		id,
		tmdb_id,
		imdb_id,
		title,
		release_date,
		poster_path
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
		imdb_id,
		title,
		release_date,
		poster_path
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
		imdb_id,
		title,
		release_date,
		poster_path
	)
	VALUES (
		?, ?, ?, ?, ?
	)`,
		m.TmdbID,
		m.ImdbID,
		m.Title,
		formatReleaseDate(m.ReleaseDate),
		m.PosterPath.String(),
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
		title = ?,
		imdb_id = ?,
		release_date = ?,
		poster_path = ?
	WHERE
		id = ?`,
		m.Title,
		m.ImdbID,
		formatReleaseDate(m.ReleaseDate),
		m.PosterPath.String(),
		m.ID.Int64()); err != nil {
		return err
	}

	return nil
}

func movieFromRow(row rowScanner) (screenjournal.Movie, error) {
	var id int
	var tmdbID int
	var imdbIDRaw *string
	var title string
	var releaseDateRaw *string
	var posterPathRaw *string

	err := row.Scan(&id, &tmdbID, &imdbIDRaw, &title, &releaseDateRaw, &posterPathRaw)
	if err == sql.ErrNoRows {
		return screenjournal.Movie{}, store.ErrMovieNotFound
	} else if err != nil {
		return screenjournal.Movie{}, err
	}

	var imdbID screenjournal.ImdbID
	if imdbIDRaw != nil {
		imdbID = screenjournal.ImdbID(*imdbIDRaw)
	}

	var releaseDate screenjournal.ReleaseDate
	if releaseDateRaw != nil {
		rd, err := parseDatetime(*releaseDateRaw)
		if err != nil {
			log.Printf("failed to parse release date %s: %v", *releaseDateRaw, err)
		} else {
			releaseDate = screenjournal.ReleaseDate(rd)
		}
	}

	var posterPath url.URL
	if posterPathRaw != nil {
		pp, err := url.Parse(*posterPathRaw)
		if err != nil {
			log.Printf("failed to parse poster path: %s", *posterPathRaw)
		} else {
			posterPath = *pp
		}
	}

	return screenjournal.Movie{
		ID:          screenjournal.MovieID(id),
		TmdbID:      screenjournal.TmdbID(tmdbID),
		ImdbID:      imdbID,
		Title:       screenjournal.MediaTitle(title),
		ReleaseDate: releaseDate,
		PosterPath:  posterPath,
	}, nil
}
