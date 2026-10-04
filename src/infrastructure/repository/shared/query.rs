use mate_pgdt::sqlx::{Postgres, QueryBuilder};

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

// The same lock serializes default-role replacement and membership joining.
pub(crate) fn lock_default(space_id: i64) -> QueryBuilder<Postgres> {
    let mut query = QueryBuilder::new(
        "SELECT pg_advisory_xact_lock(hashtextextended('lockmate/default-role/' || ",
    );
    query.push_bind(space_id).push("::bigint::text, 0))");
    query
}
