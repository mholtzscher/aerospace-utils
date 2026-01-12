use clap::{Parser, Subcommand};
use std::path::PathBuf;

#[derive(Parser, Debug)]
#[command(version, about = "Utilities for adjusting Aerospace workspace sizing")]
pub(crate) struct Args {
    #[command(flatten)]
    pub(crate) options: GlobalOptions,
    #[command(subcommand)]
    pub(crate) command: Commands,
}

#[derive(Parser, Debug)]
pub(crate) struct GlobalOptions {
    /// Path to aerospace.toml
    #[arg(long)]
    pub(crate) config_path: Option<PathBuf>,
    /// Path to workspace-size.toml
    #[arg(long)]
    pub(crate) state_path: Option<PathBuf>,
    /// Skip reload-config
    #[arg(long)]
    pub(crate) no_reload: bool,
    /// Print actions without writing
    #[arg(long)]
    pub(crate) dry_run: bool,
    /// Print verbose output
    #[arg(short, long)]
    pub(crate) verbose: bool,
}

#[derive(Subcommand, Debug)]
pub(crate) enum Commands {
    /// Set workspace size percentage
    Size {
        /// Workspace percentage
        percent: Option<i64>,
        /// Also set default percentage
        #[arg(long)]
        set_default: bool,
    },
    /// Adjust workspace size by amount
    Adjust {
        /// Amount to adjust (default 5)
        #[arg(default_value_t = 5)]
        amount: i64,
    },
}
