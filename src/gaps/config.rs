use std::io::IsTerminal;
use std::path::Path;

use colored::Colorize;

use crate::cli::CommonOptions;
use crate::config::MonitorGap;
use crate::config::read_config_summary;
use crate::state::read_state_file;
use crate::util::resolve_config_path;
use crate::util::resolve_state_path;

fn should_colorize(options: &CommonOptions) -> bool {
    std::io::stdout().is_terminal() && !options.no_color && std::env::var_os("NO_COLOR").is_none()
}

fn label_text(label: &str) -> colored::ColoredString {
    label.cyan()
}

fn path_label_text(label: &str) -> colored::ColoredString {
    label.blue()
}

fn path_text(path: &Path) -> colored::ColoredString {
    path.display().to_string().dimmed()
}

fn value_text(value: &str) -> colored::ColoredString {
    value.green()
}

fn unset_text() -> colored::ColoredString {
    "not set".yellow()
}

fn missing_text(message: &str) -> colored::ColoredString {
    message.red()
}

fn print_path(label: &str, path: &Path) {
    println!("{} {}", path_label_text(label), path_text(path));
}

fn print_gap_value(label: &str, value: Option<i64>) {
    match value {
        Some(value) => println!("{}: {}", label_text(label), value_text(&value.to_string())),
        None => println!("{}: {}", label_text(label), unset_text()),
    }
}

fn print_monitor_gaps(label: &str, gaps: &[MonitorGap]) {
    if gaps.is_empty() {
        println!("{}: {}", label_text(label), unset_text());
        return;
    }

    let list = gaps
        .iter()
        .map(|gap| format!("{}={}", gap.name, gap.value))
        .collect::<Vec<_>>()
        .join(", ");
    println!("{}: {}", label_text(label), value_text(&list));
}

pub(crate) fn handle_config(options: &CommonOptions) -> Result<(), String> {
    colored::control::set_override(should_colorize(options));

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
        println!("{}", missing_text("Config file not found."));
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
                    label_text("Current percentage"),
                    value_text(&value)
                ),
                None => println!("{}: {}", label_text("Current percentage"), unset_text()),
            }
            match default_percentage {
                Some(value) => println!(
                    "{}: {}",
                    label_text("Default percentage"),
                    value_text(&value)
                ),
                None => println!("{}: {}", label_text("Default percentage"), unset_text()),
            }
        }
        None => println!("{}", missing_text("State file not found.")),
    }

    Ok(())
}
