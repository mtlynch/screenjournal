-- Migration 011: Convert to STRICT tables and improve constraints

-- Recreate movies table
CREATE TABLE movies2 (
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

INSERT INTO movies (
    id,
    tmdb_id,
    imdb_id,
    title,
    release_date,
    poster_path,
    backdrop_path,
    summary
)
SELECT
    id,
    tmdb_id,
    imdb_id,
    title,
    release_date,
    poster_path,
    backdrop_path,
    summary
FROM movies;

DROP TABLE movies;
ALTER TABLE movies2 RENAME TO movies;

-- Recreate tv_shows table
CREATE TABLE tv_shows2 (
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

INSERT INTO tv_shows (
    id,
    tmdb_id,
    imdb_id,
    title,
    first_air_date,
    poster_path,
    backdrop_path,
    summary
)
SELECT
    id,
    tmdb_id,
    imdb_id,
    title,
    first_air_date,
    poster_path,
    backdrop_path,
    summary
FROM tv_shows;

DROP TABLE tv_shows;
ALTER TABLE tv_shows2 RENAME TO tv_shows;

-- Recreate users table
CREATE TABLE users2 (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE CHECK (length(username) >= 3),
    is_admin INTEGER NOT NULL CHECK (is_admin IN (0, 1)) DEFAULT 0,
    email TEXT NOT NULL UNIQUE CHECK (email LIKE '%@%.%'),
    password_hash TEXT NOT NULL CHECK (length(password_hash) > 0),
    created_time TEXT NOT NULL CHECK (datetime(created_time) IS NOT NULL),
    last_modified_time TEXT NOT NULL CHECK (
        datetime(last_modified_time) IS NOT NULL
    )
) STRICT;

INSERT INTO users (
    id,
    username,
    is_admin,
    email,
    password_hash,
    created_time,
    last_modified_time
)
SELECT
    id,
    username,
    is_admin,
    email,
    password_hash,
    created_time,
    last_modified_time
FROM users;

DROP TABLE users;
ALTER TABLE users2 RENAME TO users;

-- Recreate reviews table
CREATE TABLE reviews2 (
    id INTEGER PRIMARY KEY,
    movie_id INTEGER,
    tv_show_id INTEGER,
    tv_show_season INTEGER CHECK (tv_show_season IS NULL OR tv_show_season > 0),
    review_owner TEXT NOT NULL,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 0 AND 10),
    blurb TEXT NOT NULL,
    watched_date TEXT NOT NULL CHECK (datetime(watched_date) IS NOT NULL),
    created_time TEXT NOT NULL CHECK (datetime(created_time) IS NOT NULL),
    last_modified_time TEXT NOT NULL CHECK (
        datetime(last_modified_time) IS NOT NULL
    ),
    FOREIGN KEY (movie_id) REFERENCES movies (id),
    FOREIGN KEY (tv_show_id) REFERENCES tv_shows (id),
    FOREIGN KEY (review_owner) REFERENCES users (username),
    CHECK ((movie_id IS NULL) != (tv_show_id IS NULL)),
    CHECK ((tv_show_id IS NOT NULL) = (tv_show_season IS NOT NULL))
) STRICT;

INSERT INTO reviews (
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
)
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

DROP TABLE reviews;
ALTER TABLE reviews2 RENAME TO reviews;

-- Recreate review_comments table
CREATE TABLE review_comments2 (
    id INTEGER PRIMARY KEY,
    review_id INTEGER NOT NULL,
    comment_owner TEXT NOT NULL,
    comment_text TEXT NOT NULL CHECK (length(comment_text) > 0),
    created_time TEXT NOT NULL CHECK (datetime(created_time) IS NOT NULL),
    last_modified_time TEXT NOT NULL CHECK (
        datetime(last_modified_time) IS NOT NULL
    ),
    FOREIGN KEY (review_id) REFERENCES reviews (id),
    FOREIGN KEY (comment_owner) REFERENCES users (username)
) STRICT;

INSERT INTO review_comments (
    id,
    review_id,
    comment_owner,
    comment_text,
    created_time,
    last_modified_time
)
SELECT
    id,
    review_id,
    comment_owner,
    comment_text,
    created_time,
    last_modified_time
FROM review_comments;

DROP TABLE review_comments;
ALTER TABLE review_comments2 RENAME TO review_comments;

-- Recreate notification_preferences table
CREATE TABLE notification_preferences2 (
    username TEXT PRIMARY KEY,
    new_reviews INTEGER NOT NULL CHECK (new_reviews IN (0, 1)) DEFAULT 1,
    all_new_comments INTEGER NOT NULL CHECK (
        all_new_comments IN (0, 1)
    ) DEFAULT 1,
    comments_on_my_reviews INTEGER NOT NULL CHECK (
        comments_on_my_reviews IN (0, 1)
    ) DEFAULT 1,
    FOREIGN KEY (username) REFERENCES users (username)
) STRICT;

INSERT INTO notification_preferences (
    username,
    new_reviews,
    all_new_comments,
    comments_on_my_reviews
)
SELECT
    username,
    new_reviews,
    all_new_comments,
    comments_on_my_reviews
FROM notification_preferences;

DROP TABLE notification_preferences;
ALTER TABLE notification_preferences2 RENAME TO notification_preferences;

-- Recreate invites table
CREATE TABLE invites2 (
    id INTEGER PRIMARY KEY,
    invitee TEXT NOT NULL CHECK (length(invitee) > 0),
    code TEXT NOT NULL UNIQUE CHECK (length(code) > 0),
    created_time TEXT NOT NULL CHECK (datetime(created_time) IS NOT NULL)
) STRICT;

INSERT INTO invites (
    id,
    invitee,
    code,
    created_time
)
SELECT
    id,
    invitee,
    code,
    created_time
FROM invites;

DROP TABLE invites;
ALTER TABLE invites2 RENAME TO invites;
