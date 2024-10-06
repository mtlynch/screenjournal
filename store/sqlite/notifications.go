package sqlite

import (
	"database/sql"
	"log"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func (s Store) ReadReviewSubscribers() ([]screenjournal.EmailSubscriber, error) {
	rows, err := s.ctx.Query(`
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

func (s Store) ReadCommentSubscribers() ([]screenjournal.EmailSubscriber, error) {
	rows, err := s.ctx.Query(`
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

func (s Store) ReadNotificationPreferences(username screenjournal.Username) (screenjournal.NotificationPreferences, error) {
	var newReviews bool
	var allNewComments bool
	err := s.ctx.QueryRow(`
	SELECT
		new_reviews,
		all_new_comments
	FROM
		notification_preferences
	WHERE
		username = :username`, sql.Named("username", username.String())).Scan(&newReviews, &allNewComments)
	if err != nil {
		return screenjournal.NotificationPreferences{}, err
	}

	return screenjournal.NotificationPreferences{
		NewReviews:     newReviews,
		AllNewComments: allNewComments,
	}, nil
}

func (s Store) UpdateNotificationPreferences(username screenjournal.Username, prefs screenjournal.NotificationPreferences) error {
	log.Printf("updating notifications preferences for %s: newReviews=%v, allNewComments=%v", username, prefs.NewReviews, prefs.AllNewComments)
	if _, err := s.ctx.Exec(`
	UPDATE notification_preferences
	SET
		new_reviews = :new_reviews,
		all_new_comments = :all_new_comments
	WHERE
		username = :username`,
		sql.Named("new_reviews", prefs.NewReviews),
		sql.Named("all_new_comments", prefs.AllNewComments),
		sql.Named("username", username)); err != nil {
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
