-- Migration 011: Convert to STRICT tables and improve constraints

-- First create all new tables
CREATE TABLE movies_new (
    id INTEGER PRIMARY KEY,
    tmdb_id INTEGER UNIQUE,
    imdb_id TEXT CHECK (imdb_id IS NULL OR imdb_id LIKE 'tt%'),
    title TEXT NOT NULL,
    release_date TEXT CHECK (
        release_date IS NULL OR datetime(release_date) IS NOT NULL
    ),
    poster_path TEXT,
    backdrop_path TEXT,
    summary TEXT
) STRICT;

CREATE TABLE tv_shows_new (
    id INTEGER PRIMARY KEY,
    tmdb_id INTEGER UNIQUE,
    imdb_id TEXT CHECK (imdb_id IS NULL OR imdb_id LIKE 'tt%'),
    title TEXT NOT NULL,
    first_air_date TEXT CHECK (
        first_air_date IS NULL OR datetime(first_air_date) IS NOT NULL
    ),
    poster_path TEXT,
    backdrop_path TEXT,
    summary TEXT
) STRICT;

CREATE TABLE users_new (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE CHECK (length(username) >= 3),
    is_admin INTEGER NOT NULL CHECK (is_admin IN (0, 1)) DEFAULT 0,
    email TEXT NOT NULL UNIQUE CHECK (email LIKE '%@%.%'),
    password_hash TEXT NOT NULL CHECK (length(password_hash) > 0),
    created_time TEXT NOT NULL CHECK (
        datetime(created_time) IS NOT NULL
        AND datetime(created_time) >= datetime('2020-01-01')
    ),
    last_modified_time TEXT NOT NULL CHECK (
        datetime(last_modified_time) IS NOT NULL
        AND datetime(last_modified_time) >= datetime('2020-01-01')
    )
) STRICT;

CREATE TABLE reviews_new (
    id INTEGER PRIMARY KEY,
    movie_id INTEGER,
    tv_show_id INTEGER,
    tv_show_season INTEGER CHECK (tv_show_season IS NULL OR tv_show_season > 0),
    review_owner TEXT NOT NULL,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 10),
    blurb TEXT NOT NULL,
    watched_date TEXT NOT NULL CHECK (
        datetime(watched_date) IS NOT NULL
        AND datetime(watched_date) >= datetime('2020-01-01')
    ),
    created_time TEXT NOT NULL CHECK (
        datetime(created_time) IS NOT NULL
        AND datetime(created_time) >= datetime('2020-01-01')
    ),
    last_modified_time TEXT NOT NULL CHECK (
        datetime(last_modified_time) IS NOT NULL
        AND datetime(last_modified_time) >= datetime('2020-01-01')
    ),
    FOREIGN KEY (movie_id) REFERENCES movies_new (id),
    FOREIGN KEY (tv_show_id) REFERENCES tv_shows_new (id),
    FOREIGN KEY (review_owner) REFERENCES users_new (username),
    CHECK ((movie_id IS NULL) != (tv_show_id IS NULL)),
    CHECK ((tv_show_id IS NOT NULL) = (tv_show_season IS NOT NULL))
) STRICT;

CREATE TABLE review_comments_new (
    id INTEGER PRIMARY KEY,
    review_id INTEGER NOT NULL,
    comment_owner TEXT NOT NULL,
    comment_text TEXT NOT NULL CHECK (length(comment_text) > 0),
    created_time TEXT NOT NULL CHECK (
        datetime(created_time) IS NOT NULL
        AND datetime(created_time) >= datetime('2020-01-01')
    ),
    last_modified_time TEXT NOT NULL CHECK (
        datetime(last_modified_time) IS NOT NULL
        AND datetime(last_modified_time) >= datetime('2020-01-01')
    ),
    FOREIGN KEY (review_id) REFERENCES reviews_new (id),
    FOREIGN KEY (comment_owner) REFERENCES users_new (username)
) STRICT;

CREATE TABLE notification_preferences_new (
    username TEXT PRIMARY KEY,
    new_reviews INTEGER NOT NULL CHECK (new_reviews IN (0, 1)) DEFAULT 1,
    all_new_comments INTEGER NOT NULL CHECK (
        all_new_comments IN (0, 1)
    ) DEFAULT 1,
    comments_on_my_reviews INTEGER NOT NULL CHECK (
        comments_on_my_reviews IN (0, 1)
    ) DEFAULT 1,
    FOREIGN KEY (username) REFERENCES users_new (username)
) STRICT;

CREATE TABLE invites_new (
    id INTEGER PRIMARY KEY,
    invitee TEXT NOT NULL CHECK (length(invitee) > 0),
    code TEXT NOT NULL UNIQUE CHECK (length(code) > 0),
    created_time TEXT NOT NULL CHECK (
        datetime(created_time) IS NOT NULL
        AND datetime(created_time) >= datetime('2020-01-01')
    )
) STRICT;

-- Copy data to new tables
INSERT INTO movies_new
SELECT DISTINCT
    movies.id,
    movies.tmdb_id,
    movies.imdb_id,
    movies.title,
    movies.release_date,
    movies.poster_path,
    movies.backdrop_path,
    movies.summary
FROM movies
INNER JOIN reviews ON movies.id = reviews.movie_id;

INSERT INTO tv_shows_new
SELECT DISTINCT
    tv_shows.id,
    tv_shows.tmdb_id,
    tv_shows.imdb_id,
    tv_shows.title,
    tv_shows.first_air_date,
    tv_shows.poster_path,
    tv_shows.backdrop_path,
    tv_shows.summary
FROM tv_shows
INNER JOIN reviews ON tv_shows.id = reviews.tv_show_id;

INSERT INTO users_new
SELECT
    id,
    username,
    is_admin,
    email,
    password_hash,
    created_time,
    last_modified_time
FROM users;

INSERT INTO reviews_new
SELECT
    id,
    movie_id,
    tv_show_id,
    tv_show_season,
    review_owner,
    rating,
    blurb,
    watched_date,
    created_time,
    last_modified_time
FROM reviews;

INSERT INTO review_comments_new
SELECT
    id,
    review_id,
    comment_owner,
    comment_text,
    created_time,
    last_modified_time
FROM review_comments;

INSERT INTO notification_preferences_new
SELECT
    username,
    new_reviews,
    all_new_comments,
    comments_on_my_reviews
FROM notification_preferences;

INSERT INTO invites_new
SELECT
    id,
    invitee,
    code,
    created_time
FROM invites;

-- Drop old tables
DROP TABLE invites;
DROP TABLE notification_preferences;
DROP TABLE review_comments;
DROP TABLE reviews;
DROP TABLE users;
DROP TABLE tv_shows;
DROP TABLE movies;

-- Rename new tables
ALTER TABLE movies_new RENAME TO movies;
ALTER TABLE tv_shows_new RENAME TO tv_shows;
ALTER TABLE users_new RENAME TO users;
ALTER TABLE reviews_new RENAME TO reviews;
ALTER TABLE review_comments_new RENAME TO review_comments;
ALTER TABLE notification_preferences_new RENAME TO notification_preferences;
ALTER TABLE invites_new RENAME TO invites;
