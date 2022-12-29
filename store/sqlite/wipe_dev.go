//go:build dev

package sqlite

import "log"

func (db DB) Clear() {
	log.Printf("clearing all SQLite tables")
	if _, err := db.ctx.Exec(`DELETE FROM movies`); err != nil {
		log.Fatalf("failed to delete movies: %v", err)
	}
	if _, err := db.ctx.Exec(`DELETE FROM reviews`); err != nil {
		log.Fatalf("failed to delete reviews: %v", err)
	}
	if _, err := db.ctx.Exec(`DELETE FROM users`); err != nil {
		log.Fatalf("failed to delete users: %v", err)
	}
	if _, err := db.ctx.Exec(`DELETE FROM invites`); err != nil {
		log.Fatalf("failed to delete invites: %v", err)
	}
}
