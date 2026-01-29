# AGENTS.md - AI Agent Guidelines for aerospace-utils

A small CLI for adjusting Aerospace workspace sizing based on monitor gaps.

**Stack**: Go 1.22+, Cobra, pelletier/go-toml/v2, fatih/color

## Rules

**Never commit code unless explicitly prompted by the user.**
**Always run linting after modifying code.**
**Always run formatter after modifying code.**
**Always run e2e tests after modifying code.**
**All tests must be e2e; do not write unit tests.**
**Always determine if a new test should be added BEFORE fixing or changing feature logic**

## Commands

Uses direnv with nix flake for automatic environment setup. Dev scripts are provided via the flake.

```bash
# Build
dev-build                      # dev build
dev-build-release              # release build

# Run
dev-run <args>                 # run locally

# Test (e2e only)
dev-test                       # all e2e tests
dev-test-verbose               # verbose e2e output

# Lint/format
dev-fmt                        # format code
dev-vet                        # static analysis
dev-lint                       # comprehensive linting (golangci-lint)
dev-check                      # run all checks (fmt, vet, lint, test)

# Dependencies
dev-tidy                       # go mod tidy

# Nix build/run
nix build                      # build package
nix run                        # run package
```

## Project Structure

```
aerospace-utils/
├── main.go                     # Entry point
├── cmd/
│   ├── root.go                 # Root command, global flags, imports workspace
│   └── workspace/              # 'workspace' subcommand package
│       ├── workspace.go        # Parent command
│       ├── use.go              # 'workspace use' subcommand
│       ├── adjust.go           # 'workspace adjust' subcommand
│       ├── shift.go            # 'workspace shift' subcommand
│       └── current.go          # 'workspace current' subcommand
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
│   │   ├── display_linux_test.go # Linux display tests
│   │   └── display_other.go    # Stub for other platforms
│   ├── gaps/
│   │   ├── gaps.go             # Gap calculations, validation
│   │   └── gaps_test.go        # Gap calculation tests
│   └── output/
│       └── output.go           # Colored output helpers
├── test/
│   └── testscript/
│       └── testscript_test.go  # E2E testscript tests
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

- All tests are e2e only; do not run unit tests.
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
