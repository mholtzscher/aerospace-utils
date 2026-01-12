use std::env;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::Command;

#[cfg(target_os = "macos")]
use std::convert::TryFrom;

#[cfg(target_os = "macos")]
use core_graphics::display::CGDisplay;

pub(crate) fn ensure_macos(allow_non_macos: bool) -> Result<(), String> {
    validate_os(env::consts::OS, allow_non_macos)
}

pub(crate) fn is_macos(os: &str) -> bool {
    os == "macos"
}

fn validate_os(os: &str, allow_non_macos: bool) -> Result<(), String> {
    if is_macos(os) || allow_non_macos {
        Ok(())
    } else {
        Err("aerospace-utils only supports macOS.".to_string())
    }
}

pub(crate) fn require_aerospace_executable() -> Result<PathBuf, String> {
    find_aerospace_executable().ok_or_else(|| "aerospace not found in PATH".to_string())
}

fn find_aerospace_executable() -> Option<PathBuf> {
    let path_var = env::var_os("PATH")?;
    let paths = env::split_paths(&path_var).collect::<Vec<_>>();
    find_aerospace_executable_in_paths(&paths)
}

fn find_aerospace_executable_in_paths(paths: &[PathBuf]) -> Option<PathBuf> {
    for path in paths {
        let candidate = path.join("aerospace");
        if is_executable(&candidate) {
            return Some(candidate);
        }
    }

    None
}

fn is_executable(path: &Path) -> bool {
    if !path.is_file() {
        return false;
    }

    let metadata = match fs::metadata(path) {
        Ok(metadata) => metadata,
        Err(_) => return false,
    };

    metadata.permissions().mode() & 0o111 != 0
}

#[cfg(target_os = "macos")]
pub(crate) fn main_display_width() -> Result<i64, String> {
    let width = CGDisplay::main().pixels_wide();
    i64::try_from(width).map_err(|_| "Main display width is too large to fit in i64".to_string())
}

#[cfg(not(target_os = "macos"))]
pub(crate) fn main_display_width() -> Result<i64, String> {
    Err("Unable to detect monitor width on non-macOS.".to_string())
}

pub(crate) fn reload_aerospace_config(aerospace_path: &Path) -> Result<(), String> {
    let output = Command::new(aerospace_path)
        .arg("reload-config")
        .output()
        .map_err(|error| format!("Failed to run aerospace reload-config: {error}"))?;

    if output.status.success() {
        Ok(())
    } else {
        let stderr = String::from_utf8_lossy(&output.stderr);
        let message = stderr.trim();
        if message.is_empty() {
            Err("aerospace reload-config failed".to_string())
        } else {
            Err(format!("aerospace reload-config failed: {message}"))
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[cfg(not(target_os = "macos"))]
    #[test]
    fn main_display_width_requires_override_on_non_macos() {
        let error = main_display_width().unwrap_err();
        assert!(error.contains("non-macOS"));
    }

    #[test]
    fn validate_os_blocks_non_macos_without_override() {
        let error = validate_os("linux", false).unwrap_err();
        assert!(error.contains("macOS"));
    }

    #[test]
    fn validate_os_allows_non_macos_with_override() {
        let result = validate_os("linux", true);
        assert!(result.is_ok());
    }

    #[test]
    fn validate_os_allows_macos_without_override() {
        let result = validate_os("macos", false);
        assert!(result.is_ok());
    }
}
