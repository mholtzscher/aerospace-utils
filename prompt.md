# Prompt: Recreate AeroSpace workspace sizing as a Rust CLI

You are an expert Rust CLI engineer. Recreate the AeroSpace “workspace sizing” utilities from `modules/home-manager/files/nushell/functions.nu` as a standalone Rust CLI app, but target the exact TOML layout used in `modules/home-manager/files/aerospace.toml`.

## What to implement
Implement a Rust CLI binary named `aerospace-utils` with two subcommands matching the Nushell behavior:

- `aerospace-utils gaps use [PERCENT]`
- `aerospace-utils gaps adjust [AMOUNT]` (default `AMOUNT=5`, supports negative)

## Exact TOML structure to support (important)
The config file contains (template):

```toml
[gaps]
inner.horizontal = 20
inner.vertical = 20
outer.right = [{ monitor.'DeskPad Display' = 0 }, { monitor.main = 300 }, 24]
outer.left  = [{ monitor.'DeskPad Display' = 0 }, { monitor.main = 300 }, 24]
outer.bottom = 10
outer.top = 10
```

Your program MUST update only:

- `gaps.outer.right[1].monitor.main`
- `gaps.outer.left[1].monitor.main`

Meaning: update the **second array element** (index 1) which is an inline table `{ monitor.main = 300 }`.

Do NOT modify:

- index 0 (`{ monitor.'DeskPad Display' = 0 }`)
- index 2 (`24`)
- any other keys

If `outer.left/right` or index `1` is missing or not a table, return an error (exit 1) with a clear message.

## Saved state file (TOML)
Replace the original plain-text percentage file with a TOML “state file” storing both current and default percentages.

- Default location: `~/.config/aerospace/workspace-size.toml`
- TOML schema:

```toml
[workspace]
current = 40
default = 40
```

Read semantics:
- When running `aerospace-utils gaps use` with no `PERCENT` argument:
  - If TOML exists and has `current`, use `current`.
  - If `current` is missing/null, fall back to `default`.
  - If the state file does not exist: print info `No percentage provided and no saved percentage file found` and exit 0.

Write semantics:
- `gaps use PERCENT`: update `current = PERCENT`.
- `gaps adjust AMOUNT`: load `current`, compute `new = current + AMOUNT`, then set `current = new`.
- `default` is **not** automatically changed when `current` changes.

Default management:
- Add a flag on `gaps use`: `--set-default`.
  - When provided with `gaps use PERCENT`, also set `default = PERCENT`.

Migration:
- If the state file exists but contains a legacy plain integer like `40` (not TOML): treat that value as both `current` and `default`, then rewrite the file as TOML (atomically).

## Behavior to match (from Nushell)
### `gaps use`
- Require `aerospace` binary in PATH; if missing, exit 1.
- Determine `percentage`:
  - If CLI arg provided: use it.
  - Else: load from TOML state file as described above.
- Validate `percentage` is 1..=100 else exit 1.
- On macOS: run `system_profiler SPDisplaysDataType` and parse main monitor width:
  - Find the line containing `Main Display: Yes`
  - Search backward from there for the last `Resolution:` line
  - Parse width from e.g. `Resolution: 3456 x 2234` or `Resolution: 3456 x 2234 Retina`
- Compute:
  - `workspace_percentage = percentage / 100.0`
  - `gap_percentage = (1.0 - workspace_percentage) / 2.0`
  - `gap_size = round(monitor_width * gap_percentage)` (match Nushell `math round` behavior)
- Edit config TOML at `~/.config/aerospace/aerospace.toml` and set both:
  - `gaps.outer.right[1].monitor.main = gap_size`
  - `gaps.outer.left[1].monitor.main = gap_size`
- Persist TOML state:
  - Always write `current = percentage`.
  - If `--set-default` is passed, also write `default = percentage`.
- Run `aerospace reload-config`:
  - If reload fails: warn but still exit 0 (config changes should remain).
  - If reload succeeds: print `Completed.` and exit 0.

### `gaps adjust`
- Require that the state file exists; if not, exit 1 with:
  - error about missing file
  - info telling user to run `gaps use <percentage>` first
- Read `current`, compute `new = current + amount`, validate 1..=100 else exit 1.
- Print adjustment info.
- Call the same logic as `gaps use new` (and update `current` accordingly).

## CLI options
Add:

- `--config-path <PATH>` default `~/.config/aerospace/aerospace.toml`
- `--state-path <PATH>` default `~/.config/aerospace/workspace-size.toml`
- `--no-reload`
- `--dry-run` (do not write files; just print what would change)
- `-v/--verbose`

## Platform handling
- Primary target is macOS (the original uses `system_profiler`).
- On non-macOS: exit with a clear “unsupported OS” error (exit 1).

## Implementation requirements
- Use `clap` for CLI.
- Use `toml_edit` to modify TOML without rewriting unrelated structure.
- Expand `~` to home directory.
- Use atomic writes (temp file + rename) for both the config and TOML state file.
- Use `std::process::Command` for external commands.
- Exit codes: 1 for errors/validation; reload failure is not fatal (exit 0).

## Tests
Include unit tests for:

- Parsing a sample `system_profiler` output to extract width
- Gap calculation rounding
- Editing a TOML string that matches the exact `outer.left/right` array shape and verifying only `[1].monitor.main` changes
- TOML state parsing + legacy integer migration

## Deliverable
Output complete code for:

- `Cargo.toml`
- `src/main.rs` (and modules if needed)

Also include a short run guide:

- `cargo build`
- Examples: `aerospace-utils gaps use 40`, `aerospace-utils gaps adjust -5`, `aerospace-utils gaps use` (uses saved)

Do not implement unrelated commands or features.
