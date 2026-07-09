package registry

import (
	"testing"
)

func TestRegistryErrorNotImplemented(t *testing.T) {
	// Create a minimal registry with nil stores (for testing only)
	// In real tests, we'd use proper mocks
	registry := &Registry{}
	
	// Test that methods return appropriate errors when stores are nil
	// (Note: registry.Push returns "registry not properly initialized" not ErrNotImplemented)
	_, err := registry.Push("test.pkl", "v1", nil)
	if err == nil {
		t.Error("Push should return an error when stores are nil")
	}
	if err.Error() != "registry not properly initialized" {
		t.Errorf("Push error = %v, want 'registry not properly initialized'", err)
	}
	
	err = registry.Pull("v1", "dest.pkl", false)
	if err == nil {
		t.Error("Pull should return an error when stores are nil")
	}
	
	_, err = registry.List()
	if err == nil {
		t.Error("List should return an error when stores are nil")
	}
	
	err = registry.Init("endpoint", "bucket", "creds", "region", "db.sqlite", false)
	if err == nil {
		t.Error("Init should return an error when stores are nil")
	}
}