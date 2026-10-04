use super::support::Fixture;
use lockmate::domain::{
    contracts::repository::{ApiKeyFilter, UpdateApiKey},
    models::RepositoryError,
};
use serde_json::json;

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn api_key_crud_filters_hash_conflicts_and_defaults() {
    let f = Fixture::new().await;
    f.role("user", true).await;
    let user = f.user("user").await;
    let member = f.member(f.space_id, user).await;
    let id = f
        .api_keys
        .create(
            &f.context,
            f.space_id,
            Fixture::key_input(member, "key", "private-key-hash"),
        )
        .await
        .unwrap();
    let item = f
        .api_keys
        .read_by_id(&f.context, f.space_id, id)
        .await
        .unwrap();
    assert_eq!(item.member_id, member);
    assert_eq!(item.space_id, f.space_id);
    assert_eq!(item.description, "");
    assert_eq!(item.hash, "private-key-hash");
    assert_eq!(item.preferences, json!({}));
    assert_eq!(item.audit.create.by, Some(42));
    assert_eq!(
        f.api_keys
            .read_by_hash(&f.context, f.space_id, "private-key-hash")
            .await
            .unwrap()
            .id,
        id
    );
    assert!(matches!(
        f.api_keys
            .create(
                &f.context,
                f.space_id,
                Fixture::key_input(member, "duplicate", "private-key-hash")
            )
            .await,
        Err(RepositoryError::Conflict)
    ));
    assert!(matches!(
        f.api_keys
            .create(
                &f.context,
                f.space_id,
                Fixture::key_input(999999, "invalid", "another-hash")
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    f.api_keys
        .update_by_id(
            &f.context,
            f.space_id,
            id,
            UpdateApiKey {
                name: Some("changed".into()),
                description: Some("description".into()),
                preferences: Some(json!([1, "two"])),
                by: Some(43),
            },
        )
        .await
        .unwrap();
    let (items, total) = f
        .api_keys
        .read_by_filter(
            &f.context,
            f.space_id,
            ApiKeyFilter {
                page: 1,
                limit: 10,
                search: Some("ABCD".into()),
                member_id: Some(member),
                user_id: Some(user),
            },
        )
        .await
        .unwrap();
    assert_eq!(total, 1);
    assert_eq!(items[0].name, "changed");
    assert_eq!(items[0].preferences, json!([1, "two"]));
    assert_eq!(items[0].audit.update.by, Some(43));
    f.api_keys
        .delete_by_id(&f.context, f.space_id, id, Some(44))
        .await
        .unwrap();
    assert!(matches!(
        f.api_keys.read_by_id(&f.context, f.space_id, id).await,
        Err(RepositoryError::ApiKeyNotFound)
    ));
    assert!(matches!(
        f.api_keys
            .read_by_hash(&f.context, f.space_id, "private-key-hash")
            .await,
        Err(RepositoryError::ApiKeyNotFound)
    ));
    assert!(matches!(
        f.api_keys
            .update_by_id(&f.context, f.space_id, id, UpdateApiKey::default())
            .await,
        Err(RepositoryError::ApiKeyNotFound)
    ));
    assert!(matches!(
        f.api_keys
            .delete_by_id(&f.context, f.space_id, id, None)
            .await,
        Err(RepositoryError::ApiKeyNotFound)
    ));
    assert!(
        f.api_keys
            .create(
                &f.context,
                f.space_id,
                Fixture::key_input(member, "reuse", "private-key-hash")
            )
            .await
            .is_ok()
    );
    for entry in f.logger.0.lock().unwrap().iter() {
        assert!(!entry.message.contains("private-key-hash"));
        assert!(!format!("{:?}", entry.meta).contains("private-key-hash"));
    }
    f.close().await;
}
