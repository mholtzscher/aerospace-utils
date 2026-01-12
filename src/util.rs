use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};

use tempfile::NamedTempFile;

use crate::cli::GlobalOptions;

pub(crate) fn validate_percentage(percentage: i64) -> Result<(), String> {
    if (1..=100).contains(&percentage) {
        Ok(())
    } else {
        Err(format!(
            "Percentage must be between 1 and 100 (got {percentage})"
        ))
    }
}

pub(crate) fn calculate_gap_size(monitor_width: i64, percentage: i64) -> i64 {
    let workspace_percentage = percentage as f64 / 100.0;
    let gap_percentage = (1.0 - workspace_percentage) / 2.0;
    (monitor_width as f64 * gap_percentage).round() as i64
}

fn default_config_path() -> Result<PathBuf, String> {
    if let Some(config_dir) = dirs::config_dir() {
        Ok(config_dir.join("aerospace").join("aerospace.toml"))
    } else if let Some(home_dir) = dirs::home_dir() {
        Ok(home_dir.join(".config/aerospace/aerospace.toml"))
    } else {
        Err("Unable to determine config directory".to_string())
    }
}

fn default_state_path() -> Result<PathBuf, String> {
    if let Some(config_dir) = dirs::config_dir() {
        Ok(config_dir.join("aerospace").join("workspace-size.toml"))
    } else if let Some(home_dir) = dirs::home_dir() {
        Ok(home_dir.join(".config/aerospace/workspace-size.toml"))
    } else {
        Err("Unable to determine config directory".to_string())
    }
}

pub(crate) fn resolve_config_path(options: &GlobalOptions) -> Result<PathBuf, String> {
    let path = match &options.config_path {
        Some(path) => path.clone(),
        None => default_config_path()?,
    };

    expand_tilde(&path)
}

pub(crate) fn resolve_state_path(options: &GlobalOptions) -> Result<PathBuf, String> {
    let path = match &options.state_path {
        Some(path) => path.clone(),
        None => default_state_path()?,
    };

    expand_tilde(&path)
}

pub(crate) fn expand_tilde(path: &Path) -> Result<PathBuf, String> {
    let path_str = path.to_string_lossy();
    if !path_str.starts_with('~') {
        return Ok(path.to_path_buf());
    }

    let home_dir = dirs::home_dir().ok_or_else(|| "Unable to expand ~".to_string())?;
    if path_str == "~" {
        return Ok(home_dir);
    }

    let trimmed = path_str.strip_prefix("~/").ok_or_else(|| {
        format!(
            "Unsupported path expansion for '{}'; use ~/ or absolute paths",
            path.display()
        )
    })?;

    Ok(home_dir.join(trimmed))
}

pub(crate) fn write_atomic(path: &Path, contents: String) -> Result<(), String> {
    let parent = path
        .parent()
        .ok_or_else(|| format!("Unable to determine parent of {}", path.display()))?;
    fs::create_dir_all(parent)
        .map_err(|error| format!("Failed to create {}: {error}", parent.display()))?;
    let mut temp_file = NamedTempFile::new_in(parent).map_err(|error| format!("{error}"))?;
    temp_file
        .write_all(contents.as_bytes())
        .map_err(|error| format!("Failed to write temp file: {error}"))?;
    temp_file
        .persist(path)
        .map_err(|error| format!("Failed to persist {}: {error}", path.display()))?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn gap_calculation_rounds() {
        let gap_size = calculate_gap_size(1000, 40);
        assert_eq!(gap_size, 300);
    }
}
