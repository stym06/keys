package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"keys/db"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import keys from a .env file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()

		var newCount, updatedCount int
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Handle export prefix
			line = strings.TrimPrefix(line, "export ")

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Strip surrounding quotes
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') ||
					(value[0] == '\'' && value[len(value)-1] == '\'') {
					value = value[1 : len(value)-1]
				}
			}

			if name == "" {
				continue
			}

			exists, err := db.KeyExists(name)
			if err != nil {
				return err
			}

			if err := db.AddKey(name, value); err != nil {
				return err
			}

			if exists {
				updatedCount++
			} else {
				newCount++
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}

		total := newCount + updatedCount
		fmt.Printf("Imported %d keys (%d new, %d updated)\n", total, newCount, updatedCount)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
