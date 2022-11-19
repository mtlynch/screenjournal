CREATE TABLE movies (
    id INTEGER PRIMARY KEY,
    tmdb_id INTEGER,
    imdb_id TEXT,
    title TEXT,
    release_date TEXT, -- Unpopulated for now
    poster_path TEXT
);

ALTER TABLE reviews
ADD COLUMN movie_id
INTEGER; -- TODO: Add foreign key?

ALTER TABLE reviews
DROP title;
