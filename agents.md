# Agent Instructions

## Release Workflow

When publishing a new version of `keys`, follow these steps in order:

### 1. Update version

Bump the version constant in `cmd/version.go`.

### 2. Update changelog

Add a new section to `CHANGELOG.md` with the version number and list of changes.

### 3. Commit and push

```bash
git add -A
git commit -m "Release vX.Y.Z"
git push
```

### 4. Tag and create GitHub release

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
gh release create vX.Y.Z --title "vX.Y.Z" --notes "release notes here"
```

### 5. Update Homebrew formula

Get the SHA256 of the new tarball and update the formula in `stym06/homebrew-tap`:

```bash
# Get SHA
curl -sL https://github.com/stym06/keys/archive/refs/tags/vX.Y.Z.tar.gz | shasum -a 256

# Update formula via GitHub API
gh api repos/stym06/homebrew-tap/contents/Formula/keys.rb \
  --method PUT \
  --field message="Update keys to vX.Y.Z" \
  --field content="$(base64 encoded formula)" \
  --field sha="$(current file sha)"
```

The formula lives at `stym06/homebrew-tap/Formula/keys.rb`. Update the `url` and `sha256` fields.

### 6. Publish skill

Update `skills/keys-manager/SKILL.md` with any new commands or behavior changes, then publish:

```bash
clawhub publish /path/to/skills/keys-manager --version X.Y.Z --changelog "description of changes"
```
