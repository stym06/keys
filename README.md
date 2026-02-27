# keys

A fast, local CLI for managing API keys and secrets. Built with Go, SQLite, and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Keys are stored locally in `~/.keys/keys.db` — nothing leaves your machine.

## Install

```bash
go install github.com/stym06/keys@latest
```

Or build from source:

```bash
git clone https://github.com/stym06/keys.git
cd keys
go install .
```

## Shell Completions

Enable tab-completion for key names:

```bash
# zsh (add to ~/.zshrc)
source <(keys completion zsh)

# bash (add to ~/.bashrc)
source <(keys completion bash)

# fish
keys completion fish | source
```

With [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions), you get inline ghost suggestions as you type.

## Usage

### Store a key

```bash
keys add OPENAI_KEY sk-abc123
```

If the key already exists, you'll be prompted to overwrite, edit, or cancel.

### Get a key

```bash
keys get OPENAI_KEY        # prints the value
keys get                   # interactive typeahead picker
```

### Browse keys

```bash
keys see
```

Interactive search with checkboxes. Keybindings:

| Key | Action |
|-----|--------|
| Type | Filter keys |
| Space | Toggle checkbox |
| Tab | Copy selected as `KEY=VAL` |
| S-tab / Ctrl+Y | Copy selected as `export KEY=VAL` |
| Ctrl+E | Export selected to `.env` file |
| Enter | Add a new key (when no matches) |
| Esc | Quit |

Keys show age indicators: green (< 30 days), yellow (30-90 days), red (> 90 days).

### Peek (masked view)

```bash
keys peek
```

Same as `see` but values are hidden as `***`. Press `r` to reveal the key under the cursor.

### Edit a key

```bash
keys edit OPENAI_KEY
```

Opens a TUI editor for the key name and value. Tab switches fields, Enter saves.

### Delete a key

```bash
keys rm OPENAI_KEY
```

### Export

```bash
keys env                   # interactive .env file generator
keys expose                # print export statements to stdout
```

`keys env` lets you select keys and choose a directory for the `.env` file.

### Import from .env

```bash
keys import .env
```

Parses `.env` files — handles comments, quotes, and `export` prefixes.

### Profiles

Isolate keys by project or environment:

```bash
keys profile use dev       # switch to "dev" profile
keys add DEV_DB localhost  # stored under "dev"
keys profile use default   # switch back
keys profile list          # show all profiles (* = active)
```

### Nuke

```bash
keys nuke                  # delete all keys in active profile
```

Requires typing `nuke` to confirm.

## License

MIT
