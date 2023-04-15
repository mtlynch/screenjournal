CREATE TABLE review_comments (
    id INTEGER PRIMARY KEY,
    review_id INTEGER,
    comment_owner TEXT,
    comment TEXT,
    created_time TEXT,
    last_modified_time TEXT,
    FOREIGN KEY(review_id) REFERENCES reviews(id),
    FOREIGN KEY(comment_owner) REFERENCES users(username)
);
