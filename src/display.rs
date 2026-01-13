#[cfg(target_os = "macos")]
use std::convert::TryFrom;

#[cfg(target_os = "macos")]
use core_graphics::display::CGDisplay;

#[cfg(target_os = "macos")]
pub(crate) fn main_display_width() -> Result<i64, String> {
    let width = CGDisplay::main().pixels_wide();
    i64::try_from(width).map_err(|_| "Main display width is too large to fit in i64".to_string())
}

#[cfg(not(target_os = "macos"))]
pub(crate) fn main_display_width() -> Result<i64, String> {
    Err("Unable to detect monitor width on non-macOS.".to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[cfg(not(target_os = "macos"))]
    #[test]
    fn main_display_width_requires_override_on_non_macos() {
        let error = main_display_width().unwrap_err();
        assert!(error.contains("non-macOS"));
    }
}
