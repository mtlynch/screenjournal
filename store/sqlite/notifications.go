package sqlite

import (
	"database/sql"

	"github.com/mtlynch/screenjournal/v2"
)

func (db DB) ReadNotificationSubscribers() ([]screenjournal.User, error) {
	rows, err := db.ctx.Query(`
	SELECT
		users.username AS username,
		users.is_admin AS is_admin,
		users.email AS email,
		users.password_hash AS password_hash
	FROM
		users, notification_preferences
	WHERE
		users.username = notification_preferences.username AND
		notification_preferences.new_reviews = 1`)
	if err != nil {
		if err == sql.ErrNoRows {
			return []screenjournal.User{}, nil
		}
		return []screenjournal.User{}, err
	}

	users := []screenjournal.User{}
	for rows.Next() {
		user, err := userFromRow(rows)
		if err != nil {
			return []screenjournal.User{}, err
		}
		users = append(users, user)
	}
	return users, nil
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
