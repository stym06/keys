package cmd

import (
	"keys/db"

	"github.com/spf13/cobra"
)

// completeKeyNames returns a ValidArgsFunction that suggests stored key names.
func completeKeyNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
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
