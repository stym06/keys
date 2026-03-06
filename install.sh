#!/bin/sh
set -e

REPO="stym06/keys"
INSTALL_DIR="/usr/local/bin"

main() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
    esac

    case "$OS" in
        linux|darwin) ;;
        *) echo "Unsupported OS: $OS"; exit 1 ;;
    esac

    # Check if Go is available
    if command -v go >/dev/null 2>&1; then
        echo "Installing keys via go install..."
        go install "github.com/${REPO}@latest"
        GOBIN=$(go env GOPATH)/bin
        echo "Installed keys to ${GOBIN}/keys"
        if ! echo "$PATH" | grep -q "$GOBIN"; then
            echo "Add to your PATH: export PATH=\"\$PATH:${GOBIN}\""
        fi
        exit 0
    fi

    # No Go — download source and build
    echo "Go not found. Install Go from https://go.dev/dl/ and run:"
    echo "  go install github.com/${REPO}@latest"
    exit 1
}

main
