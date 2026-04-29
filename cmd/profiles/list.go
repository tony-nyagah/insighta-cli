package profiles

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"insighta-cli/internal/client"
	"insighta-cli/internal/display"
)

var (
	listGender    string
	listAgeGroup  string
	listCountry   string
	listMinAge    int
	listMaxAge    int
	listSortBy    string
	listOrder     string
	listPage      int
	listLimit     int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List profiles with optional filters",
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVar(&listGender, "gender", "", "Filter by gender (male|female)")
	listCmd.Flags().StringVar(&listAgeGroup, "age-group", "", "Filter by age group (child|teenager|adult|senior)")
	listCmd.Flags().StringVar(&listCountry, "country", "", "Filter by country code (e.g. NG)")
	listCmd.Flags().IntVar(&listMinAge, "min-age", 0, "Minimum age")
	listCmd.Flags().IntVar(&listMaxAge, "max-age", 0, "Maximum age")
	listCmd.Flags().StringVar(&listSortBy, "sort-by", "created_at", "Sort field (age|gender_probability|created_at)")
	listCmd.Flags().StringVar(&listOrder, "order", "desc", "Sort order (asc|desc)")
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number")
	listCmd.Flags().IntVar(&listLimit, "limit", 10, "Results per page (max 50)")
}

func runList(cmd *cobra.Command, args []string) error {
	params := url.Values{}
	if listGender != "" {
		params.Set("gender", listGender)
	}
	if listAgeGroup != "" {
		params.Set("age_group", listAgeGroup)
	}
	if listCountry != "" {
		params.Set("country_id", listCountry)
	}
	if listMinAge > 0 {
		params.Set("min_age", strconv.Itoa(listMinAge))
	}
	if listMaxAge > 0 {
		params.Set("max_age", strconv.Itoa(listMaxAge))
	}
	params.Set("sort_by", listSortBy)
	params.Set("order", listOrder)
	params.Set("page", strconv.Itoa(listPage))
	params.Set("limit", strconv.Itoa(listLimit))

	path := "/api/profiles?" + params.Encode()

	spin := display.NewSpinner("Fetching profiles...")
	spin.Start()
	raw, status, err := client.New().Do("GET", path, nil)
	spin.Stop()

	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("server error (%d): %s", status, string(raw))
	}

	var resp struct {
		Status     string                   `json:"status"`
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
