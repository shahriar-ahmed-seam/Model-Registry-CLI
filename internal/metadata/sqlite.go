package metadata

import (
	"database/sql"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite" // CGo-free SQLite driver

	"ml-reg/internal/errors"
)

// SQLiteStore implements the MetadataStore interface using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens or creates a SQLite database at the given path.
func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database at %s: %w", path, err)
	}

	// Enable foreign keys and set busy timeout
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	_, err = db.Exec("PRAGMA busy_timeout = 5000;")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.InitSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// InitSchema creates the version_records table and indices if they do not exist.
func (s *SQLiteStore) InitSchema() error {
	return initSchema(s.db)
}

// Insert stores a new VersionRecord.
func (s *SQLiteStore) Insert(rec VersionRecord) error {
	// Convert accuracy pointer to nullable SQL value
	var accuracy sql.NullFloat64
	if rec.Accuracy != nil {
		accuracy = sql.NullFloat64{Float64: *rec.Accuracy, Valid: true}
	}

	// Convert time to RFC3339 string for SQLite storage
	createdAt := rec.CreatedAt.Format(time.RFC3339)

	_, err := s.db.Exec(`
		INSERT INTO version_records (tag, content_hash, storage_key, accuracy, size_bytes, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		rec.Tag, rec.ContentHash, rec.StorageKey, accuracy, rec.SizeBytes, createdAt)

	// Map primary key conflict to ErrTagExists
	if err != nil {
		// Check if this is a unique constraint violation
		// Different SQLite drivers may return different error messages
		errStr := err.Error()
		if strings.Contains(errStr, "UNIQUE constraint failed") && strings.Contains(errStr, "tag") ||
			strings.Contains(errStr, "constraint failed") && strings.Contains(errStr, "tag") ||
			strings.Contains(errStr, "duplicate") && strings.Contains(errStr, "tag") {
			return errors.ErrTagExists
		}
		return fmt.Errorf("failed to insert version record: %w", err)
	}

	return nil
}

// GetByTag retrieves the VersionRecord identified by the given tag.
func (s *SQLiteStore) GetByTag(tag string) (VersionRecord, error) {
	var rec VersionRecord
	var accuracy sql.NullFloat64
	var createdAtStr string

	err := s.db.QueryRow(`
		SELECT tag, content_hash, storage_key, accuracy, size_bytes, created_at
		FROM version_records
		WHERE tag = ?`,
		tag).Scan(&rec.Tag, &rec.ContentHash, &rec.StorageKey, &accuracy, &rec.SizeBytes, &createdAtStr)

	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return VersionRecord{}, errors.ErrTagNotFound
		}
		return VersionRecord{}, fmt.Errorf("failed to get version record: %w", err)
	}

	// Convert nullable accuracy
	if accuracy.Valid {
		rec.Accuracy = &accuracy.Float64
	}

	// Parse RFC3339 timestamp
	rec.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return VersionRecord{}, fmt.Errorf("failed to parse created_at timestamp: %w", err)
	}

	return rec, nil
}

// List returns all VersionRecords ordered by creation timestamp (oldest first).
func (s *SQLiteStore) List() ([]VersionRecord, error) {
	rows, err := s.db.Query(`
		SELECT tag, content_hash, storage_key, accuracy, size_bytes, created_at
		FROM version_records
		ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("failed to query version records: %w", err)
	}
	defer rows.Close()

	var records []VersionRecord
	for rows.Next() {
		var rec VersionRecord
		var accuracy sql.NullFloat64
		var createdAtStr string

		err := rows.Scan(&rec.Tag, &rec.ContentHash, &rec.StorageKey, &accuracy, &rec.SizeBytes, &createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version record: %w", err)
		}

		// Convert nullable accuracy
		if accuracy.Valid {
			rec.Accuracy = &accuracy.Float64
		}

		// Parse RFC3339 timestamp
		rec.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at timestamp: %w", err)
		}

		records = append(records, rec)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating version records: %w", err)
	}

	return records, nil
}

// Close releases the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
