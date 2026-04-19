CREATE TABLE IF NOT EXISTS auth_sessions (
    session_id TEXT PRIMARY KEY,
    session_data BLOB,
    expires_at TEXT NOT NULL
);
