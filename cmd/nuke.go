package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"keys/db"

	"github.com/spf13/cobra"
)

var nukeCmd = &cobra.Command{
	Use:   "nuke",
	Short: "Delete all keys from the active profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile := db.GetActiveProfile()
		fmt.Printf("This will delete ALL keys from profile %q.\n", profile)
		fmt.Print("Type 'nuke' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input != "nuke" {
			fmt.Println("Cancelled.")
			return nil
		}

		count, err := db.NukeKeys()
		if err != nil {
			return err
		}
		fmt.Printf("Deleted %d key(s) from profile %q\n", count, profile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(nukeCmd)
}
