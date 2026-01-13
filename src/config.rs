use std::fs;
use std::path::{Path, PathBuf};

use toml_edit::{DocumentMut, InlineTable, Item, Table, Value};

use crate::cli::CommonOptions;
use crate::util::{expand_tilde, write_atomic};

// ============================================================================
// AerospaceConfig
// ============================================================================

pub(crate) struct AerospaceConfig {
    path: PathBuf,
    doc: DocumentMut,
}

impl AerospaceConfig {
    pub(crate) fn default_path() -> Result<PathBuf, String> {
        let home_dir =
            dirs::home_dir().ok_or_else(|| "Unable to determine home directory".to_string())?;
        Ok(home_dir.join(".config/aerospace/aerospace.toml"))
    }

    pub(crate) fn resolve_path(options: &CommonOptions) -> Result<PathBuf, String> {
        let path = match &options.config_path {
            Some(path) => path.clone(),
            None => Self::default_path()?,
        };
        expand_tilde(&path)
    }

    pub(crate) fn load(path: PathBuf) -> Result<Self, String> {
        let contents = fs::read_to_string(&path)
            .map_err(|error| format!("Failed to read config {}: {error}", path.display()))?;
        let doc = contents
            .parse::<DocumentMut>()
            .map_err(|error| format!("Failed to parse TOML: {error}"))?;
        Ok(Self { path, doc })
    }

    pub(crate) fn from_options(options: &CommonOptions) -> Result<Self, String> {
        let path = Self::resolve_path(options)?;
        Self::load(path)
    }

    pub(crate) fn path(&self) -> &Path {
        &self.path
    }

    pub(crate) fn summary(&self) -> Result<ConfigSummary, String> {
        let gaps = self
            .doc
            .get("gaps")
            .and_then(Item::as_table)
            .ok_or_else(|| "Missing [gaps] table".to_string())?;
        let outer = gaps
            .get("outer")
            .and_then(Item::as_table)
            .ok_or_else(|| "Missing [gaps.outer] table".to_string())?;

        let inner = gaps.get("inner").and_then(Item::as_table);
        let inner_horizontal = inner.map_or(Ok(None), |table| {
            read_optional_integer(table, "horizontal", "gaps.inner.horizontal")
        })?;
        let inner_vertical = inner.map_or(Ok(None), |table| {
            read_optional_integer(table, "vertical", "gaps.inner.vertical")
        })?;
        let outer_top = read_optional_integer(outer, "top", "gaps.outer.top")?;
        let outer_bottom = read_optional_integer(outer, "bottom", "gaps.outer.bottom")?;
        let left_gaps = read_monitor_gaps(outer, "left")?;
        let right_gaps = read_monitor_gaps(outer, "right")?;

        Ok(ConfigSummary {
            inner_horizontal,
            inner_vertical,
            outer_top,
            outer_bottom,
            left_gaps,
            right_gaps,
        })
    }

    pub(crate) fn set_main_gaps(&mut self, gap_size: i64) -> Result<(), String> {
        update_gap_side(&mut self.doc, "right", gap_size).map_err(|error| {
            format!("Failed to update gaps in {}: {error}", self.path.display())
        })?;
        update_gap_side(&mut self.doc, "left", gap_size).map_err(|error| {
            format!("Failed to update gaps in {}: {error}", self.path.display())
        })?;
        Ok(())
    }

    pub(crate) fn write(&self) -> Result<(), String> {
        write_atomic(&self.path, self.doc.to_string())
    }
}

// ============================================================================
// WorkspaceState
// ============================================================================

#[derive(Debug, Clone)]
pub(crate) struct WorkspaceState {
    path: PathBuf,
    current: Option<i64>,
    default_percentage: Option<i64>,
    migrated: bool,
}

impl WorkspaceState {
    pub(crate) fn default_path() -> Result<PathBuf, String> {
        let home_dir =
            dirs::home_dir().ok_or_else(|| "Unable to determine home directory".to_string())?;
        Ok(home_dir.join(".config/aerospace/workspace-size.toml"))
    }

    pub(crate) fn resolve_path(options: &CommonOptions) -> Result<PathBuf, String> {
        let path = match &options.state_path {
            Some(path) => path.clone(),
            None => Self::default_path()?,
        };
        expand_tilde(&path)
    }

    pub(crate) fn load(path: PathBuf, dry_run: bool) -> Result<Option<Self>, String> {
        if !path.exists() {
            return Ok(None);
        }

        let contents = fs::read_to_string(&path)
            .map_err(|error| format!("Failed to read state file {}: {error}", path.display()))?;
        let state = parse_state_contents(&contents, path.clone())
            .map_err(|error| format!("Failed to parse state file {}: {error}", path.display()))?;

        if state.migrated && !dry_run {
            state.write()?;
        }

        Ok(Some(state))
    }

    pub(crate) fn from_options(options: &CommonOptions) -> Result<Option<Self>, String> {
        let path = Self::resolve_path(options)?;
        Self::load(path, options.dry_run)
    }

    pub(crate) fn new(path: PathBuf, current: i64, default_percentage: Option<i64>) -> Self {
        Self {
            path,
            current: Some(current),
            default_percentage,
            migrated: false,
        }
    }

    pub(crate) fn path(&self) -> &Path {
        &self.path
    }

    pub(crate) fn current(&self) -> Option<i64> {
        self.current
    }

    pub(crate) fn default_percentage(&self) -> Option<i64> {
        self.default_percentage
    }

    #[allow(dead_code)]
    pub(crate) fn was_migrated(&self) -> bool {
        self.migrated
    }

    pub(crate) fn resolve_percentage(&self, explicit: Option<i64>) -> Option<i64> {
        explicit.or(self.current).or(self.default_percentage)
    }

    pub(crate) fn update(&mut self, percentage: i64, set_default: bool) {
        if set_default || self.default_percentage.is_none() {
            self.default_percentage = Some(percentage);
        }
        self.current = Some(percentage);
    }

    pub(crate) fn write(&self) -> Result<(), String> {
        let mut document = DocumentMut::new();
        let mut workspace = Table::new();

        if let Some(current) = self.current {
            workspace.insert("current", Item::Value(Value::from(current)));
        }
        if let Some(default_percentage) = self.default_percentage {
            workspace.insert("default", Item::Value(Value::from(default_percentage)));
        }

        document["workspace"] = Item::Table(workspace);
        let mut toml = document.to_string();
        if !toml.ends_with('\n') {
            toml.push('\n');
        }
        write_atomic(&self.path, toml)
    }

    pub(crate) fn missing_file_message(state_path: &Path) -> String {
        format!(
            "State file not found at {}.\nRun `aerospace-utils gaps use <percentage>` first.",
            state_path.display()
        )
    }
}

fn parse_state_contents(contents: &str, path: PathBuf) -> Result<WorkspaceState, String> {
    let trimmed = contents.trim();
    if trimmed.is_empty() {
        return Err("State file is empty".to_string());
    }

    if let Ok(value) = trimmed.parse::<i64>() {
        return Ok(WorkspaceState {
            path,
            current: Some(value),
            default_percentage: Some(value),
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

    Ok(WorkspaceState {
        path,
        current,
        default_percentage,
        migrated: false,
    })
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

// ============================================================================
// ConfigSummary and helpers
// ============================================================================

pub(crate) struct ConfigSummary {
    pub(crate) inner_horizontal: Option<i64>,
    pub(crate) inner_vertical: Option<i64>,
    pub(crate) outer_top: Option<i64>,
    pub(crate) outer_bottom: Option<i64>,
    pub(crate) left_gaps: Vec<MonitorGap>,
    pub(crate) right_gaps: Vec<MonitorGap>,
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub(crate) struct MonitorGap {
    pub(crate) name: String,
    pub(crate) value: i64,
}

fn read_optional_integer(table: &Table, key: &str, label: &str) -> Result<Option<i64>, String> {
    let Some(item) = table.get(key) else {
        return Ok(None);
    };

    item.as_value()
        .and_then(Value::as_integer)
        .map(Some)
        .ok_or_else(|| format!("{label} is not an integer"))
}

fn read_monitor_gaps(outer: &Table, side: &str) -> Result<Vec<MonitorGap>, String> {
    let array_item = outer
        .get(side)
        .ok_or_else(|| format!("Missing gaps.outer.{side} entry in config"))?;

    let array = match array_item {
        Item::Value(Value::Array(array)) => array,
        _ => return Err(format!("gaps.outer.{side} is not an array in config")),
    };

    let mut entries = Vec::new();
    for value in array.iter() {
        if let Some(table) = value.as_inline_table() {
            append_monitor_entries(&mut entries, table)?;
        }
    }

    Ok(entries)
}

fn append_monitor_entries(
    entries: &mut Vec<MonitorGap>,
    table: &InlineTable,
) -> Result<(), String> {
    for (key, value) in table.iter() {
        let key_name = key;
        if let Some(name) = key_name.strip_prefix("monitor.") {
            let gap = value
                .as_integer()
                .ok_or_else(|| format!("monitor.{name} gap is not an integer"))?;
            entries.push(MonitorGap {
                name: name.to_string(),
                value: gap,
            });
            continue;
        }

        if key_name == "monitor" {
            let inner_table = value
                .as_inline_table()
                .ok_or_else(|| "monitor entry is not an inline table".to_string())?;
            for (inner_key, inner_value) in inner_table.iter() {
                let gap = inner_value
                    .as_integer()
                    .ok_or_else(|| format!("monitor.{inner_key} gap is not an integer"))?;
                entries.push(MonitorGap {
                    name: inner_key.to_string(),
                    value: gap,
                });
            }
        }
    }

    Ok(())
}

fn update_gap_side(document: &mut DocumentMut, side: &str, gap_size: i64) -> Result<(), String> {
    let gaps = document
        .get_mut("gaps")
        .and_then(Item::as_table_mut)
        .ok_or_else(|| "Missing [gaps] table".to_string())?;
    let outer = gaps
        .get_mut("outer")
        .and_then(Item::as_table_mut)
        .ok_or_else(|| "Missing [gaps.outer] table".to_string())?;
    let array_item = outer
        .get_mut(side)
        .ok_or_else(|| format!("Missing gaps.outer.{side} entry in config"))?;

    let array = match array_item {
        Item::Value(Value::Array(array)) => array,
        _ => return Err(format!("gaps.outer.{side} is not an array in config")),
    };

    let table = if let Some(table) = array.get_mut(1).and_then(Value::as_inline_table_mut) {
        table
    } else {
        array
            .iter_mut()
            .find_map(Value::as_inline_table_mut)
            .ok_or_else(|| {
                format!(
                    "gaps.outer.{side} has no inline tables; expected entries like {{ monitor.main = 0 }}"
                )
            })?
    };

    update_monitor_entry(table, gap_size)
}

fn update_monitor_entry(table: &mut InlineTable, gap_size: i64) -> Result<(), String> {
    if let Some(value) = table.get_mut("monitor") {
        let inner_table = value
            .as_inline_table_mut()
            .ok_or_else(|| "monitor entry is not an inline table".to_string())?;
        inner_table.insert("main", Value::from(gap_size));
        return Ok(());
    }

    if let Some(value) = table.get_mut("monitor.main") {
        *value = Value::from(gap_size);
        return Ok(());
    }

    for (key, value) in table.iter_mut() {
        if key == "monitor.main" {
            *value = Value::from(gap_size);
            return Ok(());
        }
    }

    table.insert("monitor.main", Value::from(gap_size));
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    use tempfile::tempdir;

    #[test]
    fn update_gaps_only_updates_index_one() {
        let input = r#"
 [gaps]
 inner.horizontal = 20
 inner.vertical = 20
 outer.right = [{ monitor.'DeskPad Display' = 0 }, { monitor.main = 300 }, 24]
 outer.left  = [{ monitor.'DeskPad Display' = 0 }, { monitor.main = 300 }, 24]
 outer.bottom = 10
 outer.top = 10
 "#;
        let original = input.parse::<DocumentMut>().unwrap();
        let original_right = original["gaps"]["outer"]["right"].as_array().unwrap();
        let original_left = original["gaps"]["outer"]["left"].as_array().unwrap();
        let original_right_first = original_right.get(0).unwrap().to_string();
        let original_right_last = original_right.get(2).unwrap().to_string();
        let original_left_first = original_left.get(0).unwrap().to_string();
        let original_left_last = original_left.get(2).unwrap().to_string();

        let mut document = input.parse::<DocumentMut>().unwrap();
        update_gap_side(&mut document, "right", 111).unwrap();
        update_gap_side(&mut document, "left", 111).unwrap();

        let right = document["gaps"]["outer"]["right"].as_array().unwrap();
        let left = document["gaps"]["outer"]["left"].as_array().unwrap();
        let right_table = right.get(1).and_then(Value::as_inline_table).unwrap();
        let left_table = left.get(1).and_then(Value::as_inline_table).unwrap();

        let right_monitor = right_table
            .get("monitor")
            .and_then(Value::as_inline_table)
            .unwrap();
        let left_monitor = left_table
            .get("monitor")
            .and_then(Value::as_inline_table)
            .unwrap();

        assert_eq!(
            right_monitor.get("main").and_then(Value::as_integer),
            Some(111)
        );
        assert_eq!(
            left_monitor.get("main").and_then(Value::as_integer),
            Some(111)
        );
        assert!(right_table.get("monitor.main").is_none());
        assert!(left_table.get("monitor.main").is_none());
        assert_eq!(right.get(0).unwrap().to_string(), original_right_first);
        assert_eq!(right.get(2).unwrap().to_string(), original_right_last);
        assert_eq!(left.get(0).unwrap().to_string(), original_left_first);
        assert_eq!(left.get(2).unwrap().to_string(), original_left_last);
    }

    #[test]
    fn aerospace_config_summary_includes_gaps() {
        let temp_dir = tempdir().unwrap();
        let path = temp_dir.path().join("aerospace.toml");
        let input = r#"
 [gaps]
 inner.horizontal = 12
 inner.vertical = 14
 outer.top = 8
 outer.bottom = 6
 outer.right = [{ monitor.main = 300 }, { monitor."DeskPad Display" = 0 }, 24]
 outer.left  = [{ monitor = { main = 250, "Studio Display" = 40 } }, 24]
 "#;
        fs::write(&path, input).unwrap();

        let config = AerospaceConfig::load(path).unwrap();
        let summary = config.summary().unwrap();
        assert_eq!(summary.inner_horizontal, Some(12));
        assert_eq!(summary.inner_vertical, Some(14));
        assert_eq!(summary.outer_top, Some(8));
        assert_eq!(summary.outer_bottom, Some(6));
        assert_eq!(
            summary.right_gaps,
            vec![
                MonitorGap {
                    name: "main".to_string(),
                    value: 300,
                },
                MonitorGap {
                    name: "DeskPad Display".to_string(),
                    value: 0,
                },
            ]
        );
        assert_eq!(
            summary.left_gaps,
            vec![
                MonitorGap {
                    name: "main".to_string(),
                    value: 250,
                },
                MonitorGap {
                    name: "Studio Display".to_string(),
                    value: 40,
                },
            ]
        );
    }

    #[test]
    fn parse_state_contents_handles_toml_and_legacy_integer() {
        let path = PathBuf::from("/tmp/test.toml");

        let toml_state = r#"
[workspace]
current = 40
default = 50
"#;
        let parsed = parse_state_contents(toml_state, path.clone()).unwrap();
        assert_eq!(parsed.current(), Some(40));
        assert_eq!(parsed.default_percentage(), Some(50));
        assert!(!parsed.was_migrated());

        let legacy_state = parse_state_contents("40", path).unwrap();
        assert_eq!(legacy_state.current(), Some(40));
        assert_eq!(legacy_state.default_percentage(), Some(40));
        assert!(legacy_state.was_migrated());
    }

    #[test]
    fn resolve_percentage_prefers_explicit_then_current_then_default() {
        let state = WorkspaceState {
            path: PathBuf::from("/tmp/test.toml"),
            current: Some(25),
            default_percentage: Some(50),
            migrated: false,
        };

        assert_eq!(state.resolve_percentage(Some(10)), Some(10));
        assert_eq!(state.resolve_percentage(None), Some(25));

        let state = WorkspaceState {
            path: PathBuf::from("/tmp/test.toml"),
            current: None,
            default_percentage: Some(40),
            migrated: false,
        };

        assert_eq!(state.resolve_percentage(None), Some(40));
    }

    #[test]
    fn write_state_preserves_existing_default() {
        let temp_dir = tempdir().unwrap();
        let path = temp_dir.path().join("state.toml");
        let mut state = WorkspaceState {
            path: path.clone(),
            current: Some(55),
            default_percentage: Some(70),
            migrated: false,
        };

        state.update(40, false);
        state.write().unwrap();

        let contents = fs::read_to_string(&path).unwrap();
        let parsed = parse_state_contents(&contents, path).unwrap();
        assert_eq!(parsed.current(), Some(40));
        assert_eq!(parsed.default_percentage(), Some(70));
    }

    #[test]
    fn write_state_sets_default_when_missing() {
        let temp_dir = tempdir().unwrap();
        let path = temp_dir.path().join("state.toml");
        let mut state = WorkspaceState {
            path: path.clone(),
            current: Some(30),
            default_percentage: None,
            migrated: false,
        };

        state.update(35, false);
        state.write().unwrap();

        let contents = fs::read_to_string(&path).unwrap();
        let parsed = parse_state_contents(&contents, path).unwrap();
        assert_eq!(parsed.current(), Some(35));
        assert_eq!(parsed.default_percentage(), Some(35));
    }
}
