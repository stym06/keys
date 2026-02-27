package cmd

import (
	"fmt"

	"keys/db"

	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage key profiles",
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		active := db.GetActiveProfile()
		profiles, err := db.ListProfiles()
		if err != nil {
			return err
		}

		// Ensure active profile is always shown
		found := false
		for _, p := range profiles {
			if p == active {
				found = true
				break
			}
		}
		if !found {
			profiles = append([]string{active}, profiles...)
		}

		for _, p := range profiles {
			if p == active {
				fmt.Printf("* %s\n", p)
			} else {
				fmt.Printf("  %s\n", p)
			}
		}
		return nil
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch to a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := db.SetActiveProfile(name); err != nil {
			return err
		}
		fmt.Printf("Switched to profile %q\n", name)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileUseCmd)
	rootCmd.AddCommand(profileCmd)
}
