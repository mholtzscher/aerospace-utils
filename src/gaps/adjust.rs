use super::size::handle_use;
use crate::cli::CommonOptions;
use crate::config::WorkspaceState;
use crate::gaps::validate_percentage;

pub(crate) fn handle_adjust(options: &CommonOptions, amount: i64) -> Result<(), String> {
    let state_path = WorkspaceState::resolve_path(options)?;
    let state = WorkspaceState::from_options(options)?
        .ok_or_else(|| WorkspaceState::missing_file_message(&state_path))?;
    let current = state.current().ok_or_else(|| {
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
