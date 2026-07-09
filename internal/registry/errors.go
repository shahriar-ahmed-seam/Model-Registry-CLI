package registry

import (
	stderrors "errors"
	"strings"

	"ml-reg/internal/errors"
)

// Re-export errors from the errors package for backward compatibility
var (
	ErrNotInitialized    = errors.ErrNotInitialized
	ErrAlreadyInitialized = errors.ErrAlreadyInitialized
	ErrTagExists         = errors.ErrTagExists
	ErrTagNotFound       = errors.ErrTagNotFound
	ErrInvalidPath       = errors.ErrInvalidPath
	ErrUploadFailed      = errors.ErrUploadFailed
	ErrIntegrity         = errors.ErrIntegrity
	ErrDestinationExists = errors.ErrDestinationExists
)

// ExitCodeFor maps sentinel errors to their documented exit codes.
// If the error is nil, returns 0 (success).
// If the error doesn't match any sentinel error, returns 1 (generic error).
func ExitCodeFor(err error) int {
	if err == nil {
		return 0
	}

	// Check each sentinel error
	switch {
	case stderrors.Is(err, errors.ErrNotInitialized):
		return 3
	case stderrors.Is(err, errors.ErrAlreadyInitialized):
		return 4
	case stderrors.Is(err, errors.ErrTagExists):
		return 5
	case stderrors.Is(err, errors.ErrTagNotFound):
		return 6
	case stderrors.Is(err, errors.ErrInvalidPath):
		return 7
	case stderrors.Is(err, errors.ErrUploadFailed):
		return 8
	case stderrors.Is(err, errors.ErrIntegrity):
		return 9
	case stderrors.Is(err, errors.ErrDestinationExists):
		return 10
	default:
		// Check for string-based error messages that might come from Cobra
		errStr := strings.ToLower(err.Error())
		// Check for usage/flag errors (exit code 2)
		if strings.Contains(errStr, "flag") || strings.Contains(errStr, "required") || 
		   strings.Contains(errStr, "usage") || strings.Contains(errStr, "argument") {
			return 2
		}
		// Generic error (includes unknown command)
		return 1
	}
}