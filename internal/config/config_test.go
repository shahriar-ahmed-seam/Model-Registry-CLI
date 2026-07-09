package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ml-reg/internal/errors"
)

func TestConfigStruct(t *testing.T) {
	cfg := &Config{
		Endpoint:     "https://s3.amazonaws.com",
		Bucket:       "my-bucket",
		CredsRef:     "default",
		Region:       "us-east-1",
		MetadataPath: "/path/to/registry.db",
	}

	// Test JSON marshaling
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	var cfg2 Config
	if err := json.Unmarshal(data, &cfg2); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if cfg.Endpoint != cfg2.Endpoint || cfg.Bucket != cfg2.Bucket || cfg.CredsRef != cfg2.CredsRef ||
		cfg.Region != cfg2.Region || cfg.MetadataPath != cfg2.MetadataPath {
		t.Errorf("config round-trip failed: got %+v, want %+v", cfg2, cfg)
	}
}

func TestExists(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Should not exist initially
	if Exists(configPath) {
		t.Errorf("Exists() = true for non-existent file, want false")
	}

	// Create the file
	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should exist now
	if !Exists(configPath) {
		t.Errorf("Exists() = false for existing file, want true")
	}
}

func TestLoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "non-existent.json")

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("Load() should return error for non-existent file")
	}
	if !isErrNotInitialized(err) {
		t.Errorf("Load() error = %v, want ErrNotInitialized", err)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		Endpoint:     "https://minio.example.com",
		Bucket:       "test-bucket",
		CredsRef:     "minio-user",
		Region:       "",
		MetadataPath: filepath.Join(tmpDir, "registry.db"),
	}

	// Save with force=false (should succeed since file doesn't exist)
	if err := Save(configPath, cfg, false); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load should succeed
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify loaded config matches original
	if cfg.Endpoint != loaded.Endpoint || cfg.Bucket != loaded.Bucket || cfg.CredsRef != loaded.CredsRef ||
		cfg.Region != loaded.Region || cfg.MetadataPath != loaded.MetadataPath {
		t.Errorf("loaded config mismatch: got %+v, want %+v", loaded, cfg)
	}
}

func TestSaveWithoutForceOnExisting(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg1 := &Config{
		Endpoint: "https://first.example.com",
		Bucket:   "first-bucket",
		CredsRef: "first-creds",
	}
	cfg2 := &Config{
		Endpoint: "https://second.example.com",
		Bucket:   "second-bucket",
		CredsRef: "second-creds",
	}

	// Save first config
	if err := Save(configPath, cfg1, false); err != nil {
		t.Fatalf("Save() failed for first config: %v", err)
	}

	// Try to save second config without force (should fail)
	err := Save(configPath, cfg2, false)
	if err == nil {
		t.Fatal("Save() should return error when file exists and force=false")
	}
	if !isErrAlreadyInitialized(err) {
		t.Errorf("Save() error = %v, want ErrAlreadyInitialized", err)
	}

	// Verify first config is still there
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if loaded.Endpoint != cfg1.Endpoint {
		t.Errorf("config was overwritten: got Endpoint=%s, want %s", loaded.Endpoint, cfg1.Endpoint)
	}
}

func TestSaveWithForceOnExisting(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg1 := &Config{
		Endpoint: "https://first.example.com",
		Bucket:   "first-bucket",
		CredsRef: "first-creds",
	}
	cfg2 := &Config{
		Endpoint: "https://second.example.com",
		Bucket:   "second-bucket",
		CredsRef: "second-creds",
	}

	// Save first config
	if err := Save(configPath, cfg1, false); err != nil {
		t.Fatalf("Save() failed for first config: %v", err)
	}

	// Save second config with force (should succeed)
	if err := Save(configPath, cfg2, true); err != nil {
		t.Fatalf("Save() with force=true failed: %v", err)
	}

	// Verify second config is now there
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if loaded.Endpoint != cfg2.Endpoint {
		t.Errorf("config was not overwritten: got Endpoint=%s, want %s", loaded.Endpoint, cfg2.Endpoint)
	}
}



// Helper functions to check error types
func isErrNotInitialized(err error) bool {
	return err != nil && err.Error() == errors.ErrNotInitialized.Error()
}

func isErrAlreadyInitialized(err error) bool {
	return err != nil && err.Error() == errors.ErrAlreadyInitialized.Error()
}