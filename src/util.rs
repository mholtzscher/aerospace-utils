use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};

use tempfile::NamedTempFile;

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
