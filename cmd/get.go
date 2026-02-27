package cmd

import (
	"fmt"

	"keys/db"
	"keys/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:               "get [name]",
	Short:             "Print the value of a stored key",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeKeyNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		if len(args) == 1 {
			key, err := db.GetKey(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(out, key.Value)
			return nil
		}

		// No arg: launch interactive picker
		keys, err := db.GetAllKeys()
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			fmt.Fprintln(out, "No keys stored. Use 'keys add <name> <value>' first.")
			return nil
		}

		m := tui.NewPicker(keys)
		p := tea.NewProgram(m)
		result, err := p.Run()
		if err != nil {
			return err
		}

		final := result.(tui.PickerModel)
		if final.Aborted() {
			return nil
		}
		if picked := final.Picked(); picked != nil {
			fmt.Fprintln(out, picked.Value)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
