package cmd

import (
	"fmt"

	"keys/db"

	"github.com/spf13/cobra"
)

var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "Print export statements for all stored keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := db.GetAllKeys()
		if err != nil {
			return err
		}
		for _, k := range keys {
			fmt.Printf("export %s=%s\n", k.Name, k.Value)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exposeCmd)
}
