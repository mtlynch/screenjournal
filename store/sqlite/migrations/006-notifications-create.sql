CREATE TABLE notification_preferences (
    username TEXT PRIMARY KEY,
    new_reviews INTEGER,
    FOREIGN KEY(username) REFERENCES users(username)
);

INSERT INTO notification_preferences
SELECT
    username,
    1
FROM
    users;
