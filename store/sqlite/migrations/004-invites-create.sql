CREATE TABLE invites (
    id INTEGER PRIMARY KEY,
    invitee TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    created_time TEXT NOT NULL
);
