use clap::Parser;

mod aerospace;
mod cli;
mod config;
mod display;
mod gaps;
mod output;
mod util;

use crate::cli::{Args, Commands, GapsCommands};
use crate::gaps::{handle_adjust, handle_current, handle_use};

fn main() {
    if let Err(message) = run() {
        eprintln!("{message}");
        std::process::exit(1);
    }
}

fn run() -> Result<(), String> {
    let args = Args::parse();

    match args.command {
        Commands::Gaps(gaps) => match gaps.command {
            GapsCommands::Use {
                percent,
                set_default,
            } => handle_use(&gaps.options, percent, set_default),
            GapsCommands::Adjust { amount } => handle_adjust(&gaps.options, amount),
            GapsCommands::Current => handle_current(&gaps.options),
        },
    }
}
