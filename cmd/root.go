package cmd

import (
	"os"

	"keys/db"

	"github.com/spf13/cobra"
)

// commands that don't need Touch ID
var noAuthCommands = map[string]bool{
	"profile":    true,
	"completion": true,
	"help":       true,
	"version":    true,
}

var rootCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage API keys locally",
}

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		name := cmd.Name()
		if p := cmd.Parent(); p != nil && p != rootCmd {
			name = p.Name()
		}
		if noAuthCommands[name] {
			return nil
		}
		return db.Authenticate()
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
