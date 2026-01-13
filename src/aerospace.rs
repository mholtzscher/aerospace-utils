use std::env;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::Command;

pub(crate) struct AerospaceBinary {
    path: PathBuf,
}

impl AerospaceBinary {
    pub(crate) fn reload(&self) -> Result<(), String> {
        let output = Command::new(&self.path)
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
}

pub(crate) fn resolve_binary() -> Result<AerospaceBinary, String> {
    find_aerospace_executable()
        .map(|path| AerospaceBinary { path })
        .ok_or_else(|| "aerospace not found in PATH".to_string())
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

#[cfg(test)]
mod tests {
    use tempfile::tempdir;

    use super::*;

    #[test]
    fn find_aerospace_executable_in_paths_returns_none_when_missing() {
        let temp_dir = tempdir().expect("tempdir should create");
        let path = temp_dir.path().join("bin");
        fs::create_dir_all(&path).expect("bin dir should create");

        let result = find_aerospace_executable_in_paths(&[path]);
        assert!(result.is_none());
    }

    #[test]
    fn find_aerospace_executable_in_paths_skips_non_executable() {
        let temp_dir = tempdir().expect("tempdir should create");
        let first_dir = temp_dir.path().join("first");
        let second_dir = temp_dir.path().join("second");
        fs::create_dir_all(&first_dir).expect("first dir should create");
        fs::create_dir_all(&second_dir).expect("second dir should create");

        let non_exec = first_dir.join("aerospace");
        fs::write(&non_exec, b"").expect("non-exec should write");

        let exec = second_dir.join("aerospace");
        fs::write(&exec, b"").expect("exec should write");
        let mut perms = fs::metadata(&exec)
            .expect("exec metadata should exist")
            .permissions();
        perms.set_mode(0o755);
        fs::set_permissions(&exec, perms).expect("exec permissions should set");

        let result = find_aerospace_executable_in_paths(&[first_dir, second_dir]);
        assert_eq!(result, Some(exec));
    }
}
