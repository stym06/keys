package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/stym06/keys/db"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [file]",
	Short: "Verify required keys are present",
	Long: `Check that all keys listed in a requirements file exist in the current profile.

Reads key names (one per line) from .keys.required by default, or a custom file.
Lines starting with # are ignored.

Examples:
  keys check
  keys check requirements.txt`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := ".keys.required"
		if len(args) == 1 {
			file = args[0]
		}

		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("cannot open %s: %w", file, err)
		}
		defer f.Close()

		var required []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			required = append(required, line)
		}
		if err := scanner.Err(); err != nil {
			return err
		}

		if len(required) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No keys listed in", file)
			return nil
		}

		allKeys, err := db.GetAllKeys()
		if err != nil {
			return err
		}

		have := make(map[string]bool)
		for _, k := range allKeys {
			have[k.Name] = true
		}

		missing := 0
		for _, name := range required {
			if have[name] {
				fmt.Fprintf(cmd.OutOrStdout(), "  ✓ %s\n", name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  ✗ %s — missing\n", name)
				missing++
			}
		}

		if missing > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d of %d keys missing.\n", missing, len(required))
			os.Exit(1)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\nAll %d keys present.\n", len(required))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
