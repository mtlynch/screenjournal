package sqlite

import (
	"database/sql"
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (s Store) ReadComments(rid screenjournal.ReviewID) ([]screenjournal.ReviewComment, error) {
	review, err := s.ReadReview(rid)
	if err != nil {
		return []screenjournal.ReviewComment{}, err
	}

	rows, err := s.ctx.Query(`
		SELECT
			id,
			review_id,
			comment_owner,
			comment_text,
			created_time,
			last_modified_time
		FROM
			review_comments
		WHERE
			review_id = :review_id
		ORDER BY
			created_time ASC
		`, sql.Named("review_id", rid))
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

func (s Store) ReadComment(cid screenjournal.CommentID) (screenjournal.ReviewComment, error) {
	row := s.ctx.QueryRow(`
		SELECT
			id,
			review_id,
			comment_owner,
			comment_text,
			created_time,
			last_modified_time
		FROM
			review_comments
		WHERE
			id = :id
		`, sql.Named("id", cid))

	return reviewCommentFromRow(row)
}

func (s Store) InsertComment(rc screenjournal.ReviewComment) (screenjournal.CommentID, error) {
	log.Printf("inserting new comment from %v on %v's review of %s", rc.Owner, rc.Review.Owner, rc.Review.Movie.Title)

	now := time.Now()

	res, err := s.ctx.Exec(`
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
        :review_id, :comment_owner, :comment_text, :created_time, :last_modified_time
    )
    `,
		sql.Named("review_id", rc.Review.ID),
		sql.Named("comment_owner", rc.Owner),
		sql.Named("comment_text", rc.CommentText),
		sql.Named("created_time", formatTime(now)),
		sql.Named("last_modified_time", formatTime(now)))
	if err != nil {
		return screenjournal.CommentID(0), err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return screenjournal.CommentID(0), err
	}

	return screenjournal.CommentID(lastID), nil
}

func (s Store) UpdateComment(rc screenjournal.ReviewComment) error {
	log.Printf("updating comment %v from %v", rc.ID, rc.Owner)

	_, err := s.ctx.Exec(`
        UPDATE review_comments
        SET
            comment_text = :comment_text,
            last_modified_time = :last_modified_time
        WHERE
            id = :id;
        `,
		sql.Named("comment_text", rc.CommentText.String()),
		sql.Named("last_modified_time", formatTime(time.Now())),
		sql.Named("id", rc.ID.UInt64()))
	if err != nil {
		return err
	}

	return nil
}

func (s Store) DeleteComment(cid screenjournal.CommentID) error {
	log.Printf("deleting comment ID=%v", cid)
	_, err := s.ctx.Exec(`DELETE FROM review_comments WHERE id = :id`, sql.Named("id", cid.String()))
	if err != nil {
		return err
	}
	return nil
}

func reviewCommentFromRow(row rowScanner) (screenjournal.ReviewComment, error) {
	var id int
	var reviewId int
	var owner string
	var comment string
	var createdTimeRaw string
	var lastModifiedTimeRaw string

	err := row.Scan(&id, &reviewId, &owner, &comment, &createdTimeRaw, &lastModifiedTimeRaw)
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
		ID: screenjournal.CommentID(id),
		Review: screenjournal.Review{
			ID: screenjournal.ReviewID(reviewId),
		},
		Owner:       screenjournal.Username(owner),
		CommentText: screenjournal.CommentText(comment),
		Created:     ct,
		Modified:    lmt,
	}, nil
}
