package sqlite

import (
	"database/sql"
	"log"
	"net/url"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (s Store) ReadTvShow(id screenjournal.TvShowID) (screenjournal.TvShow, error) {
	row := s.ctx.QueryRow(`
	SELECT
		id,
		tmdb_id,
		imdb_id,
		title,
		first_air_date,
		poster_path
	FROM
		tv_shows
	WHERE
		id = :id`, sql.Named("id", id.Int64()))

	return tvShowFromRow(row)
}

func (s Store) ReadTvShowByTmdbID(tmdbID screenjournal.TmdbID) (screenjournal.TvShow, error) {
	row := s.ctx.QueryRow(`
	SELECT
		id,
		tmdb_id,
		imdb_id,
		title,
		first_air_date,
		poster_path
	FROM
		tv_shows
	WHERE
		tmdb_id = :tmdb_id`, sql.Named("tmdb_id", tmdbID.Int32()))

	return tvShowFromRow(row)
}

func (s Store) InsertTvShow(t screenjournal.TvShow) (screenjournal.TvShowID, error) {
	log.Printf("inserting new TV show %s", t.Title)

	res, err := s.ctx.Exec(`
	INSERT INTO
		tv_shows
	(
		tmdb_id,
		imdb_id,
		title,
		first_air_date,
		poster_path
	)
	VALUES (
		:tmdb_id, :imdb_id, :title, :first_air_date, :poster_path
	)`,
		sql.Named("tmdb_id", t.TmdbID),
		sql.Named("imdb_id", t.ImdbID),
		sql.Named("title", t.Title),
		sql.Named("first_air_date", formatReleaseDate(t.AirDate)),
		sql.Named("poster_path", t.PosterPath.String()),
	)
	if err != nil {
		return screenjournal.TvShowID(0), err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return screenjournal.TvShowID(0), err
	}

	return screenjournal.TvShowID(lastID), nil
}

func (s Store) UpdateTvShow(t screenjournal.TvShow) error {
	log.Printf("updating TV show %s", t.Title)

	_, err := s.ctx.Exec(`
	UPDATE
		tv_shows
	SET
		tmdb_id = :tmdb_id,
		imdb_id = :imdb_id,
		title = :title,
		first_air_date = :first_air_date,
		poster_path = :poster_path
	WHERE tmdb_id = :tmdb_id
	`,
		sql.Named("tmdb_id", t.TmdbID),
		sql.Named("imdb_id", t.ImdbID),
		sql.Named("title", t.Title),
		sql.Named("first_air_date", formatReleaseDate(t.AirDate)),
		sql.Named("poster_path", t.PosterPath.String()),
	)
	if err != nil {
		return err
	}

	return nil
}

func tvShowFromRow(row rowScanner) (screenjournal.TvShow, error) {
	var id int
	var tmdbID int
	var imdbIDRaw *string
	var title string
	var firstAirDateRaw *string
	var posterPathRaw *string

	err := row.Scan(&id, &tmdbID, &imdbIDRaw, &title, &firstAirDateRaw, &posterPathRaw)
	if err == sql.ErrNoRows {
		return screenjournal.TvShow{}, store.ErrTvShowNotFound
	} else if err != nil {
		log.Printf("failed to read TV show from row: %v", err)
		return screenjournal.TvShow{}, err
	}

	var imdbID screenjournal.ImdbID
	if imdbIDRaw != nil {
		imdbID = screenjournal.ImdbID(*imdbIDRaw)
	}

	var firstAirDate screenjournal.ReleaseDate
	if firstAirDateRaw != nil {
		rd, err := parseDatetime(*firstAirDateRaw)
		if err != nil {
			log.Printf("failed to parse release date %s: %v", *firstAirDateRaw, err)
		} else {
			firstAirDate = screenjournal.ReleaseDate(rd)
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

	return screenjournal.TvShow{
		ID:         screenjournal.TvShowID(id),
		TmdbID:     screenjournal.TmdbID(tmdbID),
		ImdbID:     imdbID,
		Title:      screenjournal.MediaTitle(title),
		AirDate:    firstAirDate,
		PosterPath: posterPath,
	}, nil
}
