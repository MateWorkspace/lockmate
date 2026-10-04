use crate::domain::{
    contracts::utility::Logger,
    models::{AppContext, CachingError, LoggerMeta},
};

pub(crate) fn failure(source: impl std::error::Error + Send + Sync + 'static) -> CachingError {
    CachingError::Failure {
        source: Box::new(source),
    }
}

pub(crate) fn redis(source: redis::RedisError) -> CachingError {
    if source.is_timeout() {
        CachingError::Timeout {
            source: Box::new(source),
        }
    } else {
        failure(source)
    }
}

pub(crate) fn invalid_payload() -> CachingError {
    failure(std::io::Error::other("invalid cached payload"))
}

pub(crate) fn finish<T>(
    logger: &dyn Logger,
    context: &AppContext,
    tag: &str,
    result: Result<T, CachingError>,
) -> Result<T, CachingError> {
    result.inspect_err(|error| {
        let meta = LoggerMeta::from([("error_code".into(), error.code().into())]);
        match error {
            CachingError::BadArgs | CachingError::BadState => {
                logger.error(context, tag, &error.to_string(), &meta)
            }
            _ => logger.warn(context, tag, &error.to_string(), &meta),
        }
    })
}

pub(crate) fn context(context: &AppContext) -> Result<(), CachingError> {
    if context.transaction.is_some() {
        return Err(CachingError::BadState);
    }
    Ok(())
}
