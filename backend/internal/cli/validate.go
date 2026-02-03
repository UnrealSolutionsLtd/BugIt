package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/unrealsolutions/bugit/internal/validate"
)

// ValidateCmd returns the validate command.
func ValidateCmd() *cobra.Command {
	var jsonOutput bool
	var showSummary bool

	cmd := &cobra.Command{
		Use:   "validate <bundle_path>",
		Short: "Validate a repro bundle for consistency",
		Long: `Validates a repro bundle directory for internal consistency.

Checks include:
  - Manifest internal consistency (duration, frame count)
  - Manifest vs timing.json (frame count, timestamps)
  - Input timestamps within video duration
  - KeyDown/KeyUp pairing

Use --summary to show a human-readable overview of bundle contents.

Exit codes:
  0 - Bundle is valid
  1 - Bundle has validation errors`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundlePath := args[0]

			// Check path exists
			if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
				return fmt.Errorf("bundle path does not exist: %s", bundlePath)
			}

			// Summary mode - just show bundle contents overview
			if showSummary {
				summary, err := validate.SummarizeBundle(bundlePath)
				if err != nil {
					return fmt.Errorf("summarizing bundle: %w", err)
				}
				
				if jsonOutput {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					if err := enc.Encode(summary); err != nil {
						return fmt.Errorf("encoding JSON: %w", err)
					}
				} else {
					fmt.Print(validate.FormatSummary(summary))
				}
				return nil
			}

			// Validation mode
			result := validate.ValidateBundle(bundlePath)

			if jsonOutput {
				// JSON output
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(result); err != nil {
					return fmt.Errorf("encoding JSON: %w", err)
				}
			} else {
				// Human-readable output
				fmt.Print(validate.FormatResult(result))
			}

			// Return error if invalid (sets exit code 1)
			if !result.Valid {
				return fmt.Errorf("bundle validation failed")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&showSummary, "summary", false, "Show bundle contents summary instead of validation")

	return cmd
}
