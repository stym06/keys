package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"keys/db"
	"keys/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var seeCmd = &cobra.Command{
	Use:   "see",
	Short: "Search and view stored keys, or add new ones",
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := db.GetAllKeys()
		if err != nil {
			return err
		}

		m := tui.NewSee(keys)
		p := tea.NewProgram(m)
		result, err := p.Run()
		if err != nil {
			return err
		}

		final := result.(tui.SeeModel)
		if msg := final.Message(); msg != "" {
			fmt.Println(msg)
		}

		// Handle ctrl+e env export
		if final.EnvExport() {
			selected := final.EnvExportKeys()
			if len(selected) > 0 {
				dir := promptDirectory()
				envPath := filepath.Join(dir, ".env")
				f, err := os.Create(envPath)
				if err != nil {
					return err
				}
				defer f.Close()
				for _, k := range selected {
					fmt.Fprintf(f, "%s=%s\n", k.Name, k.Value)
				}
				fmt.Printf("Wrote %d key(s) to %s\n", len(selected), envPath)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(seeCmd)
}
