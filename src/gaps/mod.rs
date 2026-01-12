pub(crate) mod adjust;
pub(crate) mod config;
pub(crate) mod size;

pub(crate) use adjust::handle_adjust;
pub(crate) use config::handle_config;
pub(crate) use size::handle_size;

pub(crate) fn validate_percentage(percentage: i64) -> Result<(), String> {
    if (1..=100).contains(&percentage) {
        Ok(())
    } else {
        Err(format!(
            "Percentage must be between 1 and 100 (got {percentage})"
        ))
    }
}

pub(crate) fn calculate_gap_size(monitor_width: i64, percentage: i64) -> i64 {
    let workspace_percentage = percentage as f64 / 100.0;
    let gap_percentage = (1.0 - workspace_percentage) / 2.0;
    (monitor_width as f64 * gap_percentage).round() as i64
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn gap_calculation_rounds() {
        let gap_size = calculate_gap_size(1000, 40);
        assert_eq!(gap_size, 300);
    }
}
