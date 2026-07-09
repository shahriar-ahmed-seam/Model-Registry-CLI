package registry

import (
	"ml-reg/internal/blob"
	"ml-reg/internal/config"
	"ml-reg/internal/metadata"
)

// Registry coordinates between configuration, metadata, and blob stores.
type Registry struct {
	configPath string
	metadata   metadata.MetadataStore
	blob       blob.BlobStore
}

// New creates a new Registry instance with the given config path.
func New(configPath string) (*Registry, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	metaStore, err := metadata.NewSQLiteStore(cfg.MetadataPath)
	if err != nil {
		return nil, err
	}

	blobStore, err := blob.NewS3Store(cfg)
	if err != nil {
		metaStore.Close()
		return nil, err
	}

	return &Registry{
		configPath: configPath,
		metadata:   metaStore,
		blob:       blobStore,
	}, nil
}

// NewWithConfig creates a new Registry instance with the given configuration.
func NewWithConfig(cfg *config.Config) (*Registry, error) {
	metaStore, err := metadata.NewSQLiteStore(cfg.MetadataPath)
	if err != nil {
		return nil, err
	}

	blobStore, err := blob.NewS3Store(cfg)
	if err != nil {
		metaStore.Close()
		return nil, err
	}

	return &Registry{
		configPath: "",
		metadata:   metaStore,
		blob:       blobStore,
	}, nil
}

// NewWithStores creates a new Registry instance with injected stores.
func NewWithStores(metadataStore metadata.MetadataStore, blobStore blob.BlobStore) *Registry {
	return &Registry{
		configPath: "",
		metadata:   metadataStore,
		blob:       blobStore,
	}
}

// Close releases resources held by the registry.
func (r *Registry) Close() error {
	if err := r.metadata.Close(); err != nil {
		return err
	}
	// Blob store doesn't have a Close method in the interface
	return nil
}

// ErrNotImplemented is a temporary error for unimplemented methods.
var ErrNotImplemented = &RegistryError{Message: "method not implemented yet"}

// RegistryError represents an error from the registry.
type RegistryError struct {
	Message string
}

func (e *RegistryError) Error() string {
	return e.Message
}