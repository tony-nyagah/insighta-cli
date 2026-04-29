package profiles

import "github.com/spf13/cobra"

// ProfilesCmd is the parent "profiles" command.
var ProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage profiles",
}

func init() {
	ProfilesCmd.AddCommand(listCmd)
	ProfilesCmd.AddCommand(getCmd)
	ProfilesCmd.AddCommand(searchCmd)
	ProfilesCmd.AddCommand(createCmd)
	ProfilesCmd.AddCommand(exportCmd)
}
