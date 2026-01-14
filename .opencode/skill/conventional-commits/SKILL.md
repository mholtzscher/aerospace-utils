---
name: conventional-commits
description: Write standardized git commit messages following the Conventional Commits 1.0.0 specification. Use when committing changes or when the user asks for help with commit messages.
license: MIT
metadata:
  version: "1.0"
  spec-version: "1.0.0"
---

# Conventional Commits

Write commit messages in this format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

## Types

| Type | Use When | SemVer |
|------|----------|--------|
| `feat` | Adding new functionality | MINOR |
| `fix` | Fixing a bug | PATCH |
| `docs` | Documentation only changes | - |
| `style` | Formatting, whitespace, semicolons (no logic) | - |
| `refactor` | Restructuring code without changing behavior | - |
| `test` | Adding or modifying tests | - |

## Rules

1. **Type is required** - Always start with a type from the table above
2. **Scope is optional** - Add context in parentheses: `feat(parser): ...`
3. **Description** - Imperative mood, lowercase start, no period, under 72 chars
4. **Body** - Separate from description with blank line, explain "why" not "what"
5. **Footer** - Use for references: `Refs: #123`, `Reviewed-by: Name`

## Breaking Changes

Mark breaking changes with `!` after type/scope:

```
feat(api)!: change authentication flow
```

Or use a `BREAKING CHANGE:` footer:

```
feat: update config format

BREAKING CHANGE: config files must now use YAML instead of JSON
```

Breaking changes trigger a MAJOR version bump regardless of type.

## Examples

**Good commits:**

```
feat: add gap presets for common monitor sizes
```

```
fix: handle missing aerospace config gracefully
```

```
docs: clarify installation steps in README
```

```
refactor: extract monitor detection into separate module
```

```
test: add unit tests for gap calculation
```

**Bad commits:**

```
Fixed stuff          # no type, vague description
feat: Add Feature.   # uppercase, period at end
update              # no type, no description
feat(): add thing    # empty scope parentheses
```

## When NOT to Use

- **Merge commits** - Let git generate these automatically
- **WIP commits** - Use `--amend` or squash before final commit
- **Reverts** - Use `revert:` type with reference to original commit SHA

## Choosing Between Types

- Changed behavior? `feat` (new) or `fix` (corrected)
- Same behavior, different code? `refactor`
- Same behavior, different formatting? `style`
- Only tests changed? `test`
- Only docs changed? `docs`
