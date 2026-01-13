# AGENTS.md - AI Agent Guidelines for aerospace-utils

A small CLI for adjusting Aerospace workspace gaps based on monitor size.

**Stack**: Go 1.22+, Cobra, pelletier/go-toml/v2, fatih/color

## Rules

**Never commit code unless explicitly prompted by the user.**
**Always run linting after modifying code.**
**Always run formatter after modifying code.**
**Always run tests after modifying code.**

## Commands

Prefer Nix tooling when available.

```bash
# Build
nix develop -c go build                       # dev build
nix develop -c go build -ldflags="-s -w"      # release build

# Run
nix develop -c go run . -- <args>             # run locally

# Test
nix develop -c go test ./...                  # all tests
nix develop -c go test ./internal/gaps        # single package
nix develop -c go test -v ./...               # verbose output
nix develop -c go test -run TestName ./...    # single test

# Lint/format
nix develop -c go fmt ./...                   # format code
nix develop -c go vet ./...                   # static analysis
nix develop -c golangci-lint run              # comprehensive linting

# Nix build/run
nix build                                     # build package
nix run                                       # run package
```

If you are not using Nix, run the same `go` commands directly.

## Project Structure

```
aerospace-utils/
├── main.go                     # Entry point
├── cmd/
│   ├── root.go                 # Root command, global flags, imports gaps
│   └── gaps/                   # 'gaps' subcommand package
│       ├── gaps.go             # Parent command
│       ├── gaps_use.go         # 'gaps use' subcommand
│       ├── gaps_adjust.go      # 'gaps adjust' subcommand
│       └── gaps_current.go     # 'gaps current' subcommand
├── internal/
│   ├── aerospace/
│   │   └── aerospace.go        # aerospace binary wrapper
│   ├── cli/
│   │   └── options.go          # GlobalOptions (shared across packages)
│   ├── config/
│   │   ├── aerospace.go        # AerospaceConfig type
│   │   └── state.go            # WorkspaceState (per-monitor)
│   ├── display/
│   │   ├── display.go          # Display info types
│   │   ├── display_darwin.go   # macOS CGO implementation
│   │   ├── display_linux.go    # Linux xrandr implementation
│   │   └── display_other.go    # Stub for other platforms
│   ├── gaps/
│   │   └── gaps.go             # Gap calculations, validation
│   └── output/
│       └── output.go           # Colored output helpers
├── go.mod
├── go.sum
└── flake.nix
```

## Code Style

### Imports

Order: stdlib -> external packages -> internal packages, separated by blank lines.
Use goimports or let `go fmt` handle ordering.

### Types

- Percentages are `int64` throughout the codebase.
- Use `string` for file paths.
- Use `*T` (pointer) for optional values instead of sentinel values.
- Return `error` for error conditions.

### Naming

- Types: `PascalCase` (WorkspaceState, MonitorState)
- Functions/methods: `PascalCase` for exported, `camelCase` for unexported
- Constants: `PascalCase` for exported, `camelCase` for unexported
- Use descriptive names (`percentage`, `statePath`)

### Error Handling

- Functions return `(T, error)` tuple.
- Wrap errors with context: `fmt.Errorf("context: %w", err)`.
- Check errors immediately after function calls.
- Use user-facing messages, not debug dumps.

### Formatting

- Run `go fmt ./...` before committing.
- Use `goimports` for import organization.
- Let the tooling handle formatting decisions.

### File IO

- Use `WriteAtomic` for config/state writes.
- Preserve behavior around `DryRun` and `NoReload` flags.
- Use `os/exec` for running external commands.

### TOML

- State data uses TOML (`WorkspaceState`).
- Config edits use regex replacement to preserve formatting.
- Keep parsing tolerant but validate required fields.

### Testing

- Tests live in `*_test.go` files alongside the code.
- Use table-driven tests where appropriate.
- Use `t.Run` for subtests.
- Prefer exact assertions.

## CLI/UX Guidelines

- `fmt.Println` for normal output, `fmt.Fprintln(os.Stderr, ...)` for errors.
- Use fatih/color for colored output, respect `NO_COLOR` env var.
- Avoid breaking existing CLI flags or subcommands.

## Platform Notes

- Tool is macOS-only for full functionality.
- Monitor detection uses CoreGraphics via CGO on macOS.
- Linux builds use xrandr for display detection (development support).
- Non-macOS/Linux builds compile but display detection returns errors.
- `aerospace` binary must exist in `PATH`.

## Dependency Updates

- Update `go.mod` and run `go mod tidy`.
- Avoid new dependencies unless required.
- Prefer the standard library before adding packages.

## Repo Hygiene

- Keep changes minimal and focused.
- Avoid mass reformatting unless necessary.
- Run `go mod tidy` after dependency changes.
