CREATE TABLE review_reactions (
    id INTEGER PRIMARY KEY,
    review_id INTEGER NOT NULL,
    reaction_owner TEXT NOT NULL,
    emoji TEXT NOT NULL,
    created_time TEXT NOT NULL,
    FOREIGN KEY (review_id) REFERENCES reviews (id),
    FOREIGN KEY (reaction_owner) REFERENCES users (username),
    UNIQUE (review_id, reaction_owner, emoji)
);
