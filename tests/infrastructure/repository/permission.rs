use super::support::Fixture;
use lockmate::domain::{
    contracts::repository::{CreatePermission, PermissionFilter, UpdatePermission},
    models::{LoggerLevel, RepositoryError},
};
use mate_pgdt::sqlx;
use serde_json::json;

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn permission_crud_defaults_audits_and_partial_uniqueness() {
    let f = Fixture::new().await;
    let id = f.permission("permission").await;
    let item = f
        .permissions
        .read_by_id(&f.context, f.space_id, id)
        .await
        .unwrap();
    assert_eq!(item.description, "");
    assert_eq!(item.preferences, json!({}));
    assert_eq!(item.audit.create.by, Some(42));
    assert!(item.audit.update.at.is_none());
    assert_eq!(
        f.permissions
            .read_by_slug(&f.context, f.space_id, "permission")
            .await
            .unwrap()
            .id,
        id
    );
    f.permissions
        .update_by_id(
            &f.context,
            f.space_id,
            id,
            UpdatePermission {
                description: Some("description".into()),
                preferences: Some(json!({"nested":[1,true,null]})),
                by: Some(43),
                ..Default::default()
            },
        )
        .await
        .unwrap();
    let item = f
        .permissions
        .read_by_id(&f.context, f.space_id, id)
        .await
        .unwrap();
    assert_eq!(item.description, "description");
    assert_eq!(item.audit.update.by, Some(43));
    assert!(item.audit.update.at.is_some());
    assert_eq!(item.preferences, json!({"nested":[1,true,null]}));
    let err = f
        .permissions
        .create(
            &f.context,
            f.space_id,
            CreatePermission {
                slug: "permission".into(),
                name: "permission".into(),
                description: None,
                by: None,
            },
        )
        .await
        .unwrap_err();
    assert!(matches!(err, RepositoryError::PermissionSlugConflict));
    f.permissions
        .delete_by_id(&f.context, f.space_id, id, Some(44))
        .await
        .unwrap();
    assert!(matches!(
        f.permissions.read_by_id(&f.context, f.space_id, id).await,
        Err(RepositoryError::PermissionNotFound)
    ));
    assert!(matches!(
        f.permissions
            .read_by_slug(&f.context, f.space_id, "permission")
            .await,
        Err(RepositoryError::PermissionNotFound)
    ));
    assert!(matches!(
        f.permissions
            .update_by_id(&f.context, f.space_id, id, UpdatePermission::default())
            .await,
        Err(RepositoryError::PermissionNotFound)
    ));
    assert!(matches!(
        f.permissions
            .delete_by_id(&f.context, f.space_id, id, None)
            .await,
        Err(RepositoryError::PermissionNotFound)
    ));
    let row =
        f.db.query_row(
            &f.context,
            sqlx::query("SELECT deleted_by FROM permissions WHERE id=$1").bind(id),
        )
        .await
        .unwrap();
    use sqlx::Row;
    assert_eq!(row.try_get::<i64, _>("deleted_by").unwrap(), 44);
    assert_ne!(f.permission("permission").await, id);
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn permission_pagination_binding_and_failure_logs() {
    let f = Fixture::new().await;
    let hostile = "quoted' OR 1=1 --";
    let id = f.permission(hostile).await;
    f.permission("another").await;
    let (items, total) = f
        .permissions
        .read_by_filter(
            &f.context,
            f.space_id,
            PermissionFilter {
                page: 1,
                limit: 10,
                search: Some(hostile.into()),
            },
        )
        .await
        .unwrap();
    assert_eq!(total, 1);
    assert_eq!(items[0].id, id);
    let (items, total) = f
        .permissions
        .read_by_filter(
            &f.context,
            f.space_id,
            PermissionFilter {
                page: -5,
                limit: 0,
                search: None,
            },
        )
        .await
        .unwrap();
    assert!(items.is_empty());
    assert_eq!(total, 2);
    let (items, total) = f
        .permissions
        .read_by_filter(
            &f.context,
            f.space_id,
            PermissionFilter {
                page: 1,
                limit: -1,
                search: None,
            },
        )
        .await
        .unwrap();
    assert!(items.is_empty());
    assert_eq!(total, 2);
    let (first, total) = f
        .permissions
        .read_by_filter(
            &f.context,
            f.space_id,
            PermissionFilter {
                page: 1,
                limit: 1,
                search: Some("".into()),
            },
        )
        .await
        .unwrap();
    assert_eq!(total, 2);
    let (second, _) = f
        .permissions
        .read_by_filter(
            &f.context,
            f.space_id,
            PermissionFilter {
                page: 2,
                limit: 1,
                search: None,
            },
        )
        .await
        .unwrap();
    assert_ne!(first[0].id, second[0].id);
    let (items, total) = f
        .permissions
        .read_by_filter(
            &f.context,
            f.space_id,
            PermissionFilter {
                page: 100,
                limit: 10,
                search: None,
            },
        )
        .await
        .unwrap();
    assert!(items.is_empty());
    assert_eq!(total, 2);
    let (items, total) = f
        .permissions
        .read_by_filter(
            &f.context,
            f.space_id,
            PermissionFilter {
                page: 1,
                limit: 10,
                search: Some("missing".into()),
            },
        )
        .await
        .unwrap();
    assert!(items.is_empty());
    assert_eq!(total, 0);
    assert!(matches!(
        f.permissions
            .read_by_filter(
                &f.context,
                f.space_id,
                PermissionFilter {
                    page: i64::MAX,
                    limit: 2,
                    search: None
                }
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    let entries = f.logger.0.lock().unwrap().clone();
    assert_eq!(entries.len(), 1);
    assert_eq!(entries[0].context, f.context);
    assert_eq!(entries[0].level, LoggerLevel::Error);
    assert_eq!(
        entries[0].tag,
        "repository/permission/postgres/read_by_filter"
    );
    assert_eq!(entries[0].meta["error_code"], "BAD_ARGS".into());
    assert!(!entries[0].message.contains(hostile));
    f.close().await;
}
