package registry

import (
	"ml-reg/internal/metadata"
)

// List returns all version records in the registry.
// It retrieves all Version_Records from the Metadata_Store, ordered by creation timestamp.
//
// Returns:
//   - slice of VersionRecord on success (empty slice if no records exist)
//   - ErrNotInitialized if registry is not properly initialized
//   - other errors for metadata store issues
//
// Requirements validated:
//   4.1 - Retrieves all Version_Records from the Metadata_Store
//   4.3 - Returns empty slice when no records exist (CLI layer displays message)
//   4.4 - Returns ErrNotInitialized if registry is not initialized
//
// Implementation follows the design document's list flow:
//   1. Check if registry is properly initialized (metadata store exists)
//   2. Call metadata.List() to retrieve all records
//   3. Return records (empty slice if none) or error
func (r *Registry) List() ([]metadata.VersionRecord, error) {
	// Check if registry is properly initialized
	// Requirement 4.4: Return ErrNotInitialized if registry is not initialized
	if r.metadata == nil {
		return nil, ErrNotInitialized
	}

	// Retrieve all records from metadata store (Requirement 4.1)
	records, err := r.metadata.List()
	if err != nil {
		return nil, err
	}

	// Return records (empty slice if none) - Requirement 4.3
	// The CLI layer will handle displaying "no model versions recorded" message
	return records, nil
}