package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/unrealsolutions/bugit/internal/db"
	"github.com/unrealsolutions/bugit/internal/models"
	"github.com/unrealsolutions/bugit/internal/storage"
)

// ListCmd returns the list command.
func ListCmd() *cobra.Command {
	var (
		buildID    string
		platform   string
		limit      int
		outputJSON bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ingested repro bundles",
		Long:  "Lists all ingested repro bundles with optional filtering.",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			// Query bundles
			query := &models.BundleListQuery{
				BuildID:  buildID,
				Platform: platform,
				Limit:    limit,
			}

			result, err := database.ListBundles(query)
			if err != nil {
				return fmt.Errorf("list bundles: %w", err)
			}

			// Output
			if outputJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			if len(result.Bundles) == 0 {
				fmt.Println("No bundles found.")
				return nil
			}

			// Table output
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "BUNDLE ID\tBUILD\tPLATFORM\tARTIFACTS\tCREATED\tTAGS")
			fmt.Fprintln(w, "---------\t-----\t--------\t---------\t-------\t----")

			for _, b := range result.Bundles {
				tags := "-"
				if len(b.Tags) > 0 {
					tags = strings.Join(b.Tags, ", ")
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
					b.BundleID,
					truncate(b.BuildID, 20),
					b.Platform,
					b.ArtifactCount,
					b.CreatedAt.Format("2006-01-02 15:04"),
					truncate(tags, 30),
				)
			}

			w.Flush()

			fmt.Printf("\nShowing %d of %d bundles\n", len(result.Bundles), result.Total)

			return nil
		},
	}

	cmd.Flags().StringVar(&buildID, "build-id", "", "Filter by build ID")
	cmd.Flags().StringVar(&platform, "platform", "", "Filter by platform")
	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	return cmd
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
