package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"keys/db"
	ksync "keys/sync"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync keys between machines on the local network",
}

var syncServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve keys for another machine to pull",
	Long: `Start a temporary server that broadcasts on the local network.
Another machine can discover it and pull your keys using the displayed passphrase.

The server shuts down after one successful sync or when you press Ctrl+C.

Examples:
  keys sync serve
  keys sync serve --profile dev`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profileFlag, _ := cmd.Flags().GetString("profile")
		profile := profileFlag
		if profile == "" {
			profile = db.GetActiveProfile()
		}

		keys, err := db.GetAllKeysForProfile(profile)
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			fmt.Println("No keys to serve in profile:", profile)
			return nil
		}

		passphrase, err := ksync.GeneratePassphrase()
		if err != nil {
			return err
		}

		server := ksync.NewServer(passphrase, profile)
		port, err := server.Start()
		if err != nil {
			return err
		}
		defer server.Stop()

		fmt.Printf("Serving %d keys from profile %q\n", len(keys), profile)
		fmt.Printf("Passphrase: %s\n", passphrase)
		fmt.Printf("Port: %d\n\n", port)
		fmt.Println("Waiting for connections... (Ctrl+C to stop)")

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-server.Done():
			fmt.Println("\nSync complete.")
		case <-sigCh:
			fmt.Println("\nStopped.")
		}

		return nil
	},
}

var syncPullCmd = &cobra.Command{
	Use:   "pull [host:port]",
	Short: "Pull keys from another machine",
	Long: `Discover peers on the local network and pull their keys.

If a host:port is provided, connects directly without discovery.

Examples:
  keys sync pull                    # auto-discover peers via mDNS
  keys sync pull 192.168.1.10:7331  # connect directly`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		if len(args) == 1 {
			fmt.Print("Enter passphrase: ")
			passphrase, _ := reader.ReadString('\n')
			passphrase = strings.TrimSpace(passphrase)

			result, err := ksync.PullDirect(args[0], passphrase)
			if err != nil {
				return err
			}
			printSyncResult(result)
			return nil
		}

		fmt.Println("Scanning for peers...")
		peers, err := ksync.DiscoverPeers(5 * time.Second)
		if err != nil {
			return err
		}

		if len(peers) == 0 {
			fmt.Println("No peers found. Make sure the other machine is running 'keys sync serve'.")
			return nil
		}

		fmt.Println()
		for i, p := range peers {
			fmt.Printf("  %d. %s (%s:%d)\n", i+1, p.Name, p.Addr, p.Port)
		}
		fmt.Println()

		var selected ksync.Peer
		if len(peers) == 1 {
			selected = peers[0]
			fmt.Printf("Found 1 peer: %s\n", selected.Name)
		} else {
			fmt.Print("Select a peer: ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			idx, err := strconv.Atoi(input)
			if err != nil || idx < 1 || idx > len(peers) {
				return fmt.Errorf("invalid selection")
			}
			selected = peers[idx-1]
		}

		fmt.Print("Enter passphrase: ")
		passphrase, _ := reader.ReadString('\n')
		passphrase = strings.TrimSpace(passphrase)

		result, err := ksync.Pull(selected, passphrase)
		if err != nil {
			return err
		}
		printSyncResult(result)
		return nil
	},
}

func printSyncResult(r *ksync.SyncResult) {
	total := r.Added + r.Updated + r.Skipped
	fmt.Printf("Synced. %d keys: %d added, %d updated, %d skipped.\n", total, r.Added, r.Updated, r.Skipped)
}

func init() {
	syncServeCmd.Flags().StringP("profile", "p", "", "profile to serve (default: active profile)")
	syncCmd.AddCommand(syncServeCmd)
	syncCmd.AddCommand(syncPullCmd)
	rootCmd.AddCommand(syncCmd)
}
