use serde_json::json;
use sha2::{Digest, Sha256};

use super::super::shared::{error, query};
use crate::domain::{
    contracts::caching::KeyBuilder,
    models::{AppContext, CacheEntryKey, CacheFamily, CachingError},
};

pub struct Sha256KeyBuilder {
    namespace: String,
}

impl Sha256KeyBuilder {
    pub fn new(namespace: impl Into<String>) -> Result<Self, CachingError> {
        let namespace = namespace.into();
        if namespace.is_empty() || namespace.contains(':') {
            return Err(CachingError::BadArgs);
        }
        Ok(Self { namespace })
    }
}

impl KeyBuilder for Sha256KeyBuilder {
    fn entry(
        &self,
        context: &AppContext,
        specification: &CacheEntryKey,
    ) -> Result<String, CachingError> {
        error::context(context)?;
        let info = query::describe(&specification.query)?;
        let mut revisions: Vec<_> = specification.revisions.iter().collect();
        revisions.sort_by_key(|revision| revision.family);
        if revisions.len() != info.families.len()
            || revisions
                .iter()
                .zip(&info.families)
                .any(|(revision, family)| revision.family != *family || revision.token.is_empty())
        {
            return Err(CachingError::BadArgs);
        }
        let revision_values: Vec<_> = revisions
            .iter()
            .map(|revision| {
                let (scope, entity) = query::family(revision.family)?;
                let id = scope
                    .strip_prefix("space:")
                    .map(str::parse::<i64>)
                    .transpose()
                    .map_err(error::failure)?;
                Ok(json!([entity, id, revision.token]))
            })
            .collect::<Result<_, CachingError>>()?;
        let selector = hex::encode(Sha256::digest(
            serde_json::to_vec(&info.selector).map_err(error::failure)?,
        ));
        let revision = hex::encode(Sha256::digest(
            serde_json::to_vec(&revision_values).map_err(error::failure)?,
        ));
        Ok(format!(
            "{}:cache:v1:{}:{}:{}:{}:{}",
            self.namespace, info.scope, info.entity, info.operation, selector, revision
        ))
    }

    fn revision(&self, context: &AppContext, family: CacheFamily) -> Result<String, CachingError> {
        error::context(context)?;
        let (scope, entity) = query::family(family)?;
        Ok(format!(
            "{}:cache:v1:revision:{}:{}",
            self.namespace, scope, entity
        ))
    }
}
