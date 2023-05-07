package sqlite

import (
	"database/sql"

	"github.com/mtlynch/screenjournal/v2"
)

func (db DB) ReadReviewSubscribers() ([]screenjournal.EmailSubscriber, error) {
	rows, err := db.ctx.Query(`
	SELECT
		users.username AS username,
		users.email AS email
	FROM
		users, notification_preferences
	WHERE
		users.username = notification_preferences.username AND
		notification_preferences.new_reviews = 1`)
	if err != nil {
		if err == sql.ErrNoRows {
			return []screenjournal.EmailSubscriber{}, nil
		}
		return []screenjournal.EmailSubscriber{}, err
	}

	subscribers := []screenjournal.EmailSubscriber{}
	for rows.Next() {
		subscriber, err := emailSubscriberFromRow(rows)
		if err != nil {
			return []screenjournal.EmailSubscriber{}, err
		}
		subscribers = append(subscribers, subscriber)
	}
	return subscribers, nil
}

func (db DB) ReadNotificationPreferences(username screenjournal.Username) (screenjournal.NotificationPreferences, error) {
	var newReviews bool
	err := db.ctx.QueryRow(`
	SELECT
		new_reviews
	FROM
		notification_preferences
	WHERE
		username = ?`, username.String()).Scan(&newReviews)
	if err != nil {
		return screenjournal.NotificationPreferences{}, err
	}

	return screenjournal.NotificationPreferences{
		NewReviews: newReviews,
	}, nil
}

func (db DB) UpdateNotificationPreferences(username screenjournal.Username, prefs screenjournal.NotificationPreferences) error {
	if _, err := db.ctx.Exec(`
	UPDATE notification_preferences
	SET
		new_reviews = ?
	WHERE
		username = ?`, prefs.NewReviews, username); err != nil {
		return err
	}

	return nil
}

func emailSubscriberFromRow(row rowScanner) (screenjournal.EmailSubscriber, error) {
	var username string
	var email string
	if err := row.Scan(&username, &email); err != nil {
		return screenjournal.EmailSubscriber{}, err
	}
	return screenjournal.EmailSubscriber{
		Username: screenjournal.Username(username),
		Email:    screenjournal.Email(email),
	}, nil
}
