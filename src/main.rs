use clap::Parser;

mod cli;
mod config;
mod handlers;
mod state;
mod system;
mod util;

use crate::cli::{Args, Commands};
use crate::handlers::{handle_adjust, handle_size};
use crate::system::ensure_macos;

fn main() {
    if let Err(message) = run() {
        eprintln!("{message}");
        std::process::exit(1);
    }
}

fn run() -> Result<(), String> {
    ensure_macos()?;

    let args = Args::parse();
    match args.command {
        Commands::Size {
            percent,
            set_default,
        } => handle_size(&args.options, percent, set_default),
        Commands::Adjust { amount } => handle_adjust(&args.options, amount),
    }
}
