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
            let stderr = String::from_utf8_lossy(&output.stderr).trim().to_string();
            let stdout = String::from_utf8_lossy(&output.stdout).trim().to_string();
            let message = if !stderr.is_empty() {
                stderr
            } else if !stdout.is_empty() {
                stdout
            } else {
                format!(
                    "exit code {}",
                    output
                        .status
                        .code()
                        .map(|c| c.to_string())
                        .unwrap_or_else(|| "unknown".to_string())
                )
            };
            Err(format!(
                "aerospace reload-config failed at {}: {message}",
                self.path.display()
            ))
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
    search_executable_in_paths(env::split_paths(&path_var))
}

fn search_executable_in_paths<I>(paths: I) -> Option<PathBuf>
where
    I: IntoIterator<Item = PathBuf>,
{
    paths
        .into_iter()
        .map(|path| path.join("aerospace"))
        .find(|candidate| is_executable(candidate))
}

fn is_executable(path: &Path) -> bool {
    fs::metadata(path)
        .ok()
        .filter(|m| m.is_file())
        .map(|m| m.permissions().mode() & 0o111 != 0)
        .unwrap_or(false)
}

#[cfg(test)]
mod tests {
    use tempfile::tempdir;

    use super::*;

    #[test]
    fn search_executable_in_paths_returns_none_when_missing() {
        let temp_dir = tempdir().expect("tempdir should create");
        let path = temp_dir.path().join("bin");
        fs::create_dir_all(&path).expect("bin dir should create");

        let result = search_executable_in_paths(vec![path]);
        assert!(result.is_none());
    }

    #[test]
    fn search_executable_in_paths_skips_non_executable() {
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

        let result = search_executable_in_paths(vec![first_dir, second_dir]);
        assert_eq!(result, Some(exec));
    }

    #[test]
    fn is_executable_returns_false_for_directory() {
        let temp_dir = tempdir().expect("tempdir should create");
        let dir_path = temp_dir.path().join("subdir");
        fs::create_dir_all(&dir_path).expect("subdir should create");

        assert!(!is_executable(&dir_path));
    }

    #[test]
    fn is_executable_returns_false_for_missing_file() {
        let path = PathBuf::from("/nonexistent/path/aerospace");
        assert!(!is_executable(&path));
    }

    #[test]
    fn is_executable_returns_true_for_executable_file() {
        let temp_dir = tempdir().expect("tempdir should create");
        let exec_path = temp_dir.path().join("test_exec");
        fs::write(&exec_path, b"").expect("file should write");
        let mut perms = fs::metadata(&exec_path)
            .expect("metadata should exist")
            .permissions();
        perms.set_mode(0o755);
        fs::set_permissions(&exec_path, perms).expect("permissions should set");

        assert!(is_executable(&exec_path));
    }

    #[test]
    fn is_executable_returns_false_for_non_executable_file() {
        let temp_dir = tempdir().expect("tempdir should create");
        let non_exec_path = temp_dir.path().join("test_non_exec");
        fs::write(&non_exec_path, b"").expect("file should write");

        assert!(!is_executable(&non_exec_path));
    }
}
