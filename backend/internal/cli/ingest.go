package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/unrealsolutions/bugit/internal/db"
	"github.com/unrealsolutions/bugit/internal/ingest"
	"github.com/unrealsolutions/bugit/internal/storage"
)

// IngestCmd returns the ingest command.
func IngestCmd() *cobra.Command {
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "ingest <path-to-zip>",
		Short: "Ingest a repro bundle from a ZIP file",
		Long:  "Reads a repro bundle ZIP file and ingests it into the BugIt database.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			zipPath := args[0]
			dataDir, _ := cmd.Flags().GetString("data-dir")

			// Verify file exists
			if _, err := os.Stat(zipPath); err != nil {
				return fmt.Errorf("file not found: %s", zipPath)
			}

			// Initialize storage
			store, err := storage.New(dataDir)
			if err != nil {
				return fmt.Errorf("init storage: %w", err)
			}

			// Initialize database
			database, err := db.Open(store.DBPath())
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer database.Close()

			// Run ingestion
			ingester := ingest.New(database, store)
			result, err := ingester.IngestZipFile(zipPath)
			if err != nil {
				return fmt.Errorf("ingest failed: %w", err)
			}

			// Output result
			if outputJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Printf("Bundle ID: %s\n", result.BundleID)
			fmt.Printf("Status: %s\n", result.Status)
			fmt.Printf("Artifacts: %d\n", result.ArtifactCount)

			return nil
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	return cmd
}
