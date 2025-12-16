package sqlite

import (
	"database/sql"
	"log"
	"time"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
	"github.com/mtlynch/screenjournal/v2/store"
)

func (s Store) ReadReactions(rid screenjournal.ReviewID) ([]screenjournal.ReviewReaction, error) {
	review, err := s.ReadReview(rid)
	if err != nil {
		return []screenjournal.ReviewReaction{}, err
	}

	rows, err := s.ctx.Query(`
	SELECT
		id,
		review_id,
		reaction_owner,
		emoji,
		created_time
	FROM
		review_reactions
	WHERE
		review_id = :review_id
	ORDER BY
		created_time ASC
	`, sql.Named("review_id", rid))
	if err != nil {
		return []screenjournal.ReviewReaction{}, err
	}

	reactions := []screenjournal.ReviewReaction{}
	for rows.Next() {
		rr, err := reviewReactionFromRow(rows)
		if err != nil {
			return []screenjournal.ReviewReaction{}, err
		}

		rr.Review = review

		reactions = append(reactions, rr)
	}

	return reactions, nil
}

func (s Store) ReadReaction(id screenjournal.ReactionID) (screenjournal.ReviewReaction, error) {
	row := s.ctx.QueryRow(`
	SELECT
		id,
		review_id,
		reaction_owner,
		emoji,
		created_time
	FROM
		review_reactions
	WHERE
		id = :id
	`, sql.Named("id", id))

	return reviewReactionFromRow(row)
}

func (s Store) InsertReaction(rr screenjournal.ReviewReaction) (screenjournal.ReactionID, error) {
	log.Printf("inserting new reaction from %v on review ID %d", rr.Owner, rr.Review.ID.UInt64())

	now := time.Now()

	res, err := s.ctx.Exec(`
	INSERT INTO
		review_reactions
	(
		review_id,
		reaction_owner,
		emoji,
		created_time
	)
	VALUES (
		:review_id, :reaction_owner, :emoji, :created_time
	)
	`,
		sql.Named("review_id", rr.Review.ID),
		sql.Named("reaction_owner", rr.Owner),
		sql.Named("emoji", rr.Emoji.String()),
		sql.Named("created_time", formatTime(now)))
	if err != nil {
		return screenjournal.ReactionID(0), err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return screenjournal.ReactionID(0), err
	}

	return screenjournal.ReactionID(lastID), nil
}

func (s Store) DeleteReaction(id screenjournal.ReactionID) error {
	log.Printf("deleting reaction ID=%v", id)
	_, err := s.ctx.Exec(`DELETE FROM review_reactions WHERE id = :id`, sql.Named("id", id.String()))
	if err != nil {
		return err
	}
	return nil
}

func reviewReactionFromRow(row rowScanner) (screenjournal.ReviewReaction, error) {
	var id int
	var reviewId int
	var owner string
	var emoji string
	var createdTimeRaw string

	err := row.Scan(&id, &reviewId, &owner, &emoji, &createdTimeRaw)
	if err == sql.ErrNoRows {
		return screenjournal.ReviewReaction{}, store.ErrReactionNotFound
	} else if err != nil {
		return screenjournal.ReviewReaction{}, err
	}

	ct, err := parseDatetime(createdTimeRaw)
	if err != nil {
		return screenjournal.ReviewReaction{}, err
	}

	return screenjournal.ReviewReaction{
		ID: screenjournal.ReactionID(id),
		Review: screenjournal.Review{
			ID: screenjournal.ReviewID(reviewId),
		},
		Owner:   screenjournal.Username(owner),
		Emoji:   screenjournal.NewReactionEmoji(emoji),
		Created: ct,
	}, nil
}
