use std::fs;
use std::path::Path;

use toml_edit::{DocumentMut, InlineTable, Item, Table, Value};

use crate::util::write_atomic;

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

pub(crate) fn read_config_summary(config_path: &Path) -> Result<ConfigSummary, String> {
    let contents = fs::read_to_string(config_path)
        .map_err(|error| format!("Failed to read config {}: {error}", config_path.display()))?;
    let document = contents
        .parse::<DocumentMut>()
        .map_err(|error| format!("Failed to parse TOML: {error}"))?;
    let gaps = document
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

pub(crate) fn update_config(config_path: &Path, gap_size: i64) -> Result<(), String> {
    let contents = fs::read_to_string(config_path)
        .map_err(|error| format!("Failed to read config {}: {error}", config_path.display()))?;
    let mut document = contents
        .parse::<DocumentMut>()
        .map_err(|error| format!("Failed to parse TOML: {error}"))?;

    update_gaps(&mut document, gap_size).map_err(|error| {
        format!(
            "Failed to update gaps in {}: {error}",
            config_path.display()
        )
    })?;
    write_atomic(config_path, document.to_string())?;
    Ok(())
}

pub(crate) fn update_gaps(document: &mut DocumentMut, gap_size: i64) -> Result<(), String> {
    update_gap_side(document, "right", gap_size)?;
    update_gap_side(document, "left", gap_size)?;
    Ok(())
}

pub(crate) fn update_gap_side(
    document: &mut DocumentMut,
    side: &str,
    gap_size: i64,
) -> Result<(), String> {
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
        update_gaps(&mut document, 111).unwrap();

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
    fn read_config_summary_includes_gaps() {
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

        let summary = read_config_summary(&path).unwrap();
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
}
