CREATE TABLE movies (
    id INTEGER PRIMARY KEY,
    tmdb_id INTEGER UNIQUE,
    imdb_id TEXT,
    title TEXT NOT NULL,
    release_date TEXT,
    poster_path TEXT,
    backdrop_path TEXT,
    summary TEXT
);

-- Move movie titles to movies table.
INSERT INTO movies
(
    title
)
SELECT title
FROM
    reviews;

ALTER TABLE reviews
ADD COLUMN movie_id
INTEGER;

-- Make reviews table reference movies table by ID.
UPDATE reviews SET movie_id = (
    SELECT movies.id FROM movies WHERE movies.title = reviews.title
    );

-- Re-do reviews table to add foreign key constraint to movie_id and drop title
-- column.
CREATE TABLE reviews2 (
    id INTEGER PRIMARY KEY,
    movie_id INTEGER,
    review_owner TEXT,
    rating INTEGER,
    blurb TEXT,
    watched_date TEXT,
    created_time TEXT,
    last_modified_time TEXT,
    FOREIGN KEY(movie_id) REFERENCES movies(id)
);

INSERT INTO reviews2
SELECT
    id,
    movie_id,
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
