# Changelog

## 0.4.0

- Add `keys audit` command — view access history for your keys
  - `keys audit` shows summary with access counts and last used time per key
  - `keys audit --log` shows full access log with action, source, and timestamp
  - `keys audit --clear` clears the audit log
- Add `keys check` command — verify required keys are present
  - Reads `.keys.required` file (one key name per line, supports `#` comments)
  - Reports which keys are present or missing, exits with code 1 if any are missing
  - Useful for CI and agent pre-flight checks
- Access logging: `get`, `inject`, and `expose` now record access events for audit
- Add `--version` flag to root command

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
