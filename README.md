# ccyolo

Smart permission filter for Claude Code that auto-approves safe operations.

```
Claude Code needs permission → ccyolo evaluates → Auto-approve or ask user
```

## WARNING: USE AT YOUR OWN RISK

This tool automatically approves operations without explicit user confirmation. While it includes safety guardrails and uses Claude to evaluate requests, **you are solely responsible for any actions taken**.

## Installation

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/9roads/ccyolo/main/install.sh | bash
```

### From Source

```bash
go install github.com/9roads/ccyolo@latest
```

## Setup

```bash
# 1. Register hook with Claude Code
ccyolo install

# 2. Store API key in system keychain (interactive)
ccyolo setup

# 3. Restart Claude Code
```

## Usage

```bash
ccyolo              # Show status
ccyolo enable       # Enable auto-approval
ccyolo disable      # Disable (ask user for everything)
ccyolo preset NAME  # Set preset: strict, balanced, permissive
ccyolo update       # Self-update to latest version
ccyolo uninstall    # Remove hook from Claude Code
```

## Presets

| Preset | Behavior |
|--------|----------|
| `strict` | Only auto-approve read operations |
| `balanced` | Auto-approve common dev tasks (default) |
| `permissive` | Auto-approve almost everything |

### What Gets Auto-Approved (balanced)

**Always approved:**
- Reading files, searching code
- `git status`, `git log`, `git diff`, `git add`
- Listing directories, version checks

**Approved by Claude AI:**
- File writes/edits in projects
- Package installation (`npm install`, `pip install`)
- Running tests, builds, linters
- `git commit`

**Always asks user:**
- `sudo` commands
- `git push --force`
- Creating PRs/releases
- Deleting important files
- Writing to system directories

## API Key

ccyolo needs an Anthropic API key for AI-based safety evaluation.

**Option 1: System keychain (recommended)**
```bash
ccyolo setup
```

**Option 2: Environment variable**
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
# or
export CCYOLO_API_KEY="sk-ant-..."
```

## How It Works

1. Claude Code requests permission for a tool use
2. ccyolo intercepts via PreToolUse hook
3. Checks static allow/deny rules (fast path)
4. If no rule matches, asks Claude API to evaluate
5. Caches decision (24h TTL)
6. Auto-approves safe ops, asks user for risky ones

## Configuration

Config stored in `~/.config/ccyolo/config.json`:

```json
{
  "enabled": true,
  "preset": "balanced",
  "model": "claude-sonnet-4-20250514",
  "cache_ttl": 86400,
  "verbose": false
}
```

## Uninstall

```bash
ccyolo uninstall        # Remove hook from Claude Code
sudo rm /usr/local/bin/ccyolo  # Remove binary
```

## License

MIT

## Disclaimer

THIS SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND. IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY ARISING FROM THE USE OF THIS SOFTWARE.
