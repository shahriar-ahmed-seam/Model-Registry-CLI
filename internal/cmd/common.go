package cmd

import (
	"ml-reg/internal/registry"
)

// defaultConfigPath is the location of the Config_Store relative to the current
// working directory. All subcommands resolve the registry from this path.
const defaultConfigPath = ".ml-reg/config.json"

// openRegistry loads the Config_Store from the default path and constructs a
// fully wired Registry: the SQLite Metadata_Store is opened at the configured
// path and the S3-compatible Blob_Store is built from the persisted config.
//
// If no Config_Store exists at the default path, config.Load (invoked by
// registry.New) returns registry.ErrNotInitialized, which maps to the
// documented "not initialized" exit code. This helper is shared by the push,
// pull, and list commands (tasks 14.3-14.5); callers are responsible for
// closing the returned Registry via Close.
func openRegistry() (*registry.Registry, error) {
	return registry.New(defaultConfigPath)
}
