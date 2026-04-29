package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"insighta-cli/cmd/profiles"
)

var rootCmd = &cobra.Command{
	Use:   "insighta",
	Short: "Insighta Labs+ — profile intelligence CLI",
	Long: `insighta is the command-line interface for the Insighta Labs+ platform.

Authenticate with GitHub, query profiles, export data, and manage your account.

Environment variables:
  INSIGHTA_API_URL   Backend URL (default: https://api.insighta.app)`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(profiles.ProfilesCmd)
}
