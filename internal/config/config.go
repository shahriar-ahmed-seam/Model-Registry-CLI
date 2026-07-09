package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ml-reg/internal/errors"
)

// Config represents the registry configuration stored as JSON.
type Config struct {
	// Endpoint is the Object_Storage_Endpoint URL (S3/MinIO)
	Endpoint string `json:"endpoint"`
	// Bucket is the object storage bucket name
	Bucket string `json:"bucket"`
	// CredsRef is a reference to credentials (e.g., profile name or env var), not the secret itself
	CredsRef string `json:"creds_ref"`
	// Region is the optional AWS region (or compatible region)
	Region string `json:"region,omitempty"`
	// MetadataPath is the path to the SQLite database file
	MetadataPath string `json:"metadata_path"`
}

// Exists returns true if a config file exists at the given path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Load reads and parses the config file at the given path.
// If the file does not exist, it returns errors.ErrNotInitialized.
// If the file exists but cannot be read or parsed, it returns the underlying error.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.ErrNotInitialized
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the config to the given path using an atomic write pattern.
// If force is false and the file already exists, it returns errors.ErrAlreadyInitialized.
// If force is true, it overwrites the existing file.
// The write is performed to a temporary file in the same directory, then atomically renamed.
func Save(path string, c *Config, force bool) error {
	if Exists(path) && !force {
		return errors.ErrAlreadyInitialized
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a temporary file in the same directory
	tmpFile, err := os.CreateTemp(dir, "*.config.json.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary config file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		// Remove the temp file if we're not renaming it (i.e., on error)
		if _, err := os.Stat(tmpPath); err == nil {
			os.Remove(tmpPath)
		}
	}()

	// Write config as JSON with indentation for readability
	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to write config to temporary file: %w", err)
	}

	// Ensure the data is flushed to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}
	tmpFile.Close()

	// Atomically rename the temporary file to the target path
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to atomically rename config file: %w", err)
	}

	return nil
}

// String returns a JSON representation of the config (without indentation).
func (c *Config) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

// PrettyString returns a pretty-printed JSON representation of the config.
func (c *Config) PrettyString() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}