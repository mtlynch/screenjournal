CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    is_admin INTEGER,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL
);
