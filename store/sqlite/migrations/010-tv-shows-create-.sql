CREATE TABLE tv_shows (
    id INTEGER PRIMARY KEY,
    tmdb_id INTEGER UNIQUE,
    imdb_id TEXT,
    title TEXT NOT NULL,
    first_air_date TEXT,
    poster_path TEXT,
    backdrop_path TEXT,
    summary TEXT
);


-- Re-do reviews table to add tv_show_id column.
CREATE TABLE reviews2 (
    id INTEGER PRIMARY KEY,
    movie_id INTEGER,
    tv_show_id INTEGER,
    tv_show_season INTEGER,
    review_owner TEXT,
    rating INTEGER,
    blurb TEXT,
    watched_date TEXT,
    created_time TEXT,
    last_modified_time TEXT,
    FOREIGN KEY(movie_id) REFERENCES movies(id),
    FOREIGN KEY(tv_show_id) REFERENCES tv_shows(id),
    FOREIGN KEY(review_owner) REFERENCES users(username)
);

INSERT INTO reviews2
SELECT
    id,
    movie_id,
    NULL AS tv_show_id,
    NULL AS tv_show_season,
    review_owner,
    rating,
    blurb,
    watched_date,
    created_time,
    last_modified_time
FROM
    reviews;

DROP TABLE reviews;

ALTER TABLE reviews2
RENAME TO reviews;
