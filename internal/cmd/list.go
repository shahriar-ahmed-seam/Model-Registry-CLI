package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"ml-reg/internal/metadata"
	"ml-reg/internal/registry"

	"github.com/spf13/cobra"
)

// newListCmd returns the fully implemented `list` subcommand.
func newListCmd() *cobra.Command {
	var (
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List model versions in the registry",
		Long: `List all model versions stored in the registry.

Display includes tag, accuracy (if recorded), content hash, and creation timestamp.
Use --json for machine-readable output.

Examples:
  # List all versions in human-readable format
  ml-reg list

  # List in JSON format for scripting
  ml-reg list --json`,
		Args: cobra.NoArgs, // list takes no arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create registry instance (loads config from default path)
			reg, err := registry.New(".ml-reg/config.json")
			if err != nil {
				return err // This will be ErrNotInitialized or config error
			}
			defer reg.Close()

			// Call the registry's List method
			records, err := reg.List()
			if err != nil {
				return err
			}

			// Handle empty list (Requirement 4.3)
			if len(records) == 0 {
				if jsonOutput {
					// Empty JSON array for machine consumption
					fmt.Fprintln(os.Stdout, "[]")
				} else {
					fmt.Fprintln(os.Stdout, "No model versions recorded in the registry")
				}
				return nil
			}

			// Output based on format
			if jsonOutput {
				return outputJSON(records)
			}
			return outputHuman(records)
		},
	}

	// Optional flag
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

// outputJSON outputs records in JSON format.
func outputJSON(records []metadata.VersionRecord) error {
	// Create a simplified struct for JSON output
	type jsonRecord struct {
		Tag         string     `json:"tag"`
		ContentHash string     `json:"content_hash"`
		Accuracy    *float64   `json:"accuracy,omitempty"`
		SizeBytes   int64      `json:"size_bytes"`
		CreatedAt   time.Time  `json:"created_at"`
		StorageKey  string     `json:"storage_key"`
	}

	jsonRecords := make([]jsonRecord, len(records))
	for i, rec := range records {
		jsonRecords[i] = jsonRecord{
			Tag:         rec.Tag,
			ContentHash: rec.ContentHash,
			Accuracy:    rec.Accuracy,
			SizeBytes:   rec.SizeBytes,
			CreatedAt:   rec.CreatedAt,
			StorageKey:  rec.StorageKey,
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonRecords)
}

// outputHuman outputs records in human-readable format.
func outputHuman(records []metadata.VersionRecord) error {
	// Find max tag length for formatting
	maxTagLen := 0
	for _, rec := range records {
		if len(rec.Tag) > maxTagLen {
			maxTagLen = len(rec.Tag)
		}
	}
	if maxTagLen < 3 {
		maxTagLen = 3 // Minimum for "TAG" header
	}

	// Print header
	fmt.Fprintf(os.Stdout, "%-*s  %-12s  %-64s  %s\n", 
		maxTagLen, "TAG", "ACCURACY", "CONTENT HASH", "CREATED AT")
	fmt.Fprintf(os.Stdout, "%s  %s  %s  %s\n",
		strings.Repeat("-", maxTagLen),
		strings.Repeat("-", 12),
		strings.Repeat("-", 64),
		strings.Repeat("-", 19))

	// Print each record
	for _, rec := range records {
		// Format accuracy
		accStr := "-"
		if rec.Accuracy != nil {
			accStr = fmt.Sprintf("%.4f", *rec.Accuracy)
		}

		// Truncate hash for display (first 12 chars)
		hashDisplay := rec.ContentHash
		if len(hashDisplay) > 12 {
			hashDisplay = hashDisplay[:12] + "..."
		}

		// Format timestamp
		timeStr := rec.CreatedAt.Format("2006-01-02 15:04:05")

		fmt.Fprintf(os.Stdout, "%-*s  %-12s  %-64s  %s\n",
			maxTagLen, rec.Tag,
			accStr,
			hashDisplay,
			timeStr)
	}

	// Print summary
	fmt.Fprintf(os.Stdout, "\nTotal: %d model version(s)\n", len(records))
	return nil
}