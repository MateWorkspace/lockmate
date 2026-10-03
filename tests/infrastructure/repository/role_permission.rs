use super::support::Fixture;
use lockmate::domain::{contracts::repository::CreateRolePermission, models::RepositoryError};

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn role_permission_joined_reads_conflicts_and_active_visibility() {
    let f = Fixture::new().await;
    let r = f.role("role", false).await;
    let p = f.permission("permission").await;
    let id = f.pivot(r, p).await;
    let item = f.role_permissions.read_by_id(&f.context, id).await.unwrap();
    assert_eq!(item.role_permission.id, id);
    assert_eq!(item.role.id, r);
    assert_eq!(item.permission.id, p);
    assert_eq!(item.role_permission.audit.by, Some(42));
    let items = f
        .role_permissions
        .read_by_role_id(&f.context, r)
        .await
        .unwrap();
    assert_eq!(items.len(), 1);
    assert_eq!(items[0].permission.id, p);
    let items = f
        .role_permissions
        .read_by_permission_id(&f.context, p)
        .await
        .unwrap();
    assert_eq!(items.len(), 1);
    assert_eq!(items[0].role.id, r);
    assert!(matches!(
        f.role_permissions
            .create(
                &f.context,
                CreateRolePermission {
                    role_id: r,
                    permission_id: p,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::RolePermissionConflict)
    ));
    assert!(matches!(
        f.role_permissions
            .create(
                &f.context,
                CreateRolePermission {
                    role_id: 999999,
                    permission_id: p,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::Conflict)
    ));
    f.permissions
        .delete_by_id(&f.context, p, None)
        .await
        .unwrap();
    assert!(matches!(
        f.role_permissions.read_by_id(&f.context, id).await,
        Err(RepositoryError::RolePermissionNotFound)
    ));
    assert!(
        f.role_permissions
            .read_by_role_id(&f.context, r)
            .await
            .unwrap()
            .is_empty()
    );
    f.role_permissions
        .delete_by_id(&f.context, id)
        .await
        .unwrap();
    assert!(matches!(
        f.role_permissions.delete_by_id(&f.context, id).await,
        Err(RepositoryError::RolePermissionNotFound)
    ));
    let p = f.permission("permission").await;
    let id = f.pivot(r, p).await;
    f.roles.delete_by_id(&f.context, r, None).await.unwrap();
    assert!(matches!(
        f.role_permissions.read_by_id(&f.context, id).await,
        Err(RepositoryError::RolePermissionNotFound)
    ));
    assert!(
        f.role_permissions
            .read_by_permission_id(&f.context, p)
            .await
            .unwrap()
            .is_empty()
    );
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn role_permission_bulk_deletion_uses_or_and_requires_an_id() {
    let f = Fixture::new().await;
    let a = f.role("a", false).await;
    let b = f.role("b", false).await;
    let x = f.permission("x").await;
    let y = f.permission("y").await;
    f.pivot(a, x).await;
    f.pivot(a, y).await;
    f.pivot(b, x).await;
    let keep = f.pivot(b, y).await;
    assert!(matches!(
        f.role_permissions
            .delete_by_role_id_or_permission_id(&f.context, None, None)
            .await,
        Err(RepositoryError::BadArgs)
    ));
    f.role_permissions
        .delete_by_role_id_or_permission_id(&f.context, Some(a), Some(x))
        .await
        .unwrap();
    assert!(
        f.role_permissions
            .read_by_role_id(&f.context, a)
            .await
            .unwrap()
            .is_empty()
    );
    assert_eq!(
        f.role_permissions
            .read_by_id(&f.context, keep)
            .await
            .unwrap()
            .role
            .id,
        b
    );
    assert!(matches!(
        f.role_permissions
            .delete_by_role_id_or_permission_id(&f.context, Some(a), None)
            .await,
        Err(RepositoryError::RolePermissionNotFound)
    ));
    f.role_permissions
        .delete_by_role_id_or_permission_id(&f.context, None, Some(y))
        .await
        .unwrap();
    assert!(
        f.role_permissions
            .read_by_permission_id(&f.context, y)
            .await
            .unwrap()
            .is_empty()
    );
    f.close().await;
}
