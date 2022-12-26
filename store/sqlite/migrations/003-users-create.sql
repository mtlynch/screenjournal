CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    is_admin INTEGER,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL
);

-- Re-do reviews table to add foreign key constraint to review_owner column.
CREATE TABLE reviews2 (
    id INTEGER PRIMARY KEY,
    movie_id INTEGER,
    review_owner TEXT,
    rating INTEGER,
    blurb TEXT,
    watched_date TEXT,
    created_time TEXT,
    last_modified_time TEXT,
    FOREIGN KEY(movie_id) REFERENCES movies(id),
    FOREIGN KEY(review_owner) REFERENCES users(username)
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
