package display

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

var (
	Bold    = color.New(color.Bold).SprintFunc()
	Green   = color.New(color.FgGreen).SprintFunc()
	Red     = color.New(color.FgRed).SprintFunc()
	Yellow  = color.New(color.FgYellow).SprintFunc()
	Cyan    = color.New(color.FgCyan).SprintFunc()
)

// Success prints a green success message.
func Success(msg string) { fmt.Println(Green("✓ " + msg)) }

// Error prints a red error message to stderr.
func Error(msg string) { fmt.Fprintln(os.Stderr, Red("✗ "+msg)) }

// Info prints a cyan informational message.
func Info(msg string) { fmt.Println(Cyan("→ " + msg)) }

// Spinner is a simple terminal spinner for indicating loading state.
type Spinner struct {
	msg    string
	done   chan struct{}
	frames []string
}

func NewSpinner(msg string) *Spinner {
	return &Spinner{
		msg:    msg,
		done:   make(chan struct{}),
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

func (s *Spinner) Start() {
	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				fmt.Printf("\r%-60s\r", "") // clear line
				return
			default:
				fmt.Printf("\r%s %s", Yellow(s.frames[i%len(s.frames)]), s.msg)
				time.Sleep(80 * time.Millisecond)
				i++
			}
		}
	}()
}

func (s *Spinner) Stop() { close(s.done); time.Sleep(90 * time.Millisecond) }

// ProfileTable renders a slice of profile maps as a formatted table.
func ProfileTable(profiles []map[string]interface{}) {
	if len(profiles) == 0 {
		fmt.Println(Yellow("No profiles found."))
		return
	}
	t := tablewriter.NewWriter(os.Stdout)
	t.Header([]string{"ID", "Name", "Gender", "Age", "Age Group", "Country", "Created At"})
	for _, p := range profiles {
		t.Append([]string{
			truncate(str(p["id"]), 8),
			str(p["name"]),
			str(p["gender"]),
			str(p["age"]),
			str(p["age_group"]),
			fmt.Sprintf("%s (%s)", str(p["country_name"]), str(p["country_id"])),
			formatTime(str(p["created_at"])),
		})
	}
	t.Render()
}

// SingleProfile renders one profile as a two-column key/value table.
func SingleProfile(p map[string]interface{}) {
	t := tablewriter.NewWriter(os.Stdout)
	t.Header([]string{"Field", "Value"})
	rows := [][]string{
		{"ID", str(p["id"])},
		{"Name", str(p["name"])},
		{"Gender", fmt.Sprintf("%s (%.0f%%)", str(p["gender"]), floatPct(p["gender_probability"]))},
		{"Age", str(p["age"])},
		{"Age Group", str(p["age_group"])},
		{"Country", fmt.Sprintf("%s (%s) %.0f%%", str(p["country_name"]), str(p["country_id"]), floatPct(p["country_probability"]))},
		{"Created At", formatTime(str(p["created_at"]))},
	}
	for _, r := range rows {
		t.Append(r)
	}
	t.Render()
}

// PaginationInfo prints a summary line after a list result.
func PaginationInfo(page, limit, total, totalPages int) {
	fmt.Printf("\n  Page %s of %s  ·  %s total records  ·  %s per page\n\n",
		Bold(fmt.Sprintf("%d", page)),
		Bold(fmt.Sprintf("%d", totalPages)),
		Bold(fmt.Sprintf("%d", total)),
		Bold(fmt.Sprintf("%d", limit)),
	)
}

// --- helpers ---

func str(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int(t)) {
			return fmt.Sprintf("%d", int(t))
		}
		return fmt.Sprintf("%.2f", t)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func floatPct(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f * 100
	}
	return 0
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func formatTime(s string) string {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z07:00", "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02 15:04")
		}
	}
	return s
}
