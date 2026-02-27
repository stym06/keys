package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"keys/db"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <name> <value>",
	Short: "Store an API key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, value := args[0], args[1]

		exists, err := db.KeyExists(name)
		if err != nil {
			return err
		}

		if exists {
			fmt.Printf("Key %q already exists. [o]verwrite / [e]dit / [c]ancel: ", name)
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))

			switch input {
			case "o":
				// proceed with overwrite
			case "e":
				fmt.Printf("Run: keys edit %s\n", name)
				return nil
			default:
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := db.AddKey(name, value); err != nil {
			return err
		}
		fmt.Printf("Stored %s\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
