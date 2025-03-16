-- Make the rating column nullable to support reviews without ratings
PRAGMA foreign_keys = OFF;

-- Create a new reviews table with the rating column as nullable
CREATE TABLE reviews_new (
    id INTEGER PRIMARY KEY,
    review_owner TEXT NOT NULL,
    movie_id INTEGER,
    tv_show_id INTEGER,
    tv_show_season INTEGER,
    rating INTEGER,  -- Now nullable
    blurb TEXT NOT NULL,
    watched_date TEXT NOT NULL,
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL,
    FOREIGN KEY (review_owner) REFERENCES users (username),
    FOREIGN KEY (movie_id) REFERENCES movies (id),
    FOREIGN KEY (tv_show_id) REFERENCES tv_shows (id),
    CHECK (
        (movie_id IS NULL AND tv_show_id IS NOT NULL)
        OR (movie_id IS NOT NULL AND tv_show_id IS NULL)
    )
);

-- Create a new review_comments table that references the new reviews table
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
    FOREIGN KEY (comment_owner) REFERENCES users (username)
) STRICT;

-- Copy all data from the old reviews table to the new table
INSERT INTO reviews_new
SELECT
    id,
    review_owner,
    movie_id,
    tv_show_id,
    tv_show_season,
    rating,
    blurb,
    watched_date,
    created_time,
    last_modified_time
FROM reviews;

-- Copy all data from the old review_comments table to the new table
INSERT INTO review_comments_new
SELECT
    id,
    review_id,
    comment_owner,
    comment_text,
    created_time,
    last_modified_time
FROM review_comments;

-- Drop the old tables (order matters - drop child tables first)
DROP TABLE review_comments;
DROP TABLE reviews;

-- Rename the new tables to the original names
ALTER TABLE reviews_new RENAME TO reviews;
ALTER TABLE review_comments_new RENAME TO review_comments;

-- Create indexes for the new tables
CREATE INDEX idx_reviews_owner ON reviews (review_owner);
CREATE INDEX idx_reviews_movie_id ON reviews (movie_id);
CREATE INDEX idx_reviews_tv_show_id ON reviews (tv_show_id);
CREATE INDEX idx_review_comments_review_id ON review_comments (review_id);

PRAGMA foreign_keys = ON;
