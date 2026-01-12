use std::fs;
use std::path::Path;

use toml_edit::{DocumentMut, Item, Table, Value};

use crate::util::write_atomic;

#[derive(Debug, Clone)]
pub(crate) struct WorkspaceState {
    pub(crate) current: Option<i64>,
    pub(crate) default_percentage: Option<i64>,
}

#[derive(Debug)]
pub(crate) struct StateLoad {
    pub(crate) state: WorkspaceState,
    pub(crate) migrated: bool,
}

pub(crate) fn read_state_file(path: &Path, dry_run: bool) -> Result<Option<StateLoad>, String> {
    if !path.exists() {
        return Ok(None);
    }

    let contents = fs::read_to_string(path)
        .map_err(|error| format!("Failed to read state file {}: {error}", path.display()))?;
    let parsed = parse_state_contents(&contents)?;

    if parsed.migrated && !dry_run {
        persist_state(path, &parsed.state)?;
    }

    Ok(Some(parsed))
}

fn parse_optional_integer(table: &Table, key: &str) -> Result<Option<i64>, String> {
    let Some(item) = table.get(key) else {
        return Ok(None);
    };

    item.as_value()
        .and_then(Value::as_integer)
        .map(Some)
        .ok_or_else(|| format!("State file key '{key}' is not an integer"))
}

pub(crate) fn parse_state_contents(contents: &str) -> Result<StateLoad, String> {
    let trimmed = contents.trim();
    if trimmed.is_empty() {
        return Err("State file is empty".to_string());
    }

    if let Ok(value) = trimmed.parse::<i64>() {
        return Ok(StateLoad {
            state: WorkspaceState {
                current: Some(value),
                default_percentage: Some(value),
            },
            migrated: true,
        });
    }

    let document = trimmed
        .parse::<DocumentMut>()
        .map_err(|error| format!("Failed to parse TOML state file: {error}"))?;
    let workspace = document
        .get("workspace")
        .and_then(Item::as_table)
        .ok_or_else(|| "State file missing [workspace] table".to_string())?;
    let current = parse_optional_integer(workspace, "current")?;
    let default_percentage = parse_optional_integer(workspace, "default")?;

    Ok(StateLoad {
        state: WorkspaceState {
            current,
            default_percentage,
        },
        migrated: false,
    })
}

pub(crate) fn resolve_percentage(
    percent: Option<i64>,
    state: Option<&StateLoad>,
) -> Result<Option<i64>, String> {
    if let Some(percent) = percent {
        return Ok(Some(percent));
    }

    let Some(state) = state else {
        return Ok(None);
    };

    Ok(state.state.current.or(state.state.default_percentage))
}

pub(crate) fn write_state(
    path: &Path,
    percentage: i64,
    existing_state: Option<WorkspaceState>,
    set_default: bool,
) -> Result<(), String> {
    let default_percentage = if set_default {
        Some(percentage)
    } else if let Some(state) = existing_state {
        state.default_percentage.or(Some(percentage))
    } else {
        Some(percentage)
    };

    let state = WorkspaceState {
        current: Some(percentage),
        default_percentage,
    };

    persist_state(path, &state)
}

pub(crate) fn persist_state(path: &Path, state: &WorkspaceState) -> Result<(), String> {
    let mut document = DocumentMut::new();
    let mut workspace = Table::new();

    if let Some(current) = state.current {
        workspace.insert("current", Item::Value(Value::from(current)));
    }
    if let Some(default_percentage) = state.default_percentage {
        workspace.insert("default", Item::Value(Value::from(default_percentage)));
    }

    document["workspace"] = Item::Table(workspace);
    let mut toml = document.to_string();
    if !toml.ends_with('\n') {
        toml.push('\n');
    }
    write_atomic(path, toml)
}

pub(crate) fn missing_state_file_message(state_path: &Path) -> String {
    format!(
        "State file not found at {}.\nRun `aerospace-utils size <percentage>` first.",
        state_path.display()
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse_state_contents_handles_toml_and_legacy_integer() {
        let toml_state = r#"
[workspace]
current = 40
default = 50
"#;
        let parsed = parse_state_contents(toml_state).unwrap();
        assert_eq!(parsed.state.current, Some(40));
        assert_eq!(parsed.state.default_percentage, Some(50));
        assert!(!parsed.migrated);

        let legacy_state = parse_state_contents("40").unwrap();
        assert_eq!(legacy_state.state.current, Some(40));
        assert_eq!(legacy_state.state.default_percentage, Some(40));
        assert!(legacy_state.migrated);
    }
}
