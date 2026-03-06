package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/stym06/keys/db"

	"github.com/grandcat/zeroconf"
)

type SyncKey struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	UpdatedAt int64  `json:"updated_at"`
}

type Server struct {
	passphrase string
	port       int
	profile    string
	httpServer *http.Server
	mdns       *zeroconf.Server
	done       chan struct{}
}

func NewServer(passphrase, profile string) *Server {
	return &Server{
		passphrase: passphrase,
		profile:    profile,
		done:       make(chan struct{}),
	}
}

func (s *Server) Start() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	s.port = listener.Addr().(*net.TCPAddr).Port

	hostname, _ := os.Hostname()

	mux := http.NewServeMux()
	mux.HandleFunc("/sync", s.handleSync)

	s.httpServer = &http.Server{Handler: mux}

	go s.httpServer.Serve(listener)

	s.mdns, err = zeroconf.Register(
		hostname,
		"_keys-sync._tcp",
		"local.",
		s.port,
		[]string{"keys-sync"},
		nil,
	)
	if err != nil {
		s.httpServer.Close()
		return 0, fmt.Errorf("mDNS registration failed: %w", err)
	}

	return s.port, nil
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	keys, err := db.GetAllKeysForProfile(s.profile)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	syncKeys := make([]SyncKey, len(keys))
	for i, k := range keys {
		syncKeys[i] = SyncKey{
			Name:      k.Name,
			Value:     k.Value,
			UpdatedAt: k.UpdatedAt,
		}
	}

	payload, err := json.Marshal(syncKeys)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	encrypted, err := Encrypt(payload, s.passphrase)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(encrypted)

	// Signal that sync is done
	select {
	case s.done <- struct{}{}:
	default:
	}
}

func (s *Server) Done() <-chan struct{} {
	return s.done
}

func (s *Server) Stop() {
	if s.mdns != nil {
		s.mdns.Shutdown()
	}
	if s.httpServer != nil {
		s.httpServer.Shutdown(context.Background())
	}
}
