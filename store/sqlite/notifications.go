package sqlite

import (
	"database/sql"
	"log"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
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

func (db DB) ReadCommentSubscribers() ([]screenjournal.EmailSubscriber, error) {
	rows, err := db.ctx.Query(`
	SELECT
		users.username AS username,
		users.email AS email
	FROM
		users, notification_preferences
	WHERE
		users.username = notification_preferences.username AND
		notification_preferences.all_new_comments = 1`)
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
	var allNewComments bool
	err := db.ctx.QueryRow(`
	SELECT
		new_reviews,
		all_new_comments
	FROM
		notification_preferences
	WHERE
		username = ?`, username.String()).Scan(&newReviews, &allNewComments)
	if err != nil {
		return screenjournal.NotificationPreferences{}, err
	}

	return screenjournal.NotificationPreferences{
		NewReviews:     newReviews,
		AllNewComments: allNewComments,
	}, nil
}

func (db DB) UpdateNotificationPreferences(username screenjournal.Username, prefs screenjournal.NotificationPreferences) error {
	log.Printf("updating notifications preferences for %s: newReviews=%v, allNewComments=%v", username, prefs.NewReviews, prefs.AllNewComments)
	if _, err := db.ctx.Exec(`
	UPDATE notification_preferences
	SET
		new_reviews = ?,
		all_new_comments = ?
	WHERE
		username = ?`, prefs.NewReviews, prefs.AllNewComments, username); err != nil {
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
