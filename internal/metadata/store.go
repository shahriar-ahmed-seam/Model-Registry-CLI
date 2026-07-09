package metadata

import (
	"time"
)

// VersionRecord represents a row in the Metadata_Store representing one versioned model,
// associating a tag with a Content_Hash, accuracy, storage key, and timestamps.
type VersionRecord struct {
	Tag         string    // user-supplied label (e.g., "v2") that identifies this version
	ContentHash string    // lowercase hexadecimal SHA256 digest computed over the entire byte content
	StorageKey  string    // key under which the blob is stored (equals ContentHash for MVP)
	Accuracy    *float64  // optional accuracy value; nil when not supplied
	SizeBytes   int64     // size of the model blob in bytes
	CreatedAt   time.Time // creation timestamp (RFC3339 UTC)
}

// MetadataStore defines the interface for storing and retrieving version records.
// The SQLite implementation will satisfy this interface.
type MetadataStore interface {
	// InitSchema creates the version_records table and indices if they do not exist.
	// Called once during registry initialization.
	InitSchema() error

	// Insert stores a new VersionRecord.
	// Returns an error equivalent to registry.ErrTagExists if a record with the same Tag already exists.
	Insert(rec VersionRecord) error

	// GetByTag retrieves the VersionRecord identified by the given tag.
	// Returns an error equivalent to registry.ErrTagNotFound if no such record exists.
	GetByTag(tag string) (VersionRecord, error)

	// List returns all VersionRecords ordered by creation timestamp (oldest first).
	List() ([]VersionRecord, error)

	// Close releases any resources held by the store (e.g., database connections).
	Close() error
}
