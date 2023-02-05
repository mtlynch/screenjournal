package sqlite

import (
	"github.com/mtlynch/screenjournal/v2"
)

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
