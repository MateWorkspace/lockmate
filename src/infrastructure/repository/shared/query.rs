use crate::domain::models::RepositoryError;

pub(crate) fn pagination(page: i64, limit: i64) -> Result<(i64, i64), RepositoryError> {
    let limit = limit.max(0);
    let offset = if page <= 1 || limit == 0 {
        0
    } else {
        (page - 1)
            .checked_mul(limit)
            .ok_or(RepositoryError::BadArgs)?
    };
    Ok((limit, offset))
}

pub(crate) fn search_pattern(search: &Option<String>) -> Option<String> {
    search
        .as_ref()
        .filter(|value| !value.is_empty())
        .map(|value| format!("%{value}%"))
}
