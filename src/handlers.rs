use std::path::PathBuf;

use crate::cli::GlobalOptions;
use crate::config::{read_main_gap_sizes, update_config};
use crate::state::{
    StateLoad, missing_state_file_message, read_state_file, resolve_percentage, write_state,
};
use crate::system::{main_display_width, reload_aerospace_config, require_aerospace_executable};
use crate::util::{
    calculate_gap_size, resolve_config_path, resolve_state_path, validate_percentage,
};

enum SizePlanResult {
    MissingPercentage { state_path: PathBuf },
    Ready(SizePlan),
}

struct SizePlan {
    config_path: PathBuf,
    state_path: PathBuf,
    percentage: i64,
    monitor_width: i64,
    gap_size: i64,
    state_load: Option<StateLoad>,
}

fn resolve_monitor_width(options: &GlobalOptions) -> Result<i64, String> {
    match options.monitor_width {
        Some(monitor_width) => {
            if monitor_width <= 0 {
                return Err("Monitor width must be positive.".to_string());
            }
            Ok(monitor_width)
        }
        None => main_display_width().map_err(|error| {
            if options.allow_non_macos {
                format!("{error}\nHint: use --monitor-width to override on non-macOS.")
            } else {
                error
            }
        }),
    }
}

fn build_size_plan(
    options: &GlobalOptions,
    percent: Option<i64>,
) -> Result<SizePlanResult, String> {
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

pub(crate) fn handle_config(options: &GlobalOptions) -> Result<(), String> {
    let config_path = resolve_config_path(options)?;
    let state_path = resolve_state_path(options)?;

    println!("Config path: {}", config_path.display());
    if config_path.exists() {
        let gaps = read_main_gap_sizes(&config_path)?;
        match gaps.left {
            Some(value) => println!("Left gap (monitor.main): {value}"),
            None => println!("Left gap (monitor.main): not set"),
        }
        match gaps.right {
            Some(value) => println!("Right gap (monitor.main): {value}"),
            None => println!("Right gap (monitor.main): not set"),
        }
    } else {
        println!("Config file not found.");
    }

    println!("State path: {}", state_path.display());
    let state_load = read_state_file(&state_path, true)?;
    match state_load {
        Some(load) => {
            let current = load
                .state
                .current
                .map(|value| value.to_string())
                .unwrap_or_else(|| "not set".to_string());
            let default_percentage = load
                .state
                .default_percentage
                .map(|value| value.to_string())
                .unwrap_or_else(|| "not set".to_string());
            println!("Current percentage: {current}");
            println!("Default percentage: {default_percentage}");
        }
        None => println!("State file not found."),
    }

    Ok(())
}

pub(crate) fn handle_size(
    options: &GlobalOptions,
    percent: Option<i64>,
    set_default: bool,
) -> Result<(), String> {
    let aerospace_path = if options.dry_run || options.no_reload {
        None
    } else {
        Some(require_aerospace_executable()?)
    };
    let plan = build_size_plan(options, percent)?;
    let plan = match plan {
        SizePlanResult::MissingPercentage { state_path } => {
            println!(
                "No percentage provided and no saved percentage file found at {}",
                state_path.display()
            );
            return Ok(());
        }
        SizePlanResult::Ready(plan) => plan,
    };

    if options.verbose {
        println!("Config path: {}", plan.config_path.display());
        println!("State path: {}", plan.state_path.display());
        println!("Monitor width: {}", plan.monitor_width);
        println!("Gap size: {}", plan.gap_size);
    }

    if options.dry_run {
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
        return Ok(());
    }

    update_config(&plan.config_path, plan.gap_size)?;
    write_state(
        &plan.state_path,
        plan.percentage,
        plan.state_load.as_ref().map(|load| load.state.clone()),
        set_default,
    )?;

    if let Some(aerospace_path) = aerospace_path
        && let Err(message) = reload_aerospace_config(&aerospace_path)
    {
        eprintln!("Warning: {message}");
        return Ok(());
    }

    println!("Completed.");
    Ok(())
}

pub(crate) fn handle_adjust(options: &GlobalOptions, amount: i64) -> Result<(), String> {
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

    handle_size(options, Some(new_percentage), false)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::cli::GlobalOptions;

    fn default_options() -> GlobalOptions {
        GlobalOptions {
            config_path: None,
            state_path: None,
            no_reload: false,
            dry_run: false,
            verbose: false,
            allow_non_macos: false,
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
