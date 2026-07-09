// Package errors defines sentinel errors used throughout the model registry.
// Having these in a separate package breaks import cycles.
package errors

import (
	"errors"
)

// Sentinel errors for the registry operations
var (
	// ErrNotInitialized is returned when the registry is not initialized
	ErrNotInitialized = errors.New("registry is not initialized")

	// ErrAlreadyInitialized is returned when attempting to initialize an already initialized registry without force
	ErrAlreadyInitialized = errors.New("registry is already initialized")

	// ErrTagExists is returned when a tag is already in use
	ErrTagExists = errors.New("tag is already in use")

	// ErrTagNotFound is returned when a tag is not found in the metadata store
	ErrTagNotFound = errors.New("tag was not found")

	// ErrInvalidPath is returned when a model file path is not a readable file
	ErrInvalidPath = errors.New("model file path is not a readable file")

	// ErrUploadFailed is returned when upload to blob store fails
	ErrUploadFailed = errors.New("upload to blob store failed")

	// ErrIntegrity is returned when integrity verification fails
	ErrIntegrity = errors.New("integrity verification failed")

	// ErrDestinationExists is returned when destination already exists
	ErrDestinationExists = errors.New("destination already exists")
)