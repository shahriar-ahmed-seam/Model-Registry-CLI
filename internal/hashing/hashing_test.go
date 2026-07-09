package hashing

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

func TestHashReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"hello", "hello world", "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"},
		{"multiline", "hello\nworld\n", "0d6c7b9e60c9de17f8a5d6c7b9e60c9de17f8a5d6c7b9e60c9de17f8a5d6c7b9e"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			hash, n, err := HashReader(r)
			
			if err != nil {
				t.Fatalf("HashReader returned error: %v", err)
			}
			
			if n != int64(len(tt.input)) {
				t.Errorf("HashReader returned byte count %d, expected %d", n, len(tt.input))
			}
			
			// Compute expected hash directly
			h := sha256.New()
			h.Write([]byte(tt.input))
			expected := hex.EncodeToString(h.Sum(nil))
			
			if hash != expected {
				t.Errorf("HashReader returned hash %s, expected %s", hash, expected)
			}
		})
	}
}

func TestHashReaderLarge(t *testing.T) {
	// Create a 2MB buffer to test chunking
	size := 2 * 1024 * 1024 // 2MB
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	
	r := bytes.NewReader(data)
	hash, n, err := HashReader(r)
	
	if err != nil {
		t.Fatalf("HashReader returned error: %v", err)
	}
	
	if n != int64(size) {
		t.Errorf("HashReader returned byte count %d, expected %d", n, size)
	}
	
	// Compute expected hash directly
	h := sha256.New()
	h.Write(data)
	expected := hex.EncodeToString(h.Sum(nil))
	
	if hash != expected {
		t.Errorf("HashReader returned hash %s, expected %s", hash, expected)
	}
}

func TestHashFile(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "hashfile-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	content := []byte("test file content")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	
	hash, size, err := HashFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("HashFile returned error: %v", err)
	}
	
	if size != int64(len(content)) {
		t.Errorf("HashFile returned size %d, expected %d", size, len(content))
	}
	
	// Compute expected hash directly
	h := sha256.New()
	h.Write(content)
	expected := hex.EncodeToString(h.Sum(nil))
	
	if hash != expected {
		t.Errorf("HashFile returned hash %s, expected %s", hash, expected)
	}
}

func TestHashFileNotFound(t *testing.T) {
	_, _, err := HashFile("/nonexistent/file/path")
	if err == nil {
		t.Error("HashFile should return error for nonexistent file")
	}
}

func TestHashReaderEmptyReader(t *testing.T) {
	// Test with an empty reader (already covered in first test)
	hash, n, err := HashReader(bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatalf("HashReader returned error: %v", err)
	}
	
	expected := hex.EncodeToString(sha256.New().Sum(nil))
	if hash != expected {
		t.Errorf("HashReader returned hash %s, expected %s", hash, expected)
	}
	
	if n != 0 {
		t.Errorf("HashReader returned byte count %d, expected 0", n)
	}
}