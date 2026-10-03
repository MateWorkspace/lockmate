use mate_pgdt::sqlx;

use crate::domain::{
    contracts::utility::Logger,
    models::{AppContext, LoggerMeta, RepositoryError},
};

pub(crate) enum OperationError {
    Driver(sqlx::Error),
    Domain(RepositoryError),
}

impl From<sqlx::Error> for OperationError {
    fn from(source: sqlx::Error) -> Self {
        Self::Driver(source)
    }
}

impl From<RepositoryError> for OperationError {
    fn from(error: RepositoryError) -> Self {
        Self::Domain(error)
    }
}

pub(crate) fn finish<T>(
    logger: &dyn Logger,
    context: &AppContext,
    tag: &str,
    result: Result<T, OperationError>,
    not_found: fn() -> RepositoryError,
) -> Result<T, RepositoryError> {
    result.map_err(|error| {
        let mut meta = LoggerMeta::new();
        let error = match error {
            OperationError::Domain(error) => error,
            OperationError::Driver(source) => {
                if let Some(error) = source.as_database_error() {
                    if let Some(code) = error.code() {
                        meta.insert("sql_state".into(), code.into_owned().into());
                    }
                    if let Some(constraint) = error.constraint() {
                        meta.insert("constraint".into(), constraint.into());
                    }
                }
                map_driver(source, not_found)
            }
        };
        meta.insert("error_code".into(), error.code().into());
        logger.error(context, tag, &error.to_string(), &meta);
        error
    })
}

fn map_driver(source: sqlx::Error, not_found: fn() -> RepositoryError) -> RepositoryError {
    if matches!(source, sqlx::Error::RowNotFound) {
        return not_found();
    }
    if matches!(source, sqlx::Error::PoolTimedOut)
        || matches!(&source, sqlx::Error::Io(error) if error.kind() == std::io::ErrorKind::TimedOut)
    {
        return RepositoryError::Timeout {
            source: Box::new(source),
        };
    }
    if let Some(error) = source.as_database_error() {
        match error.code().as_deref() {
            Some("23505") => {
                return match error.constraint() {
                    Some("uq_permissions_name") => RepositoryError::PermissionNameConflict,
                    Some("uq_roles_name") => RepositoryError::RoleNameConflict,
                    Some("uq_users_username") => RepositoryError::UserUsernameConflict,
                    Some("uq_users_email") => RepositoryError::UserEmailConflict,
                    Some("uq_users_phone") => RepositoryError::UserPhoneConflict,
                    Some("uq_role_permission_role_id_permission_id") => {
                        RepositoryError::RolePermissionConflict
                    }
                    _ => RepositoryError::Conflict,
                };
            }
            Some("23503") => return RepositoryError::Conflict,
            Some("22001" | "22P02" | "23502" | "23514") => return RepositoryError::BadArgs,
            _ => {}
        }
    }
    RepositoryError::Failure {
        source: Box::new(source),
    }
}

pub(crate) fn require_affected(
    rows: u64,
    not_found: fn() -> RepositoryError,
) -> Result<(), OperationError> {
    if rows == 0 {
        Err(not_found().into())
    } else {
        Ok(())
    }
}
