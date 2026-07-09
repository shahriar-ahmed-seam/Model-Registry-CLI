package blob

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestInMemoryStore_BasicOperations(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Test that a non-existent blob returns false
	exists, err := store.Exists(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Expected non-existent blob to return false")
	}

	// Put a blob
	content := []byte("test content")
	err = store.Put(ctx, "test-key", bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify it exists
	exists, err = store.Exists(ctx, "test-key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected blob to exist after Put")
	}

	// Get the blob
	var buf bytes.Buffer
	err = store.Get(ctx, "test-key", &buf)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), content) {
		t.Errorf("Get returned wrong content: got %q, want %q", buf.String(), string(content))
	}
}

func TestInMemoryStore_NotFoundError(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	var buf bytes.Buffer
	err := store.Get(ctx, "nonexistent", &buf)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestInMemoryStore_InjectUploadFailure(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Inject a failure for a specific key
	injectedErr := errors.New("injected upload failure")
	store.InjectUploadFailure("fail-key", injectedErr)

	// Try to put to the failing key
	content := []byte("test content")
	err := store.Put(ctx, "fail-key", bytes.NewReader(content), int64(len(content)))
	if !errors.Is(err, injectedErr) {
		t.Errorf("Expected injected error, got: %v", err)
	}

	// Verify the blob was not stored
	exists, err := store.Exists(ctx, "fail-key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Blob should not exist after failed upload")
	}

	// Clear failures and try again
	store.ClearUploadFailures()
	err = store.Put(ctx, "fail-key", bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("Put should succeed after clearing failures: %v", err)
	}
}

func TestInMemoryStore_InjectCorruption(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Store a blob
	original := []byte("original content")
	err := store.Put(ctx, "corrupt-key", bytes.NewReader(original), int64(len(original)))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Inject a corruption that replaces all 'o' with 'x'
	store.InjectCorruption("corrupt-key", func(data []byte) []byte {
		return []byte(strings.ReplaceAll(string(data), "o", "x"))
	})

	// Get the corrupted blob
	var buf bytes.Buffer
	err = store.Get(ctx, "corrupt-key", &buf)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	corrupted := buf.Bytes()
	expected := []byte("xriginal cxntent")
	if !bytes.Equal(corrupted, expected) {
		t.Errorf("Corruption not applied: got %q, want %q", string(corrupted), string(expected))
	}

	// Clear corruptions and get original content
	store.ClearCorruptions()
	buf.Reset()
	err = store.Get(ctx, "corrupt-key", &buf)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), original) {
		t.Error("Should get original content after clearing corruptions")
	}
}

func TestInMemoryStore_GetContentHelper(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	content := []byte("helper test")
	err := store.Put(ctx, "helper-key", bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Use the test helper
	retrieved, ok := store.GetContent("helper-key")
	if !ok {
		t.Error("GetContent should find the blob")
	}
	if !bytes.Equal(retrieved, content) {
		t.Errorf("GetContent returned wrong content: got %q, want %q", string(retrieved), string(content))
	}

	// Test non-existent key
	_, ok = store.GetContent("nonexistent")
	if ok {
		t.Error("GetContent should return false for non-existent key")
	}
}

func TestInMemoryStore_ListKeys(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Add some blobs
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		content := []byte("content for " + key)
		err := store.Put(ctx, key, bytes.NewReader(content), int64(len(content)))
		if err != nil {
			t.Fatalf("Put failed for %s: %v", key, err)
		}
	}

	// List keys
	list := store.ListKeys()
	if len(list) != len(keys) {
		t.Errorf("ListKeys returned wrong number of keys: got %d, want %d", len(list), len(keys))
	}

	// Check that all expected keys are present
	keySet := make(map[string]bool)
	for _, key := range list {
		keySet[key] = true
	}
	for _, key := range keys {
		if !keySet[key] {
			t.Errorf("Missing key in ListKeys: %s", key)
		}
	}
}

func TestInMemoryStore_Clear(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Add a blob without injection first
	content := []byte("test")
	err := store.Put(ctx, "test-key", bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Now inject failure and corruption for future operations
	store.InjectUploadFailure("test-key", errors.New("failure"))
	store.InjectCorruption("test-key", func(data []byte) []byte { return data })

	// Verify it exists
	exists, _ := store.Exists(ctx, "test-key")
	if !exists {
		t.Fatal("Blob should exist before clear")
	}

	// Clear everything
	store.Clear()

	// Verify blob is gone
	exists, _ = store.Exists(ctx, "test-key")
	if exists {
		t.Error("Blob should not exist after clear")
	}

	// Verify upload failure is cleared (should be able to put now)
	err = store.Put(ctx, "test-key", bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Errorf("Put should succeed after clear: %v", err)
	}
}
