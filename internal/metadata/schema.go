package metadata

import (
	"database/sql"
	"fmt"
)

// initSchema creates the version_records table and index if they do not exist.
// This is the exact DDL from the design document.
func initSchema(db *sql.DB) error {
	// Create version_records table
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS version_records (
    tag           TEXT    PRIMARY KEY,
    content_hash  TEXT    NOT NULL,
    storage_key   TEXT    NOT NULL,
    accuracy      REAL,
    size_bytes    INTEGER NOT NULL,
    created_at    TEXT    NOT NULL
);`)
	if err != nil {
		return fmt.Errorf("failed to create version_records table: %w", err)
	}

	// Create index on content_hash for dedup checks
	_, err = db.Exec(`
CREATE INDEX IF NOT EXISTS idx_version_records_hash
    ON version_records (content_hash);`)
	if err != nil {
		return fmt.Errorf("failed to create content_hash index: %w", err)
	}

	return nil
}
