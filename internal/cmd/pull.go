package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"ml-reg/internal/registry"

	"github.com/spf13/cobra"
)

// newPullCmd returns the fully implemented `pull` subcommand.
func newPullCmd() *cobra.Command {
	var (
		force bool
	)

	cmd := &cobra.Command{
		Use:   "pull <tag> <dest>",
		Short: "Pull a model version from the registry",
		Long: `Pull a model version by its tag to the specified destination path.

The downloaded file's integrity is verified by re-computing its SHA256 hash
and comparing it to the recorded hash. Downloads use a temporary file and
atomic rename to ensure partial writes don't corrupt the destination.

If the destination file already exists, use --force to overwrite it.

Examples:
  # Pull tag v1 to current directory
  ml-reg pull v1 ./model-v1.pkl

  # Pull tag latest and overwrite if exists
  ml-reg pull latest ./model.pkl --force

  # Pull to a specific directory
  ml-reg pull v2 /models/production/model.pkl`,
		Args: cobra.ExactArgs(2), // Exactly two arguments: tag and dest
		RunE: func(cmd *cobra.Command, args []string) error {
			tag := args[0]
			destPath := args[1]

			// Validate destination path
			if destPath == "" {
				return fmt.Errorf("destination path cannot be empty")
			}

			// Handle special destination "-" for stdout
			if destPath == "-" {
				// For stdout output, we need different handling
				// For now, we'll implement file-based pull first
				return fmt.Errorf("stdout destination not yet implemented")
			}

			// Ensure the destination directory exists
			destDir := filepath.Dir(destPath)
			if destDir != "." && destDir != "" {
				if err := os.MkdirAll(destDir, 0755); err != nil {
					return fmt.Errorf("failed to create destination directory: %w", err)
				}
			}

			// Create registry instance (loads config from default path)
			reg, err := registry.New(".ml-reg/config.json")
			if err != nil {
				return err // This will be ErrNotInitialized or config error
			}
			defer reg.Close()

			// Call the registry's Pull method
			err = reg.Pull(tag, destPath, force)
			if err != nil {
				return err
			}

			// Success message
			fmt.Fprintf(os.Stdout, "Pulled tag %q to %q\n", tag, destPath)

			return nil
		},
	}

	// Optional flag
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite destination file if it exists")

	// Argument completion hints for tags
	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// First argument: tag
		if len(args) == 0 {
			// We could load the registry and list tags for completion
			// But for simplicity, we'll just return file completion for now
			return nil, cobra.ShellCompDirectiveDefault
		}
		// Second argument: destination path (file completion)
		if len(args) == 1 {
			return nil, cobra.ShellCompDirectiveDefault
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return cmd
}