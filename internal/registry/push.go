package registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"ml-reg/internal/hashing"
	"ml-reg/internal/metadata"
)

// PushResult describes the outcome of a successful Push operation.
type PushResult struct {
	ContentHash  string
	Deduplicated bool
	SizeBytes    int64
}

// Push pushes a model file to the registry with the given tag and optional accuracy.
func (r *Registry) Push(modelPath, tag string, accuracy *float64) (PushResult, error) {
	if r.metadata == nil || r.blob == nil {
		return PushResult{}, errors.New("registry not properly initialized")
	}

	_, err := r.metadata.GetByTag(tag)
	if err == nil {
		return PushResult{}, ErrTagExists
	}
	if !IsTagNotFound(err) {
		return PushResult{}, fmt.Errorf("failed to check if tag exists: %w", err)
	}

	fileInfo, err := os.Stat(modelPath)
	if err != nil {
		if os.IsNotExist(err) {
			return PushResult{}, ErrInvalidPath
		}
		return PushResult{}, fmt.Errorf("failed to stat model file: %w", err)
	}
	if !fileInfo.Mode().IsRegular() {
		return PushResult{}, ErrInvalidPath
	}

	file, err := os.Open(modelPath)
	if err != nil {
		return PushResult{}, ErrInvalidPath
	}
	file.Close()

	contentHash, sizeBytes, err := hashing.HashFile(modelPath)
	if err != nil {
		return PushResult{}, fmt.Errorf("failed to hash model file: %w", err)
	}

	ctx := context.Background()
	exists, err := r.blob.Exists(ctx, contentHash)
	if err != nil {
		return PushResult{}, fmt.Errorf("failed to check if blob exists: %w", err)
	}

	deduplicated := exists

	if !exists {
		file, err := os.Open(modelPath)
		if err != nil {
			return PushResult{}, ErrInvalidPath
		}
		defer file.Close()

		err = r.blob.Put(ctx, contentHash, file, sizeBytes)
		if err != nil {
			return PushResult{}, fmt.Errorf("%w: %v", ErrUploadFailed, err)
		}
	}

	record := metadata.VersionRecord{
		Tag:         tag,
		ContentHash: contentHash,
		StorageKey:  contentHash,
		Accuracy:    accuracy,
		SizeBytes:   sizeBytes,
		CreatedAt:   time.Now(),
	}

	err = r.metadata.Insert(record)
	if err != nil {
		return PushResult{}, fmt.Errorf("failed to create version record: %w", err)
	}

	return PushResult{
		ContentHash:  contentHash,
		Deduplicated: deduplicated,
		SizeBytes:    sizeBytes,
	}, nil
}

// Helper function to check if an error is ErrTagNotFound
func IsTagNotFound(err error) bool {
	return errors.Is(err, ErrTagNotFound)
}