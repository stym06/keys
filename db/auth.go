package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	touchid "github.com/ansxuman/go-touchid"
)

func sessionPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".keys", ".session")
}

func isSessionValid() bool {
	data, err := os.ReadFile(sessionPath())
	if err != nil {
		return false
	}
	storedPPID := strings.TrimSpace(string(data))
	currentPPID := strconv.Itoa(os.Getppid())
	return storedPPID == currentPPID
}

func saveSession() {
	ppid := strconv.Itoa(os.Getppid())
	os.WriteFile(sessionPath(), []byte(ppid), 0600)
}

func Authenticate() error {
	if isSessionValid() {
		return nil
	}
	success, err := touchid.Auth(touchid.DeviceTypeAny, "access your keys")
	if err != nil {
		// Biometrics unavailable — allow access
		return nil
	}
	if !success {
		return fmt.Errorf("authentication failed")
	}
	saveSession()
	return nil
}
