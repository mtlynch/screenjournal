
ALTER TABLE notification_preferences ADD COLUMN all_new_comments INTEGER;
ALTER TABLE notification_preferences ADD COLUMN comments_on_my_reviews INTEGER;

-- Subscribe everyone to new review notifications by default.
UPDATE notification_preferences SET all_new_comments = 1;
UPDATE notification_preferences SET comments_on_my_reviews = 1;
