CREATE TABLE IF NOT EXISTS auth_sessions (
    session_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    created_at TEXT NOT NULL CHECK (datetime(created_at) IS NOT NULL),
    expires_at TEXT NOT NULL CHECK (datetime(expires_at) IS NOT NULL),
    FOREIGN KEY (user_id) REFERENCES users (username)
) STRICT;

CREATE INDEX idx_auth_sessions_expires_at ON auth_sessions (expires_at);
