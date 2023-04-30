package sqlite

import (
	"database/sql"
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (db DB) ReadComments(rid screenjournal.ReviewID) ([]screenjournal.ReviewComment, error) {
	review, err := db.ReadReview(rid)
	if err != nil {
		return []screenjournal.ReviewComment{}, err
	}

	rows, err := db.ctx.Query(`
	SELECT
		id,
		comment_owner,
		comment_text,
		created_time,
		last_modified_time
	FROM
		review_comments
	WHERE
		review_id = ?
	ORDER BY
		created_time ASC
	`, rid)
	if err != nil {
		return []screenjournal.ReviewComment{}, err
	}

	comments := []screenjournal.ReviewComment{}
	for rows.Next() {
		rc, err := reviewCommentFromRow(rows)
		if err != nil {
			return []screenjournal.ReviewComment{}, err
		}

		rc.Review = review

		comments = append(comments, rc)
	}

	return comments, nil
}

func (db DB) InsertComment(rc screenjournal.ReviewComment) (screenjournal.CommentID, error) {
	log.Printf("inserting new comment from %v on %v's review of %s", rc.Owner, rc.Review.Owner, rc.Review.Movie.Title)

	now := time.Now()

	res, err := db.ctx.Exec(`
	INSERT INTO
		review_comments
	(
		review_id,
		comment_owner,
		comment_text,
		created_time,
		last_modified_time
	)
	VALUES (
		?, ?, ?, ?, ?
	)
	`,
		rc.Review.ID,
		rc.Owner,
		rc.CommentText,
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

func reviewCommentFromRow(row rowScanner) (screenjournal.ReviewComment, error) {
	var id int
	var owner string
	var comment string
	var createdTimeRaw string
	var lastModifiedTimeRaw string

	err := row.Scan(&id, &owner, &comment, &createdTimeRaw, &lastModifiedTimeRaw)
	if err == sql.ErrNoRows {
		return screenjournal.ReviewComment{}, store.ErrCommentNotFound
	} else if err != nil {
		return screenjournal.ReviewComment{}, err
	}

	ct, err := parseDatetime(createdTimeRaw)
	if err != nil {
		return screenjournal.ReviewComment{}, err
	}

	lmt, err := parseDatetime(lastModifiedTimeRaw)
	if err != nil {
		return screenjournal.ReviewComment{}, err
	}

	return screenjournal.ReviewComment{
		ID:          screenjournal.CommentID(id),
		Owner:       screenjournal.Username(owner),
		CommentText: screenjournal.CommentText(comment),
		Created:     ct,
		Modified:    lmt,
	}, nil
}
