CREATE TABLE tv_show_seasons (
    id INTEGER PRIMARY KEY,
    tv_show_id INTEGER NOT NULL,
    season_number INTEGER NOT NULL,
    poster_path TEXT,
    FOREIGN KEY (tv_show_id) REFERENCES tv_shows (id),
    UNIQUE (tv_show_id, season_number)
);
