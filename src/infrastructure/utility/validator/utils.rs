use std::ops::RangeInclusive;

use regex::Regex;

use crate::domain::models::ValidatorError;

pub(super) fn validate_length(
    value: &str,
    bounds: RangeInclusive<usize>,
    too_short: ValidatorError,
    too_long: ValidatorError,
) -> Result<(), ValidatorError> {
    let length = value.chars().count();
    if length < *bounds.start() {
        return Err(too_short);
    }
    if length > *bounds.end() {
        return Err(too_long);
    }
    Ok(())
}

pub(super) fn validate_pattern(
    value: &str,
    bounds: RangeInclusive<usize>,
    pattern: &Regex,
    errors: (ValidatorError, ValidatorError, ValidatorError),
) -> Result<(), ValidatorError> {
    let (invalid, too_short, too_long) = errors;
    validate_length(value, bounds, too_short, too_long)?;
    if value.trim() != value || !pattern.is_match(value) {
        return Err(invalid);
    }
    Ok(())
}
