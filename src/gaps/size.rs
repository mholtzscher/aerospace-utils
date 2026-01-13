use std::path::PathBuf;

use colored::ColoredString;

use crate::aerospace::resolve_binary;
use crate::cli::CommonOptions;
use crate::config::{AerospaceConfig, WorkspaceState};
use crate::display::main_display_width;
use crate::gaps::{calculate_gap_size, validate_percentage};
use crate::output;

enum SizePlanResult {
    MissingPercentage { state_path: PathBuf },
    Ready(Box<SizePlan>),
}

struct SizePlan {
    config: AerospaceConfig,
    state: WorkspaceState,
    percentage: i64,
    monitor_width: i64,
    gap_size: i64,
}

enum ReloadStatus {
    Skipped,
    Ok,
    Failed,
}

fn reload_status_text(status: ReloadStatus) -> ColoredString {
    match status {
        ReloadStatus::Skipped => output::reload_skipped(),
        ReloadStatus::Ok => output::reload_ok(),
        ReloadStatus::Failed => output::reload_failed(),
    }
}

fn resolve_monitor_width(options: &CommonOptions) -> Result<i64, String> {
    match options.monitor_width {
        Some(monitor_width) => {
            if monitor_width <= 0 {
                return Err("Monitor width must be positive.".to_string());
            }
            Ok(monitor_width)
        }
        None => main_display_width()
            .map_err(|error| format!("{error}\nHint: use --monitor-width to override detection.")),
    }
}

fn build_plan(
    options: &CommonOptions,
    percent: Option<i64>,
    existing_state: Option<WorkspaceState>,
) -> Result<SizePlanResult, String> {
    let config = AerospaceConfig::from_options(options)?;
    let state_path = WorkspaceState::resolve_path(options)?;

    let percentage = match &existing_state {
        Some(state) => state.resolve_percentage(percent),
        None => percent,
    };

    let Some(percentage) = percentage else {
        return Ok(SizePlanResult::MissingPercentage { state_path });
    };

    validate_percentage(percentage)?;
    let monitor_width = resolve_monitor_width(options)?;
    let gap_size = calculate_gap_size(monitor_width, percentage);

    let state = match existing_state {
        Some(state) => state,
        None => WorkspaceState::new(state_path, percentage, None),
    };

    Ok(SizePlanResult::Ready(Box::new(SizePlan {
        config,
        state,
        percentage,
        monitor_width,
        gap_size,
    })))
}

fn handle_dry_run(plan: &SizePlan) {
    println!(
        "Dry run: would set gaps to {} in {}",
        plan.gap_size,
        plan.config.path().display()
    );
    println!(
        "Dry run: would write current percentage {} to {}",
        plan.percentage,
        plan.state.path().display()
    );
}

fn execute_plan(
    mut plan: SizePlan,
    set_default: bool,
    verbose: bool,
    dry_run: bool,
    no_reload: bool,
) -> Result<(), String> {
    if verbose {
        println!("Config path: {}", plan.config.path().display());
        println!("State path: {}", plan.state.path().display());
        println!("Monitor width: {}", plan.monitor_width);
        println!("Gap size: {}", plan.gap_size);
    }

    if dry_run {
        handle_dry_run(&plan);
        return Ok(());
    }

    plan.config.set_main_gaps(plan.gap_size)?;
    plan.config.write()?;

    plan.state.update(plan.percentage, set_default);
    plan.state.write()?;

    let reload_status = if no_reload {
        ReloadStatus::Skipped
    } else {
        let aerospace = resolve_binary()?;
        match aerospace.reload() {
            Ok(()) => ReloadStatus::Ok,
            Err(message) => {
                eprintln!("Warning: {message}");
                ReloadStatus::Failed
            }
        }
    };

    println!(
        "Gaps set to {} (from {} of {}) ({}).",
        output::value(plan.gap_size),
        output::value(format!("{}%", plan.percentage)),
        output::value(format!("{}px", plan.monitor_width)),
        reload_status_text(reload_status)
    );
    Ok(())
}

pub(crate) fn handle_use(
    options: &CommonOptions,
    percent: Option<i64>,
    set_default: bool,
) -> Result<(), String> {
    output::configure(options);

    let existing_state = WorkspaceState::from_options(options)?;
    let plan = match build_plan(options, percent, existing_state)? {
        SizePlanResult::Ready(plan) => *plan,
        SizePlanResult::MissingPercentage { state_path } => {
            return Err(format!(
                "No percentage provided and no saved percentage found at {}",
                state_path.display()
            ));
        }
    };

    execute_plan(
        plan,
        set_default,
        options.verbose,
        options.dry_run,
        options.no_reload,
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    fn default_options() -> CommonOptions {
        CommonOptions {
            config_path: None,
            state_path: None,
            no_reload: false,
            dry_run: false,
            verbose: false,
            no_color: false,
            monitor_width: None,
        }
    }

    #[test]
    fn resolve_monitor_width_uses_override() {
        let mut options = default_options();
        options.monitor_width = Some(1800);

        let monitor_width = resolve_monitor_width(&options).unwrap();
        assert_eq!(monitor_width, 1800);
    }

    #[test]
    fn resolve_monitor_width_rejects_non_positive() {
        let mut options = default_options();
        options.monitor_width = Some(0);

        let error = resolve_monitor_width(&options).unwrap_err();
        assert!(error.contains("positive"));
    }

    #[test]
    fn resolve_monitor_width_rejects_negative() {
        let mut options = default_options();
        options.monitor_width = Some(-1200);

        let error = resolve_monitor_width(&options).unwrap_err();
        assert!(error.contains("positive"));
    }
}
