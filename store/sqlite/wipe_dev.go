//go:build dev

package sqlite

import "log"

func (db DB) Clear() {
	log.Printf("clearing all SQLite tables")
	if _, err := db.ctx.Exec(`DELETE FROM reviews`); err != nil {
		log.Fatalf("failed to delete rows: %v", err)
	}
}
