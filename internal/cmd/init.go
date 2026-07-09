package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"ml-reg/internal/registry"

	"github.com/spf13/cobra"
)

// newInitCmd returns the fully implemented `init` subcommand.
func newInitCmd() *cobra.Command {
	var (
		endpoint     string
		bucket       string
		accessKey    string
		secretKey    string
		credsRef     string
		region       string
		dbPath       string
		force        bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a model registry",
		Long: `Initialize a new model registry with S3-compatible storage configuration.

The registry requires an object storage endpoint (AWS S3 or MinIO URL) and a bucket name.
Credentials can be provided via flags, environment variables, or AWS credential chain.

Examples:
  # Initialize with MinIO (local testing)
  ml-reg init --endpoint http://localhost:9000 --bucket models

  # Initialize with AWS S3 using profile
  ml-reg init --endpoint https://s3.amazonaws.com --bucket my-models --region us-east-1

  # Initialize with explicit credentials
  ml-reg init --endpoint https://s3.amazonaws.com --bucket my-models \
    --access-key AKIA... --secret-key ...`,
		Args: cobra.NoArgs, // init takes no positional arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate required flags (Cobra's Required flag doesn't handle mutual exclusivity well)
			if endpoint == "" {
				return fmt.Errorf("required flag(s) \"endpoint\" not set")
			}
			if bucket == "" {
				return fmt.Errorf("required flag(s) \"bucket\" not set")
			}

			// Determine the credentials *reference* to persist in the Config_Store.
			//
			// SECURITY: secret material is never written to config.json. Only a
			// non-secret reference (e.g. an AWS shared-config profile name) is
			// persisted; secrets are resolved at runtime from the environment or
			// the AWS default credential chain (see internal/blob/s3.go).
			//
			// The --access-key/--secret-key flags remain for convenience. When
			// supplied they are exported into this invocation's environment so the
			// AWS credential chain can pick them up for any immediate validation,
			// but they are NOT stored in the config. Later push/pull invocations
			// resolve credentials via the standard chain (AWS_ACCESS_KEY_ID /
			// AWS_SECRET_ACCESS_KEY env vars, or the shared credentials file /
			// profile recorded in creds_ref).
			finalCredsRef := credsRef
			if accessKey != "" {
				os.Setenv("AWS_ACCESS_KEY_ID", accessKey)
			}
			if secretKey != "" {
				os.Setenv("AWS_SECRET_ACCESS_KEY", secretKey)
			}

			// Set default metadata path if not provided
			if dbPath == "" {
				dbPath = ".ml-reg/registry.db"
			}

			// Ensure the directory for the database exists
			dbDir := filepath.Dir(dbPath)
			if dbDir != "." && dbDir != "" {
				if err := os.MkdirAll(dbDir, 0755); err != nil {
					return fmt.Errorf("failed to create database directory: %w", err)
				}
			}

			// Create a temporary registry instance for initialization
			// The registry doesn't need stores for init, just the Init method
			reg := &registry.Registry{}

			// Call the registry's Init method
			if err := reg.Init(endpoint, bucket, finalCredsRef, region, dbPath, force); err != nil {
				return err
			}

			// Success message
			fmt.Fprintf(os.Stdout, "Registry initialized successfully\n")
			fmt.Fprintf(os.Stdout, "  Config: .ml-reg/config.json\n")
			fmt.Fprintf(os.Stdout, "  Database: %s\n", dbPath)
			fmt.Fprintf(os.Stdout, "  Storage: %s (bucket: %s)\n", endpoint, bucket)
			if region != "" {
				fmt.Fprintf(os.Stdout, "  Region: %s\n", region)
			}

			return nil
		},
	}

	// Required flags
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "Object storage endpoint URL (required, e.g., http://localhost:9000 for MinIO, https://s3.amazonaws.com for AWS S3)")
	cmd.Flags().StringVar(&bucket, "bucket", "", "Object storage bucket name (required)")
	cmd.MarkFlagRequired("endpoint")
	cmd.MarkFlagRequired("bucket")

	// Credential flags.
	// NOTE: --access-key/--secret-key are convenience-only and are NOT persisted
	// to config.json. They are exported to this invocation's environment so the
	// AWS credential chain can use them now; later invocations resolve secrets at
	// runtime from the environment or the profile recorded in --creds-ref.
	cmd.Flags().StringVar(&accessKey, "access-key", "", "AWS access key ID (used only for this invocation via AWS_ACCESS_KEY_ID; never stored in config)")
	cmd.Flags().StringVar(&secretKey, "secret-key", "", "AWS secret access key (used only for this invocation via AWS_SECRET_ACCESS_KEY; never stored in config)")
	cmd.Flags().StringVar(&credsRef, "creds-ref", "", "Non-secret credentials reference persisted in config (e.g., an AWS shared-config profile name)")
	cmd.Flags().StringVar(&credsRef, "profile", "", "Alias for --creds-ref: AWS shared-config profile name to record as the credentials reference")

	// Optional flags
	cmd.Flags().StringVar(&region, "region", "", "AWS region (defaults to provider default)")
	cmd.Flags().StringVar(&dbPath, "db", ".ml-reg/registry.db", "Path to SQLite metadata database")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing configuration if it exists")

	return cmd
}