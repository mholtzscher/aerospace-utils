use std::fs;
use std::path::Path;

use toml_edit::{DocumentMut, Item, Value};

use crate::util::write_atomic;

pub(crate) fn update_config(config_path: &Path, gap_size: i64) -> Result<(), String> {
    let contents = fs::read_to_string(config_path)
        .map_err(|error| format!("Failed to read config {}: {error}", config_path.display()))?;
    let mut document = contents
        .parse::<DocumentMut>()
        .map_err(|error| format!("Failed to parse TOML: {error}"))?;

    update_gaps(&mut document, gap_size)?;
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

    let entry = array
        .get_mut(1)
        .ok_or_else(|| format!("gaps.outer.{side}[1] is missing in config"))?;

    let table = match entry {
        Value::InlineTable(table) => table,
        _ => return Err(format!("gaps.outer.{side}[1] is not a table in config")),
    };

    table.insert("monitor.main", Value::from(gap_size));
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

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
}
