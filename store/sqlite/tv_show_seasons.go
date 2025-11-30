package sqlite

import (
	"database/sql"
	"log"
	"net/url"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (s Store) InsertTvShowSeason(season screenjournal.TvShowSeasonInfo) error {
	_, err := s.ctx.Exec(`
	INSERT OR REPLACE INTO tv_show_seasons (
		tv_show_id,
		season_number,
		poster_path
	)
	VALUES (
		:tv_show_id, :season_number, :poster_path
	)`,
		sql.Named("tv_show_id", season.TvShowID.Int64()),
		sql.Named("season_number", season.SeasonNumber.UInt8()),
		sql.Named("poster_path", season.PosterPath.String()),
	)
	return err
}

func (s Store) ReadTvShowSeason(tvShowID screenjournal.TvShowID, seasonNumber screenjournal.TvShowSeason) (screenjournal.TvShowSeasonInfo, error) {
	row := s.ctx.QueryRow(`
	SELECT
		tv_show_id,
		season_number,
		poster_path
	FROM
		tv_show_seasons
	WHERE
		tv_show_id = :tv_show_id AND season_number = :season_number`,
		sql.Named("tv_show_id", tvShowID.Int64()),
		sql.Named("season_number", seasonNumber.UInt8()),
	)

	var tvShowIDRaw int64
	var seasonNumberRaw int
	var posterPathRaw string

	err := row.Scan(&tvShowIDRaw, &seasonNumberRaw, &posterPathRaw)
	if err == sql.ErrNoRows {
		return screenjournal.TvShowSeasonInfo{}, store.ErrTvShowSeasonNotFound
	}
	if err != nil {
		return screenjournal.TvShowSeasonInfo{}, err
	}

	var posterPath url.URL
	if posterPathRaw != "" {
		pp, err := url.Parse(posterPathRaw)
		if err != nil {
			log.Printf("failed to parse poster path: %s", posterPathRaw)
		} else {
			posterPath = *pp
		}
	}

	return screenjournal.TvShowSeasonInfo{
		TvShowID:     screenjournal.TvShowID(tvShowIDRaw),
		SeasonNumber: screenjournal.TvShowSeason(seasonNumberRaw),
		PosterPath:   posterPath,
	}, nil
}
