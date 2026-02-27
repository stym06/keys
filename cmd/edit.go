package cmd

import (
	"fmt"

	"keys/db"
	"keys/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:               "edit <name>",
	Short:             "Edit a stored key",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeKeyNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		key, err := db.GetKey(args[0])
		if err != nil {
			return err
		}

		m := tui.NewEdit(*key)
		p := tea.NewProgram(m)
		result, err := p.Run()
		if err != nil {
			return err
		}

		final := result.(tui.EditModel)
		if msg := final.Message(); msg != "" {
			fmt.Println(msg)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
