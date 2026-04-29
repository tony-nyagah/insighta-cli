package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"insighta-cli/internal/client"
	"insighta-cli/internal/credentials"
	"insighta-cli/internal/display"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Invalidate your session and clear stored credentials",
	RunE:  runLogout,
}

func runLogout(cmd *cobra.Command, args []string) error {
	creds, err := credentials.Load()
	if err != nil {
		// Already logged out — treat as success
		display.Success("Already logged out.")
		return nil
	}

	apiURL := os.Getenv("INSIGHTA_API_URL")
	if apiURL == "" {
		apiURL = "https://api.insighta.app"
	}

	// Best-effort server-side revocation
	client.PostNoAuth(apiURL+"/auth/logout", map[string]string{
		"refresh_token": creds.RefreshToken,
	})

	if err := credentials.Clear(); err != nil {
		return fmt.Errorf("failed to clear credentials: %w", err)
	}
	display.Success("Logged out successfully.")
	return nil
}

// whoamiCmd ---------------------------------------------------------

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the currently authenticated user",
	RunE:  runWhoami,
}

func runWhoami(cmd *cobra.Command, args []string) error {
	creds, err := credentials.Load()
	if err != nil {
		return err
	}

	c := client.New()
	raw, status, err := c.Do("GET", "/api/profiles", nil) // just need any authed call
	// Actually hit a dedicated endpoint if available; for now decode local creds
	_ = raw
	_ = status
	_ = err
	_ = c

	// Whoami reads from local credentials — no extra API call needed
	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID        string `json:"id"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			AvatarURL string `json:"avatar_url"`
			Role      string `json:"role"`
		} `json:"user"`
	}

	// We don't store the full user in credentials — ask the backend
	apiURL := os.Getenv("INSIGHTA_API_URL")
	if apiURL == "" {
		apiURL = "https://api.insighta.app"
	}

	rawMe, statusMe, errMe := client.New().Do("GET", "/auth/me", nil)
	if errMe == nil && statusMe == 200 {
		if jsonErr := json.Unmarshal(rawMe, &result); jsonErr == nil && result.User.Username != "" {
			fmt.Printf("Username : @%s\n", result.User.Username)
			fmt.Printf("Email    : %s\n", result.User.Email)
			fmt.Printf("Role     : %s\n", result.User.Role)
			return nil
		}
	}

	// Fallback to local credentials
	fmt.Printf("Username : @%s\n", creds.Username)
	fmt.Printf("Role     : %s\n", creds.Role)
	return nil
}
