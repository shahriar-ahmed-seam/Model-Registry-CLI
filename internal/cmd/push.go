package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newPushCmd returns the fully implemented `push` subcommand (task 14.3).
//
// It accepts a single positional <model-path>, a required --tag flag, and an
// optional --accuracy float. The registry is constructed from the persisted
// Config_Store via the shared openRegistry helper, then the push use case is
// invoked. Errors returned from RunE propagate to the root command's Execute,
// which maps them to exit codes and prints them to stderr; success/skip
// messages are written to stdout.
func newPushCmd() *cobra.Command {
	var (
		tag      string
		accuracy float64
	)

	cmd := &cobra.Command{
		Use:   "push <model-path>",
		Short: "Push a model version to the registry",
		Long: `Push a model file to the registry under a unique tag.

The model file is hashed with SHA256 for content addressing. If a blob with
the same hash already exists in storage, the upload is skipped
(deduplication) and only a new Version_Record is created. The tag must be
unique within the registry.

Examples:
  # Push a model with tag v1
  ml-reg push model.pkl --tag v1

  # Push with an accuracy metric
  ml-reg push model.pkl --tag v2 --accuracy 0.95`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelPath := args[0]

			// Only pass a non-nil accuracy pointer when the user actually set
			// the flag, so that "not supplied" is distinguished from 0.0
			// (Requirement 2.5).
			var accPtr *float64
			if cmd.Flags().Changed("accuracy") {
				accPtr = &accuracy
			}

			// Construct the registry from config. Returns
			// registry.ErrNotInitialized when the Config_Store is absent
			// (Requirement 2.8).
			reg, err := openRegistry()
			if err != nil {
				return err
			}
			defer reg.Close()

			// Invoke the push use case. Content hashing (2.1),
			// deduplication (2.2), upload (2.3), and record creation
			// (2.4, 2.5) are handled by the registry.
			result, err := reg.Push(modelPath, tag, accPtr)
			if err != nil {
				return err
			}

			// Success message to stdout. When the blob was deduplicated, the
			// upload was skipped because identical content already exists in
			// storage (Requirement 2.2); otherwise report the normal upload.
			if result.Deduplicated {
				fmt.Fprintf(cmd.OutOrStdout(),
					"Content already exists in storage; skipped upload. Registered %s as tag %q\n",
					modelPath, tag)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Pushed %s as tag %q\n", modelPath, tag)
			}
			if accPtr != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  accuracy: %g\n", *accPtr)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Tag identifying this model version (required)")
	cmd.MarkFlagRequired("tag")

	cmd.Flags().Float64Var(&accuracy, "accuracy", 0.0, "Optional accuracy metric for this model version")

	return cmd
}
