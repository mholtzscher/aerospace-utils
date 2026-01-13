use std::path::Path;

use crate::cli::CommonOptions;
use crate::config::MonitorGap;
use crate::config::read_config_summary;
use crate::output;
use crate::state::read_state_file;
use crate::util::resolve_config_path;
use crate::util::resolve_state_path;

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

    let config_path = resolve_config_path(options)?;
    let state_path = resolve_state_path(options)?;

    print_path("Config path:", &config_path);
    if config_path.exists() {
        let summary = read_config_summary(&config_path)?;
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
    let state_load = read_state_file(&state_path, true)?;
    match state_load {
        Some(load) => {
            let current = load.state.current.map(|value| value.to_string());
            let default_percentage = load.state.default_percentage.map(|value| value.to_string());
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
