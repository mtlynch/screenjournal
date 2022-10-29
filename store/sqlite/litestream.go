package sqlite

import "log"

func (d db) optimizeForLitestream() {
	if _, err := d.ctx.Exec(`
		-- Apply Litestream recommendations: https://litestream.io/tips/
		PRAGMA busy_timeout = 5000;
		PRAGMA synchronous = NORMAL;
		PRAGMA wal_autocheckpoint = 0;
			`); err != nil {
		log.Fatalf("failed to set Litestream compatibility pragmas: %v", err)
	}
}
