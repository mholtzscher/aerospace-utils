use crate::cli::CommonOptions;
use crate::config::MonitorGap;
use crate::config::read_config_summary;
use crate::state::read_state_file;
use crate::util::resolve_config_path;
use crate::util::resolve_state_path;

fn print_gap_value(label: &str, value: Option<i64>) {
    match value {
        Some(value) => println!("{label}: {value}"),
        None => println!("{label}: not set"),
    }
}

fn print_monitor_gaps(label: &str, gaps: &[MonitorGap]) {
    if gaps.is_empty() {
        println!("{label}: not set");
        return;
    }

    let list = gaps
        .iter()
        .map(|gap| format!("{}={}", gap.name, gap.value))
        .collect::<Vec<_>>()
        .join(", ");
    println!("{label}: {list}");
}

pub(crate) fn handle_config(options: &CommonOptions) -> Result<(), String> {
    let config_path = resolve_config_path(options)?;
    let state_path = resolve_state_path(options)?;

    println!("Config path: {}", config_path.display());
    if config_path.exists() {
        let summary = read_config_summary(&config_path)?;
        print_gap_value("Inner gap horizontal", summary.inner_horizontal);
        print_gap_value("Inner gap vertical", summary.inner_vertical);
        print_gap_value("Outer gap top", summary.outer_top);
        print_gap_value("Outer gap bottom", summary.outer_bottom);
        print_monitor_gaps("Left gaps", &summary.left_gaps);
        print_monitor_gaps("Right gaps", &summary.right_gaps);
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
