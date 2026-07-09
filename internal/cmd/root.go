// Package cmd implements the Cobra command-line layer for ml-reg.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"ml-reg/internal/registry"
)

// NewRootCmd builds the root ml-reg command with all subcommands registered.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "ml-reg",
		Short:         "Git-style version control for AI model files",
		Long: "ml-reg is a single-binary CLI that provides git-style version " +
			"control for large AI model files, storing blobs in S3-compatible " +
			"object storage and version metadata in a local SQLite database.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newInitCmd(), newPushCmd(), newPullCmd(), newListCmd())
	return root
}

// Execute runs the root command and returns a process exit code.
func Execute() int {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return exitCodeFor(err)
	}
	return 0
}

// exitCodeFor maps an error to its documented process exit code.
func exitCodeFor(err error) int {
	return registry.ExitCodeFor(err)
}
