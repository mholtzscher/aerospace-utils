use crate::cli::GlobalOptions;
use crate::config::update_config;
use crate::state::{missing_state_file_message, read_state_file, resolve_percentage, write_state};
use crate::system::{main_display_width, reload_aerospace_config, require_aerospace_executable};
use crate::util::{
    calculate_gap_size, resolve_config_path, resolve_state_path, validate_percentage,
};

pub(crate) fn handle_size(
    options: &GlobalOptions,
    percent: Option<i64>,
    set_default: bool,
) -> Result<(), String> {
    let aerospace_path = require_aerospace_executable()?;
    let config_path = resolve_config_path(options)?;
    let state_path = resolve_state_path(options)?;

    let state_load = read_state_file(&state_path, options.dry_run)?;
    let percentage = resolve_percentage(percent, state_load.as_ref())?;

    let Some(percentage) = percentage else {
        println!("No percentage provided and no saved percentage file found");
        return Ok(());
    };

    validate_percentage(percentage)?;
    let monitor_width = main_display_width()?;
    let gap_size = calculate_gap_size(monitor_width, percentage);

    if options.verbose {
        println!("Config path: {}", config_path.display());
        println!("State path: {}", state_path.display());
        println!("Monitor width: {monitor_width}");
        println!("Gap size: {gap_size}");
    }

    if options.dry_run {
        println!(
            "Dry run: would set gaps to {gap_size} in {}",
            config_path.display()
        );
        println!(
            "Dry run: would write current percentage {percentage} to {}",
            state_path.display()
        );
        return Ok(());
    }

    update_config(&config_path, gap_size)?;
    write_state(
        &state_path,
        percentage,
        state_load.as_ref().map(|load| load.state.clone()),
        set_default,
    )?;

    if !options.no_reload
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
    let current = state
        .state
        .current
        .ok_or_else(|| "State file is missing current percentage".to_string())?;

    let new_percentage = current + amount;
    validate_percentage(new_percentage)?;
    println!("Adjusting saved percentage {current} by {amount} to {new_percentage}.");

    handle_size(options, Some(new_percentage), false)
}
