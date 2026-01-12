use std::env;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::Command;

pub(crate) fn ensure_macos() -> Result<(), String> {
    if env::consts::OS == "macos" {
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

pub(crate) fn main_display_width() -> Result<i64, String> {
    let output = Command::new("system_profiler")
        .arg("SPDisplaysDataType")
        .output()
        .map_err(|error| format!("Failed to run system_profiler: {error}"))?;

    if !output.status.success() {
        return Err("system_profiler failed to run".to_string());
    }

    let stdout = String::from_utf8_lossy(&output.stdout);
    parse_main_display_width(&stdout)
}

fn parse_main_display_width(output: &str) -> Result<i64, String> {
    let lines = output.lines().collect::<Vec<_>>();
    let mut resolution_line = None;

    for (index, line) in lines.iter().enumerate() {
        if line.contains("Main Display: Yes") {
            for candidate in lines[..=index].iter().rev() {
                if candidate.contains("Resolution:") {
                    resolution_line = Some(*candidate);
                    break;
                }
            }
            break;
        }
    }

    let resolution_line = resolution_line.ok_or_else(|| {
        "Unable to locate main display resolution in system_profiler output".to_string()
    })?;

    let after_label = resolution_line
        .split_once("Resolution:")
        .map(|(_, value)| value)
        .ok_or_else(|| "Malformed resolution line".to_string())?
        .trim();

    let width_part = after_label
        .split('x')
        .next()
        .ok_or_else(|| "Malformed resolution line".to_string())?
        .trim();

    width_part
        .parse::<i64>()
        .map_err(|error| format!("Failed to parse monitor width: {error}"))
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

    #[test]
    fn parse_system_profiler_output_extracts_width() {
        let output = r#"
Graphics/Displays:

    Display:
      Resolution: 3456 x 2234 Retina
      Main Display: Yes
"#;

        let width = parse_main_display_width(output).unwrap();
        assert_eq!(width, 3456);
    }
}
