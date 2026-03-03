# Changelog

## 0.3.0

- Add `keys inject` command — output keys as inline env vars or Docker `-e` flags
- Supports `--docker` / `-d` for Docker format, `--all` / `-a` for all keys, `--profile` / `-p` to target a specific profile
- Shell completion for key names with multiple argument support

## 0.2.0

- Add Touch ID authentication (macOS) — biometric prompt before accessing keys
- Session caching: authenticate once per terminal session, subsequent commands skip the prompt
- Graceful degradation on non-macOS or when biometrics are unavailable
- Add `keys version` command

## 0.1.0

- Initial release
- Store, retrieve, edit, and delete API keys locally with SQLite
- Interactive TUI for browsing, searching, and selecting keys
- Peek mode with masked values
- Export to `.env` files and `export` statements
- Import from `.env` files
- Profile support for isolating keys by project
- Shell completions (zsh, bash, fish)
