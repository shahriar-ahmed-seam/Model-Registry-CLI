// Package blob provides the BlobStore interface for storing and retrieving
// model blobs, with an in-memory fake implementation for testing.
package blob

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
)

// BlobStore defines the interface for storing and retrieving model blobs.
// Implementations are expected to support streaming of large files.
type BlobStore interface {
	// Exists reports whether an object with the given content-hash key is present.
	Exists(ctx context.Context, key string) (bool, error)
	// Put streams content to the store under key. Size may be -1 if unknown.
	Put(ctx context.Context, key string, r io.Reader, size int64) error
	// Get streams the object content for key to w.
	Get(ctx context.Context, key string, w io.Writer) error
}

// ErrNotFound is returned when a blob does not exist.
var ErrNotFound = errors.New("blob not found")

// InMemoryStore is an in-memory implementation of BlobStore for testing.
// It supports injectable upload failures and blob-byte corruption for property tests.
type InMemoryStore struct {
	// store maps blob keys to their content.
	store map[string][]byte
	// uploadFailures maps keys that should fail on Put.
	uploadFailures map[string]error
	// corruptions maps keys to corruption functions that modify content on Get.
	corruptions map[string]func([]byte) []byte
}

// NewInMemoryStore creates a new empty in-memory blob store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		store:          make(map[string][]byte),
		uploadFailures: make(map[string]error),
		corruptions:    make(map[string]func([]byte) []byte),
	}
}

// Exists reports whether an object with the given key exists in the store.
func (s *InMemoryStore) Exists(_ context.Context, key string) (bool, error) {
	_, ok := s.store[key]
	return ok, nil
}

// Put streams content to the store under the given key.
func (s *InMemoryStore) Put(_ context.Context, key string, r io.Reader, size int64) error {
	// Check if we should inject an upload failure for this key.
	if err, ok := s.uploadFailures[key]; ok {
		return err
	}

	// Read all bytes from the reader.
	// For testing purposes, we read the entire content into memory.
	// In production S3 implementation, this would stream directly.
	// We use a bytes.Buffer which grows efficiently for unknown size.
	var buf bytes.Buffer
	// If size is known, pre-allocate capacity for efficiency
	if size >= 0 && size < 1<<30 { // Cap at 1GB for safety
		buf.Grow(int(size))
	}
	
	// Copy with a reasonable buffer size
	_, err := io.CopyBuffer(&buf, r, make([]byte, 1024*1024)) // 1MB buffer
	if err != nil {
		return fmt.Errorf("failed to read blob content: %w", err)
	}
	
	data := buf.Bytes()

	// Store the blob content.
	s.store[key] = data
	return nil
}

// Get streams the object content for the given key to the writer.
func (s *InMemoryStore) Get(_ context.Context, key string, w io.Writer) error {
	data, ok := s.store[key]
	if !ok {
		return ErrNotFound
	}

	// Apply corruption if configured for this key.
	if corrupt, ok := s.corruptions[key]; ok {
		data = corrupt(data)
	}

	// Write the data to the writer.
	n, err := w.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write blob content: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("short write: wrote %d of %d bytes", n, len(data))
	}
	return nil
}

// InjectUploadFailure configures the store to return the given error
// when Put is called with the specified key.
func (s *InMemoryStore) InjectUploadFailure(key string, err error) {
	s.uploadFailures[key] = err
}

// ClearUploadFailures removes all configured upload failures.
func (s *InMemoryStore) ClearUploadFailures() {
	s.uploadFailures = make(map[string]error)
}

// InjectCorruption configures the store to apply the given corruption function
// when Get is called for the specified key.
func (s *InMemoryStore) InjectCorruption(key string, corrupt func([]byte) []byte) {
	s.corruptions[key] = corrupt
}

// ClearCorruptions removes all configured corruptions.
func (s *InMemoryStore) ClearCorruptions() {
	s.corruptions = make(map[string]func([]byte) []byte)
}

// Clear removes all blobs and clears all injected failures and corruptions.
func (s *InMemoryStore) Clear() {
	s.store = make(map[string][]byte)
	s.uploadFailures = make(map[string]error)
	s.corruptions = make(map[string]func([]byte) []byte)
}

// GetContent returns the raw content for a key, if it exists.
// This is a test helper method not part of the BlobStore interface.
func (s *InMemoryStore) GetContent(key string) ([]byte, bool) {
	data, ok := s.store[key]
	return data, ok
}

// ListKeys returns all keys in the store.
// This is a test helper method not part of the BlobStore interface.
func (s *InMemoryStore) ListKeys() []string {
	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}
	return keys
}
