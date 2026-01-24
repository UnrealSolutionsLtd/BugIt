package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/unrealsolutions/bugit/internal/db"
	"github.com/unrealsolutions/bugit/internal/storage"
)

// InspectCmd returns the inspect command.
func InspectCmd() *cobra.Command {
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "inspect <bundle_id>",
		Short: "Show details of a repro bundle",
		Long:  "Displays full details of a repro bundle including all artifacts, tags, and notes.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundleID := args[0]
			dataDir, _ := cmd.Flags().GetString("data-dir")

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

			// Get bundle
			bundle, err := database.GetBundle(bundleID)
			if err != nil {
				return fmt.Errorf("get bundle: %w", err)
			}

			if bundle == nil {
				return fmt.Errorf("bundle not found: %s", bundleID)
			}

			// Output
			if outputJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(bundle)
			}

			// Human-readable output
			fmt.Printf("Bundle: %s\n", bundle.BundleID)
			fmt.Printf("  Content Hash: %s\n", bundle.ContentHash)
			fmt.Printf("  Build ID:     %s\n", bundle.BuildID)
			fmt.Printf("  Map Name:     %s\n", bundle.MapName)
			fmt.Printf("  Platform:     %s\n", bundle.Platform)
			fmt.Printf("  RVR Version:  %s\n", bundle.RVRVersion)
			fmt.Printf("  Schema:       %s\n", bundle.SchemaVersion)
			fmt.Printf("  Size:         %s\n", formatBytes(bundle.SizeBytes))
			fmt.Printf("  Created:      %s\n", bundle.CreatedAt.Format("2006-01-02 15:04:05 MST"))

			// Tags
			if len(bundle.Tags) > 0 {
				fmt.Printf("  Tags:         %s\n", strings.Join(bundle.Tags, ", "))
			}

			// Artifacts
			fmt.Printf("\nArtifacts (%d):\n", len(bundle.Artifacts))
			if len(bundle.Artifacts) > 0 {
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "  ID\tFILENAME\tTYPE\tSIZE")
				for _, a := range bundle.Artifacts {
					fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n",
						a.ArtifactID,
						a.Filename,
						a.ArtifactType,
						formatBytes(a.SizeBytes),
					)
				}
				w.Flush()
			}

			// Notes
			if len(bundle.Notes) > 0 {
				fmt.Printf("\nQA Notes (%d):\n", len(bundle.Notes))
				for _, n := range bundle.Notes {
					fmt.Printf("  [%s] %s (%s):\n", n.NoteID, n.Author, n.CreatedAt.Format("2006-01-02 15:04"))
					fmt.Printf("    %s\n", n.Content)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	return cmd
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
