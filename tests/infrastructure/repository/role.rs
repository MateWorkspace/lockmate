use super::support::Fixture;
use lockmate::domain::{
    contracts::repository::{CreateRole, RoleFilter, UpdateRole},
    models::RepositoryError,
};
use serde_json::json;

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn role_crud_default_replacement_rollback_and_deletion() {
    let f = Fixture::new().await;
    assert!(matches!(
        f.roles.read_default(&f.context, f.space_id).await,
        Err(RepositoryError::RoleNotFound)
    ));
    let a = f.role("a", true).await;
    let b = f.role("b", false).await;
    assert_eq!(
        f.roles
            .read_by_slug(&f.context, f.space_id, "a")
            .await
            .unwrap()
            .id,
        a
    );
    f.roles
        .update_by_id(
            &f.context,
            f.space_id,
            b,
            UpdateRole {
                is_default: Some(true),
                description: Some("role".into()),
                preferences: Some(json!(null)),
                by: Some(43),
                ..Default::default()
            },
        )
        .await
        .unwrap();
    assert_eq!(
        f.roles
            .read_default(&f.context, f.space_id)
            .await
            .unwrap()
            .id,
        b
    );
    let item = f.roles.read_by_id(&f.context, f.space_id, b).await.unwrap();
    assert_eq!(item.description, "role");
    assert_eq!(item.preferences, json!(null));
    assert_eq!(item.audit.update.by, Some(43));
    assert!(
        !f.roles
            .read_by_id(&f.context, f.space_id, a)
            .await
            .unwrap()
            .is_default
    );
    assert!(matches!(
        f.roles
            .create(
                &f.context,
                f.space_id,
                CreateRole {
                    slug: "a".into(),
                    name: "a".into(),
                    description: None,
                    is_default: Some(true),
                    by: None
                }
            )
            .await,
        Err(RepositoryError::RoleSlugConflict)
    ));
    assert_eq!(
        f.roles
            .read_default(&f.context, f.space_id)
            .await
            .unwrap()
            .id,
        b
    );
    assert!(matches!(
        f.roles
            .update_by_id(
                &f.context,
                f.space_id,
                999999,
                UpdateRole {
                    is_default: Some(true),
                    ..Default::default()
                }
            )
            .await,
        Err(RepositoryError::RoleNotFound)
    ));
    assert_eq!(
        f.roles
            .read_default(&f.context, f.space_id)
            .await
            .unwrap()
            .id,
        b
    );
    f.roles
        .delete_by_id(&f.context, f.space_id, b, Some(44))
        .await
        .unwrap();
    assert!(matches!(
        f.roles.read_by_id(&f.context, f.space_id, b).await,
        Err(RepositoryError::RoleNotFound)
    ));
    assert!(matches!(
        f.roles.delete_by_id(&f.context, f.space_id, b, None).await,
        Err(RepositoryError::RoleNotFound)
    ));
    assert!(matches!(
        f.roles.read_default(&f.context, f.space_id).await,
        Err(RepositoryError::RoleNotFound)
    ));
    let c = f.role("b", true).await;
    assert_eq!(
        f.roles
            .read_default(&f.context, f.space_id)
            .await
            .unwrap()
            .id,
        c
    );
    let (items, total) = f
        .roles
        .read_by_filter(
            &f.context,
            f.space_id,
            RoleFilter {
                page: 1,
                limit: 10,
                is_default: Some(true),
                search: Some("B".into()),
            },
        )
        .await
        .unwrap();
    assert_eq!(total, 1);
    assert_eq!(items[0].id, c);
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn concurrent_default_creation_and_updates_serialize() {
    let f = Fixture::new().await;
    let make = |name: &str| CreateRole {
        slug: name.into(),
        name: name.into(),
        description: None,
        is_default: Some(true),
        by: None,
    };
    let (a, b) = tokio::join!(
        f.roles.create(&f.context, f.space_id, make("a")),
        f.roles.create(&f.context, f.space_id, make("b"))
    );
    let (a, b) = (a.unwrap(), b.unwrap());
    let (_, total) = f
        .roles
        .read_by_filter(
            &f.context,
            f.space_id,
            RoleFilter {
                page: 1,
                limit: 10,
                is_default: Some(true),
                search: None,
            },
        )
        .await
        .unwrap();
    assert_eq!(total, 1);
    let update = || UpdateRole {
        is_default: Some(true),
        ..Default::default()
    };
    let (a, b) = tokio::join!(
        f.roles.update_by_id(&f.context, f.space_id, a, update()),
        f.roles.update_by_id(&f.context, f.space_id, b, update())
    );
    a.unwrap();
    b.unwrap();
    let (_, total) = f
        .roles
        .read_by_filter(
            &f.context,
            f.space_id,
            RoleFilter {
                page: 1,
                limit: 10,
                is_default: Some(true),
                search: None,
            },
        )
        .await
        .unwrap();
    assert_eq!(total, 1);
    assert!(f.logger.0.lock().unwrap().is_empty());
    f.close().await;
}
