package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"keys/db"
	"keys/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Interactively select keys to write to .env file",
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := db.GetAllKeys()
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			fmt.Println("No keys stored. Use 'keys add <name> <value>' first.")
			return nil
		}

		m := tui.NewSelector(keys)
		p := tea.NewProgram(m)
		result, err := p.Run()
		if err != nil {
			return err
		}

		final := result.(tui.SelectorModel)
		if final.Aborted() {
			fmt.Println("Cancelled.")
			return nil
		}

		selected := final.SelectedKeys()
		if len(selected) == 0 {
			fmt.Println("No keys selected.")
			return nil
		}

		// Prompt for directory
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
		return nil
	},
}

func promptDirectory() string {
	fmt.Print("Directory for .env [.]: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return "."
	}
	return input
}

func init() {
	rootCmd.AddCommand(envCmd)
}
