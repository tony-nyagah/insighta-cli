package profiles

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"insighta-cli/internal/client"
	"insighta-cli/internal/display"
)

var createName string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new profile (admin only)",
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().StringVar(&createName, "name", "", "Full name for the profile (required)")
	createCmd.MarkFlagRequired("name")
}

func runCreate(cmd *cobra.Command, args []string) error {
	spin := display.NewSpinner(fmt.Sprintf("Creating profile for %q...", createName))
	spin.Start()
	raw, status, err := client.New().Do("POST", "/api/profiles", map[string]string{"name": createName})
	spin.Stop()

	if err != nil {
		return err
	}
	if status == 403 {
		return fmt.Errorf("only admins can create profiles")
	}
	if status == 409 {
		return fmt.Errorf("a profile with that name already exists")
	}
	if status != 201 {
		return fmt.Errorf("server error (%d): %s", status, string(raw))
	}

	var resp struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("unexpected response: %w", err)
	}

	display.Success(fmt.Sprintf("Profile created for %q", createName))
	display.SingleProfile(resp.Data)
	return nil
}
