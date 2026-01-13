use std::path::PathBuf;

use colored::ColoredString;

use crate::cli::CommonOptions;
use crate::config::update_config;
use crate::gaps::{calculate_gap_size, validate_percentage};
use crate::output;
use crate::state::{StateLoad, read_state_file, resolve_percentage, write_state};
use crate::system::{main_display_width, reload_aerospace_config, require_aerospace_executable};
use crate::util::{resolve_config_path, resolve_state_path};

enum SizePlanResult {
    MissingPercentage { state_path: PathBuf },
    Ready(SizePlan),
}

#[derive(Debug)]
pub(crate) struct SizePlan {
    pub(crate) config_path: PathBuf,
    pub(crate) state_path: PathBuf,
    pub(crate) percentage: i64,
    pub(crate) monitor_width: i64,
    pub(crate) gap_size: i64,
    pub(crate) state_load: Option<StateLoad>,
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

fn resolve_aerospace_path(dry_run: bool, no_reload: bool) -> Result<Option<PathBuf>, String> {
    if dry_run || no_reload {
        Ok(None)
    } else {
        require_aerospace_executable().map(Some)
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

fn build_use_plan(options: &CommonOptions, percent: Option<i64>) -> Result<SizePlanResult, String> {
    let config_path = resolve_config_path(options)?;
    let state_path = resolve_state_path(options)?;
    let state_load = read_state_file(&state_path, options.dry_run)?;
    let percentage = resolve_percentage(percent, state_load.as_ref())?;

    let Some(percentage) = percentage else {
        return Ok(SizePlanResult::MissingPercentage { state_path });
    };

    validate_percentage(percentage)?;
    let monitor_width = resolve_monitor_width(options)?;
    let gap_size = calculate_gap_size(monitor_width, percentage);

    Ok(SizePlanResult::Ready(SizePlan {
        config_path,
        state_path,
        percentage,
        monitor_width,
        gap_size,
        state_load,
    }))
}

fn handle_dry_run(plan: &SizePlan) {
    println!(
        "Dry run: would set gaps to {} in {}",
        plan.gap_size,
        plan.config_path.display()
    );
    println!(
        "Dry run: would write current percentage {} to {}",
        plan.percentage,
        plan.state_path.display()
    );
}

fn execute_plan(
    plan: &SizePlan,
    set_default: bool,
    verbose: bool,
    dry_run: bool,
    no_reload: bool,
) -> Result<(), String> {
    if verbose {
        println!("Config path: {}", plan.config_path.display());
        println!("State path: {}", plan.state_path.display());
        println!("Monitor width: {}", plan.monitor_width);
        println!("Gap size: {}", plan.gap_size);
    }

    if dry_run {
        handle_dry_run(plan);
        return Ok(());
    }

    update_config(&plan.config_path, plan.gap_size)?;
    write_state(
        &plan.state_path,
        plan.percentage,
        plan.state_load.as_ref().map(|load| load.state.clone()),
        set_default,
    )?;

    let aerospace_path = resolve_aerospace_path(dry_run, no_reload)?;
    let reload_status = match aerospace_path {
        None => ReloadStatus::Skipped,
        Some(path) => match reload_aerospace_config(&path) {
            Ok(()) => ReloadStatus::Ok,
            Err(message) => {
                eprintln!("Warning: {message}");
                ReloadStatus::Failed
            }
        },
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

    let plan = match build_use_plan(options, percent)? {
        SizePlanResult::Ready(plan) => plan,
        SizePlanResult::MissingPercentage { state_path } => {
            println!(
                "No percentage provided and no saved percentage found at {}",
                state_path.display()
            );
            return Ok(());
        }
    };

    execute_plan(
        &plan,
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
}
