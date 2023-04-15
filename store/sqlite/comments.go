package sqlite

import (
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2"
)

func (db DB) InsertComment(rc screenjournal.ReviewComment) (screenjournal.CommentID, error) {
	log.Printf("inserting new comment from %v on %v's of %s", rc.Owner, rc.Review.Owner, rc.Review.Movie.Title)

	now := time.Now()

	res, err := db.ctx.Exec(`
	INSERT INTO
		review_comments
	(
		review_id,
		comment_owner,
		comment,
		created_time,
		last_modified_time
	)
	VALUES (
		?, ?, ?, ?, ?
	)
	`,
		rc.Review.ID,
		rc.Owner,
		rc.Comment,
		formatTime(now),
		formatTime(now))
	if err != nil {
		return screenjournal.CommentID(0), err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return screenjournal.CommentID(0), err
	}

	return screenjournal.CommentID(lastID), nil
}
