use crate::cli::CommonOptions;
use crate::gaps::validate_percentage;
use crate::state::{missing_state_file_message, read_state_file};
use crate::util::resolve_state_path;

use super::size::handle_use;

pub(crate) fn handle_adjust(options: &CommonOptions, amount: i64) -> Result<(), String> {
    let state_path = resolve_state_path(options)?;
    let state_load = read_state_file(&state_path, options.dry_run)?;
    let state = state_load
        .as_ref()
        .ok_or_else(|| missing_state_file_message(&state_path))?;
    let current = state.state.current.ok_or_else(|| {
        format!(
            "State file at {} is missing current percentage",
            state_path.display()
        )
    })?;

    let new_percentage = current + amount;
    validate_percentage(new_percentage)?;
    println!("Adjusting saved percentage {current} by {amount} to {new_percentage}.");

    handle_use(options, Some(new_percentage), false)
}
