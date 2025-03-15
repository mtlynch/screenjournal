-- Make the rating column nullable to support reviews without ratings
PRAGMA foreign_keys=off;

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
    FOREIGN KEY (review_owner) REFERENCES users(username),
    FOREIGN KEY (movie_id) REFERENCES movies(id),
    FOREIGN KEY (tv_show_id) REFERENCES tv_shows(id),
    CHECK ((movie_id IS NULL AND tv_show_id IS NOT NULL) OR (movie_id IS NOT NULL AND tv_show_id IS NULL))
);

-- Copy all data from the old table to the new table
INSERT INTO reviews_new
SELECT id, review_owner, movie_id, tv_show_id, tv_show_season, rating, blurb, watched_date, created_time, last_modified_time
FROM reviews;

-- Drop the old table
DROP TABLE reviews;

-- Rename the new table to the original name
ALTER TABLE reviews_new RENAME TO reviews;

-- Create indexes for the new table
CREATE INDEX idx_reviews_owner ON reviews(review_owner);
CREATE INDEX idx_reviews_movie_id ON reviews(movie_id);
CREATE INDEX idx_reviews_tv_show_id ON reviews(tv_show_id);

PRAGMA foreign_keys=on;
