package registry

import (
	"errors"
	"testing"
)

func TestExitCodeFor(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"nil error", nil, 0},
		{"ErrNotInitialized", ErrNotInitialized, 3},
		{"ErrAlreadyInitialized", ErrAlreadyInitialized, 4},
		{"ErrTagExists", ErrTagExists, 5},
		{"ErrTagNotFound", ErrTagNotFound, 6},
		{"ErrInvalidPath", ErrInvalidPath, 7},
		{"ErrUploadFailed", ErrUploadFailed, 8},
		{"ErrIntegrity", ErrIntegrity, 9},
		{"ErrDestinationExists", ErrDestinationExists, 10},
		{"generic error", errors.New("some error"), 1},
		{"flag error", errors.New("flag required"), 2},
		{"usage error", errors.New("usage: ml-reg"), 2},
		{"required error", errors.New("required flag"), 2},
		{"argument error", errors.New("invalid argument"), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExitCodeFor(tt.err)
			if got != tt.expected {
				t.Errorf("ExitCodeFor(%v) = %d, want %d", tt.err, got, tt.expected)
			}
		})
	}
}