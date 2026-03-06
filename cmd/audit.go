package cmd

import (
	"fmt"
	"time"

	"github.com/stym06/keys/db"

	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Show key access history",
	Long: `Show a summary of when and how often keys have been accessed.

Examples:
  keys audit              # summary: access counts per key
  keys audit --log        # full access log (most recent first)
  keys audit --log -n 20  # last 20 access events
  keys audit --clear      # clear the audit log`,
	RunE: func(cmd *cobra.Command, args []string) error {
		clearFlag, _ := cmd.Flags().GetBool("clear")
		logFlag, _ := cmd.Flags().GetBool("log")
		limit, _ := cmd.Flags().GetInt("limit")

		if clearFlag {
			if err := db.ClearAuditLog(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Audit log cleared.")
			return nil
		}

		if logFlag {
			return showAuditLog(cmd, limit)
		}

		return showAuditSummary(cmd)
	},
}

func showAuditSummary(cmd *cobra.Command) error {
	entries, err := db.GetAuditSummary()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No access events recorded yet.")
		return nil
	}

	// Find max key name length for alignment
	maxLen := 0
	for _, e := range entries {
		if len(e.KeyName) > maxLen {
			maxLen = len(e.KeyName)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%-*s  %s  %s\n", maxLen, "KEY", "ACCESSED", "LAST USED")
	for _, e := range entries {
		lastUsed := formatTimeAgo(e.AccessedAt)
		timesLabel := "time"
		if e.Count != 1 {
			timesLabel = "times"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%-*s  %d %s    %s\n", maxLen, e.KeyName, e.Count, timesLabel, lastUsed)
	}
	return nil
}

func showAuditLog(cmd *cobra.Command, limit int) error {
	entries, err := db.GetAuditLog(limit)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No access events recorded yet.")
		return nil
	}

	maxLen := 0
	for _, e := range entries {
		if len(e.KeyName) > maxLen {
			maxLen = len(e.KeyName)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%-*s  %-8s  %-8s  %s\n", maxLen, "KEY", "ACTION", "SOURCE", "TIME")
	for _, e := range entries {
		t := formatTimeAgo(e.AccessedAt)
		fmt.Fprintf(cmd.OutOrStdout(), "%-*s  %-8s  %-8s  %s\n", maxLen, e.KeyName, e.Action, e.Source, t)
	}
	return nil
}

func formatTimeAgo(unix int64) string {
	if unix == 0 {
		return "never"
	}
	d := time.Since(time.Unix(unix, 0))
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func init() {
	auditCmd.Flags().BoolP("log", "l", false, "show full access log")
	auditCmd.Flags().BoolP("clear", "c", false, "clear the audit log")
	auditCmd.Flags().IntP("limit", "n", 50, "number of log entries to show")
	rootCmd.AddCommand(auditCmd)
}
