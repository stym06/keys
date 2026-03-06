package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"keys/db"

	"github.com/grandcat/zeroconf"
)

type Peer struct {
	Name string
	Addr string
	Port int
}

func (p Peer) URL() string {
	return fmt.Sprintf("http://%s:%d", p.Addr, p.Port)
}

func DiscoverPeers(timeout time.Duration) ([]Peer, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	var peers []Peer

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		for entry := range entries {
			addr := ""
			if len(entry.AddrIPv4) > 0 {
				addr = entry.AddrIPv4[0].String()
			} else if len(entry.AddrIPv6) > 0 {
				addr = entry.AddrIPv6[0].String()
			}
			if addr != "" {
				peers = append(peers, Peer{
					Name: entry.Instance,
					Addr: addr,
					Port: entry.Port,
				})
			}
		}
	}()

	err = resolver.Browse(ctx, "_keys-sync._tcp", "local.", entries)
	if err != nil {
		return nil, fmt.Errorf("mDNS browse failed: %w", err)
	}

	<-ctx.Done()
	return peers, nil
}

type SyncResult struct {
	Added   int
	Updated int
	Skipped int
}

func Pull(peer Peer, passphrase string) (*SyncResult, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(peer.URL() + "/sync")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	encrypted, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	decrypted, err := Decrypt(encrypted, passphrase)
	if err != nil {
		return nil, fmt.Errorf("wrong passphrase or corrupted data")
	}

	var remoteKeys []SyncKey
	if err := json.Unmarshal(decrypted, &remoteKeys); err != nil {
		return nil, fmt.Errorf("invalid data from peer")
	}

	result := &SyncResult{}
	for _, rk := range remoteKeys {
		localKey, err := db.GetKey(rk.Name)
		if err != nil {
			// Key doesn't exist locally — add it
			if err := db.AddKey(rk.Name, rk.Value); err != nil {
				return nil, fmt.Errorf("failed to add %s: %w", rk.Name, err)
			}
			result.Added++
			continue
		}

		if rk.UpdatedAt > localKey.UpdatedAt {
			if err := db.UpdateKey(rk.Name, rk.Name, rk.Value); err != nil {
				return nil, fmt.Errorf("failed to update %s: %w", rk.Name, err)
			}
			result.Updated++
		} else {
			result.Skipped++
		}
	}

	return result, nil
}

func PullDirect(addr, passphrase string) (*SyncResult, error) {
	peer := Peer{Name: addr, Addr: addr}
	// Parse host:port if provided
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			peer.Addr = addr[:i]
			fmt.Sscanf(addr[i+1:], "%d", &peer.Port)
			break
		}
	}
	if peer.Port == 0 {
		return nil, fmt.Errorf("invalid address, use host:port format")
	}
	return Pull(peer, passphrase)
}
