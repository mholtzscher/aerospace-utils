use std::fs;
use std::path::Path;

use toml_edit::{DocumentMut, Item, Value};

use crate::util::write_atomic;

pub(crate) struct MainGapSizes {
    pub(crate) left: Option<i64>,
    pub(crate) right: Option<i64>,
}

pub(crate) fn read_main_gap_sizes(config_path: &Path) -> Result<MainGapSizes, String> {
    let contents = fs::read_to_string(config_path)
        .map_err(|error| format!("Failed to read config {}: {error}", config_path.display()))?;
    let document = contents
        .parse::<DocumentMut>()
        .map_err(|error| format!("Failed to parse TOML: {error}"))?;
    let left = read_main_gap_side(&document, "left")?;
    let right = read_main_gap_side(&document, "right")?;

    Ok(MainGapSizes { left, right })
}

fn read_main_gap_side(document: &DocumentMut, side: &str) -> Result<Option<i64>, String> {
    let gaps = document
        .get("gaps")
        .and_then(Item::as_table)
        .ok_or_else(|| "Missing [gaps] table".to_string())?;
    let outer = gaps
        .get("outer")
        .and_then(Item::as_table)
        .ok_or_else(|| "Missing [gaps.outer] table".to_string())?;
    let array_item = outer
        .get(side)
        .ok_or_else(|| format!("Missing gaps.outer.{side} entry in config"))?;

    let array = match array_item {
        Item::Value(Value::Array(array)) => array,
        _ => return Err(format!("gaps.outer.{side} is not an array in config")),
    };

    for value in array.iter() {
        if let Some(table) = value.as_inline_table() {
            if let Some(gap) = table.get("monitor.main").and_then(Value::as_integer) {
                return Ok(Some(gap));
            }
            if let Some(gap) = table
                .get("monitor")
                .and_then(Value::as_inline_table)
                .and_then(|table| table.get("main").and_then(Value::as_integer))
            {
                return Ok(Some(gap));
            }
        }
    }

    Ok(None)
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

        assert_eq!(
            right_table.get("monitor.main").and_then(Value::as_integer),
            Some(111)
        );
        assert_eq!(
            left_table.get("monitor.main").and_then(Value::as_integer),
            Some(111)
        );
        assert_eq!(right.get(0).unwrap().to_string(), original_right_first);
        assert_eq!(right.get(2).unwrap().to_string(), original_right_last);
        assert_eq!(left.get(0).unwrap().to_string(), original_left_first);
        assert_eq!(left.get(2).unwrap().to_string(), original_left_last);
    }

    #[test]
    fn read_main_gap_sizes_returns_values() {
        let temp_dir = tempdir().unwrap();
        let path = temp_dir.path().join("aerospace.toml");
        let input = r#"
 [gaps]
 outer.right = [{ monitor.main = 300 }, 24]
 outer.left  = [{ monitor.main = 250 }, 24]
 "#;
        fs::write(&path, input).unwrap();

        let gaps = read_main_gap_sizes(&path).unwrap();
        assert_eq!(gaps.left, Some(250));
        assert_eq!(gaps.right, Some(300));
    }

    #[test]
    fn read_main_gap_sizes_handles_missing_entries() {
        let temp_dir = tempdir().unwrap();
        let path = temp_dir.path().join("aerospace.toml");
        let input = r#"
 [gaps]
 outer.right = [24]
 outer.left  = [24]
 "#;
        fs::write(&path, input).unwrap();

        let gaps = read_main_gap_sizes(&path).unwrap();
        assert_eq!(gaps.left, None);
        assert_eq!(gaps.right, None);
    }
}
