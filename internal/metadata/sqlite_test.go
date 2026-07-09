package metadata

import (
	"os"
	"testing"
	"time"

	"ml-reg/internal/errors"
)

func TestSQLiteStore(t *testing.T) {
	// Create a temporary database file
	tmpfile, err := os.CreateTemp("", "testdb-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// Create store
	store, err := NewSQLiteStore(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to create SQLite store: %v", err)
	}
	defer store.Close()

	// Test InitSchema (already called by NewSQLiteStore)
	// We'll just verify we can query the table structure
	err = store.InitSchema()
	if err != nil {
		t.Errorf("InitSchema failed: %v", err)
	}

	// Test Insert
	accuracy := 0.95
	rec := VersionRecord{
		Tag:         "test-tag",
		ContentHash: "abc123",
		StorageKey:  "abc123",
		Accuracy:    &accuracy,
		SizeBytes:   1024,
		CreatedAt:   time.Now().UTC(),
	}

	err = store.Insert(rec)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}

	// Test GetByTag
	fetched, err := store.GetByTag("test-tag")
	if err != nil {
		t.Errorf("GetByTag failed: %v", err)
	}

	if fetched.Tag != rec.Tag {
		t.Errorf("GetByTag tag mismatch: got %s, want %s", fetched.Tag, rec.Tag)
	}
	if fetched.ContentHash != rec.ContentHash {
		t.Errorf("GetByTag content hash mismatch: got %s, want %s", fetched.ContentHash, rec.ContentHash)
	}
	if *fetched.Accuracy != *rec.Accuracy {
		t.Errorf("GetByTag accuracy mismatch: got %f, want %f", *fetched.Accuracy, *rec.Accuracy)
	}
	if fetched.SizeBytes != rec.SizeBytes {
		t.Errorf("GetByTag size mismatch: got %d, want %d", fetched.SizeBytes, rec.SizeBytes)
	}

	// Test duplicate tag insert
	err = store.Insert(rec)
	if err == nil {
		t.Error("Expected error when inserting duplicate tag")
	}
	if !isTagExistsError(err) {
		t.Errorf("Expected ErrTagExists, got: %v", err)
	}

	// Test GetByTag non-existent
	_, err = store.GetByTag("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent tag")
	}
	if !isTagNotFoundError(err) {
		t.Errorf("Expected ErrTagNotFound, got: %v", err)
	}

	// Test List
	records, err := store.List()
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}
}

func isTagExistsError(err error) bool {
	// Check if error is ErrTagExists or contains tag exists message
	return err != nil && (err.Error() == errors.ErrTagExists.Error() ||
		err.Error() == "tag is already in use")
}

func isTagNotFoundError(err error) bool {
	// Check if error is ErrTagNotFound or contains tag not found message
	return err != nil && (err.Error() == errors.ErrTagNotFound.Error() ||
		err.Error() == "tag was not found")
}
