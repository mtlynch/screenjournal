CREATE TABLE password_reset_tokens (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL,
    FOREIGN KEY (username) REFERENCES users (username)
);
