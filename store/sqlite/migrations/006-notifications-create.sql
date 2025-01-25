CREATE TABLE notification_preferences (
    username TEXT PRIMARY KEY,
    new_reviews INTEGER,
    FOREIGN KEY (username) REFERENCES users (username)
);

INSERT INTO notification_preferences (
    username
)
SELECT username
FROM
    users;

-- Subscribe everyone to new review notifications by default.
UPDATE notification_preferences
SET new_reviews = 1;
