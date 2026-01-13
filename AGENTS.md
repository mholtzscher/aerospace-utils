# AGENTS.md - AI Agent Guidelines for aerospace-utils

A small CLI for adjusting Aerospace workspace gaps based on monitor size.

**Stack**: Rust 2024, clap, toml_edit, tempfile, dirs

## Rules

**Never commit code unless explicitly prompted by the user.**
**Always run clippy after modifying code.**
**Always run formatter after modifying code.**
**Always run tests after modifying code.**

## Commands

Prefer Nix tooling when available.

```bash
# Build
nix develop -c cargo build                    # dev build
nix develop -c cargo build --release          # release build
nix develop -c cargo check                    # type check only

# Run
nix develop -c cargo run -- <args>            # run locally

# Test
nix develop -c cargo test                     # all tests
nix develop -c cargo test validate_os_blocks_non_macos_without_override  # single test
nix develop -c cargo test tests::gap_calculation_rounds  # module test
nix develop -c cargo test -- --nocapture      # with output

# Lint/format
nix develop -c cargo fmt                      # format code
nix develop -c cargo fmt -- --check           # check only
nix develop -c cargo clippy -- -D warnings    # lints
nix develop -c cargo clippy --tests -- -D warnings

# Nix build/run
nix build                                     # build package
nix run                                       # run package
```

If you are not using Nix, run the same `cargo` commands directly.

## Project Structure

```
src/
  main.rs          # CLI entry + command dispatch
  cli.rs           # CLI argument definitions (clap)
  config.rs        # Aerospace TOML config handling
  state.rs         # Workspace state persistence
  system.rs        # System/OS interactions
  util.rs          # Path utilities, atomic writes
  output.rs        # Shared output formatting helpers
  gaps/            # Gaps command handlers
    mod.rs         # Gap calculations, validation
    adjust.rs      # 'gaps adjust' handler
    current.rs     # 'gaps current' handler
    size.rs        # 'gaps use' handler
```

## Code Style

### Imports

Order: std -> external crates -> local modules, separated by blank lines. Avoid glob imports.

### Types

- Percentages are `i64` throughout the codebase.
- Use `Path`/`PathBuf` for filesystem paths.
- Use `Option<T>` for optional values (see `WorkspaceState`).
- Use `String` for error messages (current pattern).

### Naming

- Types: `PascalCase` (WorkspaceState, StateLoad)
- Functions: `snake_case` (build_options, validate_percentage)
- Constants: `SCREAMING_SNAKE_CASE`
- Use descriptive names (`percentage`, `state_path`)

### Error Handling

- Functions return `Result<T, String>`; propagate with `?`.
- Add context: `map_err(|e| format!("...: {e}"))`.
- Avoid `unwrap`/`expect` outside tests.
- Use user-facing messages, not debug dumps.

### Formatting

- Run `cargo fmt` and follow rustfmt defaults.
- Keep long chained calls split across lines.
- Prefer trailing commas in multiline structures.
- Avoid manual alignment; let rustfmt decide.

### File IO

- Use `write_atomic` for config/state writes.
- Preserve behavior around `dry_run` and `no_reload`.
- Avoid shelling out unless required; use `Command` when needed.

### TOML

- State data uses TOML (`WorkspaceState`).
- Config edits use `toml_edit` to preserve formatting.
- Keep parsing tolerant but validate required fields.

### Testing

- Unit tests live in `mod tests` within each module.
- Use descriptive test names and inline fixtures.
- Prefer exact assertions on values and types.
- Avoid `#[ignore]` unless strictly necessary.

## CLI/UX Guidelines

- `println!` for normal output, `eprintln!` for warnings/errors.
- Avoid breaking existing CLI flags or subcommands.

## Platform Notes

- Tool is macOS-only (`ensure_macos`).
- Monitor detection uses CoreGraphics on macOS.
- `aerospace` binary must exist in `PATH`.

## Dependency Updates

- Update `Cargo.toml` and `Cargo.lock` together.
- Avoid new dependencies unless required.
- Prefer the standard library before adding crates.

## Repo Hygiene

- Do not edit files under `target/`.
- Keep changes minimal and focused.
- Avoid mass reformatting unless necessary.
