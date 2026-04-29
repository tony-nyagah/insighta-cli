package profiles

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"insighta-cli/internal/client"
	"insighta-cli/internal/display"
)

var (
	exportFormat  string
	exportGender  string
	exportCountry string
	exportAgeGroup string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export profiles to a file",
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "csv", "Export format (currently: csv)")
	exportCmd.Flags().StringVar(&exportGender, "gender", "", "Filter by gender")
	exportCmd.Flags().StringVar(&exportCountry, "country", "", "Filter by country code")
	exportCmd.Flags().StringVar(&exportAgeGroup, "age-group", "", "Filter by age group")
}

func runExport(cmd *cobra.Command, args []string) error {
	if exportFormat != "csv" {
		return fmt.Errorf("only --format csv is supported")
	}

	params := url.Values{"format": {"csv"}}
	if exportGender != "" {
		params.Set("gender", exportGender)
	}
	if exportCountry != "" {
		params.Set("country_id", exportCountry)
	}
	if exportAgeGroup != "" {
		params.Set("age_group", exportAgeGroup)
	}

	path := "/api/profiles/export?" + params.Encode()

	spin := display.NewSpinner("Downloading CSV export...")
	spin.Start()
	raw, contentDisposition, err := client.New().GetRaw(path)
	spin.Stop()

	if err != nil {
		return err
	}

	// Derive filename from Content-Disposition or generate one
	filename := filenameFromDisposition(contentDisposition)
	if filename == "" {
		filename = fmt.Sprintf("profiles_%s.csv", time.Now().UTC().Format("20060102T150405Z"))
	}

	dest := filepath.Join(".", filename)
	if err := os.WriteFile(dest, raw, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	display.Success(fmt.Sprintf("Saved to %s (%d bytes)", dest, len(raw)))
	return nil
}

var reFilename = regexp.MustCompile(`filename="([^"]+)"`)

func filenameFromDisposition(header string) string {
	if header == "" {
		return ""
	}
	m := reFilename.FindStringSubmatch(header)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}
