use std::path::PathBuf;

use clap::{Args as ClapArgs, Parser, Subcommand};

#[derive(Parser, Debug)]
#[command(version, about = "Utilities for adjusting Aerospace workspace sizing")]
pub(crate) struct Args {
    #[command(subcommand)]
    pub(crate) command: Commands,
}

#[derive(ClapArgs, Debug)]
#[command(subcommand_required = true, arg_required_else_help = true)]
pub(crate) struct GapsArgs {
    #[command(flatten)]
    pub(crate) options: CommonOptions,
    #[command(subcommand)]
    pub(crate) command: GapsCommands,
}

#[derive(ClapArgs, Debug)]
pub(crate) struct CommonOptions {
    /// Path to aerospace.toml
    #[arg(long, global = true)]
    pub(crate) config_path: Option<PathBuf>,
    /// Path to workspace-size.toml
    #[arg(long, global = true)]
    pub(crate) state_path: Option<PathBuf>,
    /// Skip reload-config
    #[arg(long, global = true)]
    pub(crate) no_reload: bool,
    /// Print actions without writing
    #[arg(long, global = true)]
    pub(crate) dry_run: bool,
    /// Print verbose output
    #[arg(short, long, global = true)]
    pub(crate) verbose: bool,
    /// Disable colored output
    #[arg(long, global = true)]
    pub(crate) no_color: bool,
    /// Override detected monitor width
    #[arg(long, hide = true, value_name = "PX", global = true)]
    pub(crate) monitor_width: Option<i64>,
}

#[derive(Subcommand, Debug)]
pub(crate) enum Commands {
    /// Manage Aerospace workspace gaps
    Gaps(GapsArgs),
}

#[derive(Subcommand, Debug)]
pub(crate) enum GapsCommands {
    /// Set workspace size percentage
    Use {
        /// Workspace percentage
        percent: Option<i64>,
        /// Also set default percentage
        #[arg(long)]
        set_default: bool,
    },
    /// Adjust workspace size by amount
    Adjust {
        /// Amount to adjust (default 5)
        #[arg(default_value_t = 5, allow_hyphen_values = true)]
        amount: i64,
    },
    /// Show resolved config and state
    Current,
}
