package cmd

import (
	"fmt"

	"keys/db"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:               "rm <name>",
	Short:             "Delete a stored key",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeKeyNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := db.DeleteKey(name); err != nil {
			return err
		}
		fmt.Printf("Deleted %s\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
