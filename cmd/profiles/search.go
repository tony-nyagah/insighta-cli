package profiles

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"insighta-cli/internal/client"
	"insighta-cli/internal/display"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Natural language search for profiles",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSearch,
}

func runSearch(cmd *cobra.Command, args []string) error {
	q := ""
	for i, a := range args {
		if i > 0 {
			q += " "
		}
		q += a
	}

	path := "/api/profiles/search?" + url.Values{"q": {q}}.Encode()

	spin := display.NewSpinner("Searching...")
	spin.Start()
	raw, status, err := client.New().Do("GET", path, nil)
	spin.Stop()

	if err != nil {
		return err
	}
	if status == 400 {
		return fmt.Errorf("could not interpret query: %q", q)
	}
	if status != 200 {
		return fmt.Errorf("server error (%d): %s", status, string(raw))
	}

	var resp struct {
		Page       int                      `json:"page"`
		Limit      int                      `json:"limit"`
		Total      int                      `json:"total"`
		TotalPages int                      `json:"total_pages"`
		Data       []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("unexpected response: %w", err)
	}

	display.ProfileTable(resp.Data)
	display.PaginationInfo(resp.Page, resp.Limit, resp.Total, resp.TotalPages)
	return nil
}
