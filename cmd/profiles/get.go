package profiles

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"insighta-cli/internal/client"
	"insighta-cli/internal/display"
)

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a single profile by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	id := args[0]

	spin := display.NewSpinner("Fetching profile...")
	spin.Start()
	raw, status, err := client.New().Do("GET", "/api/profiles/"+id, nil)
	spin.Stop()

	if err != nil {
		return err
	}
	if status == 404 {
		return fmt.Errorf("profile %q not found", id)
	}
	if status != 200 {
		return fmt.Errorf("server error (%d): %s", status, string(raw))
	}

	var resp struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("unexpected response: %w", err)
	}

	display.SingleProfile(resp.Data)
	return nil
}
