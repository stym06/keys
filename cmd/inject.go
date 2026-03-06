package cmd

import (
	"fmt"
	"strings"

	"github.com/stym06/keys/db"

	"github.com/spf13/cobra"
)

var injectCmd = &cobra.Command{
	Use:   "inject [key names...]",
	Short: "Output keys as inline env vars or Docker -e flags",
	Long: `Output specified keys as inline environment variables or Docker -e flags.

Examples:
  $(keys inject API_KEY DB_HOST) ./my-script.sh
  docker run $(keys inject -d API_KEY DB_HOST) my-image
  $(keys inject --all) ./my-script.sh
  $(keys inject --all --profile dev) ./my-script.sh`,
	ValidArgsFunction: completeKeyNamesMulti,
	RunE: func(cmd *cobra.Command, args []string) error {
		allFlag, _ := cmd.Flags().GetBool("all")
		dockerFlag, _ := cmd.Flags().GetBool("docker")
		profileFlag, _ := cmd.Flags().GetString("profile")

		if !allFlag && len(args) == 0 {
			return fmt.Errorf("specify key names or use --all")
		}

		profile := profileFlag
		if profile == "" {
			profile = db.GetActiveProfile()
		}

		var keys []db.Key
		var err error

		if allFlag {
			keys, err = db.GetAllKeysForProfile(profile)
		} else {
			keys, err = db.GetKeysByNamesForProfile(args, profile)
		}
		if err != nil {
			return err
		}

		if len(keys) == 0 {
			return nil
		}

		for _, k := range keys {
			_ = db.LogAccess(k.Name, "inject", "cli")
		}

		var parts []string
		for _, k := range keys {
			if dockerFlag {
				parts = append(parts, fmt.Sprintf("-e %s=%s", k.Name, k.Value))
			} else {
				parts = append(parts, fmt.Sprintf("%s=%s", k.Name, k.Value))
			}
		}

		fmt.Fprint(cmd.OutOrStdout(), strings.Join(parts, " "))
		return nil
	},
}

// completeKeyNamesMulti suggests key names and allows multiple arguments.
func completeKeyNamesMulti(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	keys, err := db.GetAllKeys()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var names []string
	for _, k := range keys {
		names = append(names, k.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	injectCmd.Flags().BoolP("docker", "d", false, "output as Docker -e flags")
	injectCmd.Flags().BoolP("all", "a", false, "inject all keys from the profile")
	injectCmd.Flags().StringP("profile", "p", "", "use a specific profile")
	rootCmd.AddCommand(injectCmd)
}
