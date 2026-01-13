use std::path::Path;

use crate::cli::CommonOptions;
use crate::config::{AerospaceConfig, MonitorGap, WorkspaceState};
use crate::output;

fn print_path(label: &str, path: &Path) {
    println!("{} {}", output::path_label(label), output::path(path));
}

fn print_gap_value(label: &str, value: Option<i64>) {
    match value {
        Some(value) => println!("{}: {}", output::label(label), output::value(value)),
        None => println!("{}: {}", output::label(label), output::unset()),
    }
}

fn print_monitor_gaps(label: &str, gaps: &[MonitorGap]) {
    if gaps.is_empty() {
        println!("{}: {}", output::label(label), output::unset());
        return;
    }

    let list = gaps
        .iter()
        .map(|gap| format!("{}={}", gap.name, gap.value))
        .collect::<Vec<_>>()
        .join(", ");
    println!("{}: {}", output::label(label), output::value(list));
}

pub(crate) fn handle_current(options: &CommonOptions) -> Result<(), String> {
    output::configure(options);

    let config_path = AerospaceConfig::resolve_path(options)?;
    let state_path = WorkspaceState::resolve_path(options)?;

    print_path("Config path:", &config_path);
    if config_path.exists() {
        let config = AerospaceConfig::load(config_path)?;
        let summary = config.summary()?;
        print_gap_value("Inner gap horizontal", summary.inner_horizontal);
        print_gap_value("Inner gap vertical", summary.inner_vertical);
        print_gap_value("Outer gap top", summary.outer_top);
        print_gap_value("Outer gap bottom", summary.outer_bottom);
        print_monitor_gaps("Left gaps", &summary.left_gaps);
        print_monitor_gaps("Right gaps", &summary.right_gaps);
    } else {
        println!("{}", output::missing("Config file not found."));
    }

    print_path("State path:", &state_path);
    // Pass dry_run=true to avoid writing during a read-only operation
    let state = WorkspaceState::from_options(&CommonOptions {
        dry_run: true,
        ..options.clone()
    })?;
    match state {
        Some(state) => {
            let current = state.current().map(|value| value.to_string());
            let default_percentage = state.default_percentage().map(|value| value.to_string());
            match current {
                Some(value) => println!(
                    "{}: {}",
                    output::label("Current percentage"),
                    output::value(value)
                ),
                None => println!(
                    "{}: {}",
                    output::label("Current percentage"),
                    output::unset()
                ),
            }
            match default_percentage {
                Some(value) => println!(
                    "{}: {}",
                    output::label("Default percentage"),
                    output::value(value)
                ),
                None => println!(
                    "{}: {}",
                    output::label("Default percentage"),
                    output::unset()
                ),
            }
        }
        None => println!("{}", output::missing("State file not found.")),
    }

    Ok(())
}
