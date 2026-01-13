use std::io::IsTerminal;
use std::path::Path;

use colored::Colorize;

use crate::cli::CommonOptions;

/// Configures the colored output based on terminal detection and options.
pub(crate) fn configure(options: &CommonOptions) {
    colored::control::set_override(should_colorize(options));
}

fn should_colorize(options: &CommonOptions) -> bool {
    std::io::stdout().is_terminal() && !options.no_color && std::env::var_os("NO_COLOR").is_none()
}

pub(crate) fn value<T: ToString>(v: T) -> colored::ColoredString {
    v.to_string().green()
}

pub(crate) fn label(s: &str) -> colored::ColoredString {
    s.cyan()
}

pub(crate) fn path_label(s: &str) -> colored::ColoredString {
    s.blue()
}

pub(crate) fn path(p: &Path) -> colored::ColoredString {
    p.display().to_string().dimmed()
}

pub(crate) fn unset() -> colored::ColoredString {
    "not set".yellow()
}

pub(crate) fn missing(message: &str) -> colored::ColoredString {
    message.red()
}

pub(crate) fn reload_skipped() -> colored::ColoredString {
    "reload skipped".yellow()
}

pub(crate) fn reload_ok() -> colored::ColoredString {
    "reload ok".green()
}

pub(crate) fn reload_failed() -> colored::ColoredString {
    "reload failed".red()
}
