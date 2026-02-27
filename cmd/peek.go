package cmd

import (
	"fmt"

	"keys/db"
	"keys/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var peekCmd = &cobra.Command{
	Use:   "peek",
	Short: "View keys with masked values (press r to reveal)",
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := db.GetAllKeys()
		if err != nil {
			return err
		}

		m := tui.NewPeek(keys)
		p := tea.NewProgram(m)
		result, err := p.Run()
		if err != nil {
			return err
		}

		final := result.(tui.SeeModel)
		if msg := final.Message(); msg != "" {
			fmt.Println(msg)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(peekCmd)
}
