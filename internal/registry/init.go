package registry

import (
	"errors"
	"fmt"

	"ml-reg/internal/blob"
	"ml-reg/internal/config"
	"ml-reg/internal/metadata"
)

// Init initializes the registry with the given configuration.
func (r *Registry) Init(endpoint, bucket, credsRef, region, metadataPath string, force bool) error {
	if endpoint == "" || bucket == "" {
		return errors.New("missing required parameters: endpoint and bucket are required")
	}

	if metadataPath == "" {
		metadataPath = ".ml-reg/registry.db"
	}

	cfg := &config.Config{
		Endpoint:     endpoint,
		Bucket:       bucket,
		CredsRef:     credsRef,
		Region:       region,
		MetadataPath: metadataPath,
	}

	configPath := ".ml-reg/config.json"

	if err := config.Save(configPath, cfg, force); err != nil {
		return err
	}

	metadataStore, err := metadata.NewSQLiteStore(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to initialize metadata store: %w", err)
	}
	defer metadataStore.Close()

	if _, err := blob.NewS3Store(cfg); err != nil {
		return fmt.Errorf("failed to create blob store: %w", err)
	}

	return nil
}