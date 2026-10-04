use super::support::Fixture;
use lockmate::domain::{
    contracts::repository::{
        ApiKeyFilter, CreateMemberRole, CreatePermission, CreateRole, CreateRolePermission,
        CreateSpaceMember, PermissionFilter, RoleFilter, SpaceMemberFilter, UpdateApiKey,
        UpdatePermission, UpdateRole, UpdateSpaceMember,
    },
    models::RepositoryError,
};
use mate_pgdt::sqlx;

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn every_scoped_repository_rejects_wrong_space_reads_and_mutations() {
    let f = Fixture::new().await;
    let other = f.space("other").await;
    let role = f.role("admin", true).await;
    let permission = f.permission("settings.read").await;
    let pivot = f.pivot(role, permission).await;
    let user = f.user("alice").await;
    let member = f.member(f.space_id, user).await;
    let assignment = f
        .member_roles
        .read_by_member_id_and_role_id(&f.context, f.space_id, member, role)
        .await
        .unwrap()
        .member_role
        .id;
    let key = f
        .api_keys
        .create(
            &f.context,
            f.space_id,
            Fixture::key_input(member, "Key", "private-hash"),
        )
        .await
        .unwrap();
    // Slugs and display names can be reused by another space independently.
    let other_role = f
        .roles
        .create(
            &f.context,
            other,
            CreateRole {
                slug: "admin".into(),
                name: "admin".into(),
                description: None,
                is_default: Some(true),
                by: None,
            },
        )
        .await
        .unwrap();
    let other_permission = f
        .permissions
        .create(
            &f.context,
            other,
            CreatePermission {
                slug: "settings.read".into(),
                name: "settings.read".into(),
                description: None,
                by: None,
            },
        )
        .await
        .unwrap();
    f.roles
        .create(
            &f.context,
            f.space_id,
            CreateRole {
                slug: "another-role".into(),
                name: "admin".into(),
                description: None,
                is_default: None,
                by: None,
            },
        )
        .await
        .unwrap();
    f.permissions
        .create(
            &f.context,
            f.space_id,
            CreatePermission {
                slug: "another-permission".into(),
                name: "settings.read".into(),
                description: None,
                by: None,
            },
        )
        .await
        .unwrap();
    assert_ne!(other_role, role);
    assert_ne!(other_permission, permission);
    assert_eq!(
        f.roles.read_default(&f.context, other).await.unwrap().id,
        other_role
    );
    assert_eq!(
        f.roles
            .read_default(&f.context, f.space_id)
            .await
            .unwrap()
            .id,
        role
    );
    assert_eq!(
        f.permissions
            .read_by_slug(&f.context, other, "settings.read")
            .await
            .unwrap()
            .id,
        other_permission
    );
    macro_rules! missing {
        ($call:expr, $variant:ident) => {
            assert!(matches!($call.await, Err(RepositoryError::$variant)));
        };
    }
    missing!(
        f.permissions.read_by_id(&f.context, other, permission),
        PermissionNotFound
    );
    missing!(
        f.permissions
            .update_by_id(&f.context, other, permission, UpdatePermission::default()),
        PermissionNotFound
    );
    missing!(
        f.permissions
            .delete_by_id(&f.context, other, permission, None),
        PermissionNotFound
    );
    missing!(f.roles.read_by_id(&f.context, other, role), RoleNotFound);
    missing!(
        f.roles.update_by_id(
            &f.context,
            other,
            role,
            UpdateRole {
                is_default: Some(true),
                ..Default::default()
            }
        ),
        RoleNotFound
    );
    missing!(
        f.roles.delete_by_id(&f.context, other, role, None),
        RoleNotFound
    );
    assert_eq!(
        f.roles.read_default(&f.context, other).await.unwrap().id,
        other_role
    );
    missing!(
        f.space_members.read_by_id(&f.context, other, member),
        SpaceMemberNotFound
    );
    missing!(
        f.space_members.read_by_user_id(&f.context, other, user),
        SpaceMemberNotFound
    );
    missing!(
        f.space_members
            .update_by_id(&f.context, other, member, UpdateSpaceMember::default()),
        SpaceMemberNotFound
    );
    missing!(
        f.space_members
            .delete_by_id(&f.context, other, member, None),
        SpaceMemberNotFound
    );
    missing!(
        f.role_permissions.read_by_id(&f.context, other, pivot),
        RolePermissionNotFound
    );
    missing!(
        f.role_permissions
            .read_by_role_id_and_permission_id(&f.context, other, role, permission),
        RolePermissionNotFound
    );
    missing!(
        f.role_permissions.delete_by_id(&f.context, other, pivot),
        RolePermissionNotFound
    );
    missing!(
        f.role_permissions.delete_by_role_id_or_permission_id(
            &f.context,
            other,
            Some(role),
            Some(permission)
        ),
        RolePermissionNotFound
    );
    assert!(
        f.role_permissions
            .read_by_role_id(&f.context, other, role)
            .await
            .unwrap()
            .is_empty()
    );
    assert!(
        f.role_permissions
            .read_by_permission_id(&f.context, other, permission)
            .await
            .unwrap()
            .is_empty()
    );
    missing!(
        f.member_roles.read_by_id(&f.context, other, assignment),
        MemberRoleNotFound
    );
    missing!(
        f.member_roles
            .read_by_member_id_and_role_id(&f.context, other, member, role),
        MemberRoleNotFound
    );
    missing!(
        f.member_roles.delete_by_id(&f.context, other, assignment),
        MemberRoleNotFound
    );
    missing!(
        f.member_roles
            .delete_by_member_id_or_role_id(&f.context, other, Some(member), Some(role)),
        MemberRoleNotFound
    );
    missing!(
        f.member_roles
            .read_effective_roles_by_member_id(&f.context, other, member),
        SpaceMemberNotFound
    );
    assert!(
        f.member_roles
            .read_by_member_id(&f.context, other, member)
            .await
            .unwrap()
            .is_empty()
    );
    assert!(
        f.member_roles
            .read_by_role_id(&f.context, other, role)
            .await
            .unwrap()
            .is_empty()
    );
    missing!(
        f.api_keys.read_by_id(&f.context, other, key),
        ApiKeyNotFound
    );
    missing!(
        f.api_keys.read_by_hash(&f.context, other, "private-hash"),
        ApiKeyNotFound
    );
    missing!(
        f.api_keys
            .read_active_by_hash(&f.context, other, "private-hash"),
        ApiKeyNotFound
    );
    missing!(
        f.api_keys
            .update_by_id(&f.context, other, key, UpdateApiKey::default()),
        ApiKeyNotFound
    );
    missing!(
        f.api_keys.delete_by_id(&f.context, other, key, None),
        ApiKeyNotFound
    );
    assert_eq!(
        f.permissions
            .read_by_filter(
                &f.context,
                other,
                PermissionFilter {
                    page: 1,
                    limit: 10,
                    ..Default::default()
                }
            )
            .await
            .unwrap()
            .1,
        1
    );
    assert_eq!(
        f.roles
            .read_by_filter(
                &f.context,
                other,
                RoleFilter {
                    page: 1,
                    limit: 10,
                    ..Default::default()
                }
            )
            .await
            .unwrap()
            .1,
        1
    );
    assert_eq!(
        f.api_keys
            .read_by_filter(
                &f.context,
                other,
                ApiKeyFilter {
                    page: 1,
                    limit: 10,
                    user_id: Some(user),
                    ..Default::default()
                }
            )
            .await
            .unwrap()
            .1,
        0
    );
    assert_eq!(
        f.space_members
            .read_by_filter(
                &f.context,
                other,
                SpaceMemberFilter {
                    page: 1,
                    limit: 10,
                    user_id: Some(user),
                    ..Default::default()
                }
            )
            .await
            .unwrap()
            .1,
        0
    );
    // No attempted mutation changed the original records.
    assert!(
        f.api_keys
            .read_active_by_hash(&f.context, f.space_id, "private-hash")
            .await
            .is_ok()
    );
    assert!(
        f.role_permissions
            .read_by_id(&f.context, f.space_id, pivot)
            .await
            .is_ok()
    );
    assert!(
        f.member_roles
            .read_by_id(&f.context, f.space_id, assignment)
            .await
            .is_ok()
    );
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn guarded_relationships_reject_missing_deleted_and_foreign_parents() {
    let f = Fixture::new().await;
    let other = f.space("other").await;
    let role = f.role("default", true).await;
    let permission = f.permission("permission").await;
    let user = f.user("alice").await;
    let member = f.member(f.space_id, user).await;
    assert!(matches!(
        f.role_permissions
            .create(
                &f.context,
                other,
                CreateRolePermission {
                    role_id: role,
                    permission_id: permission,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    assert!(matches!(
        f.member_roles
            .create(
                &f.context,
                other,
                CreateMemberRole {
                    member_id: member,
                    role_id: role,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    assert!(matches!(
        f.api_keys
            .create(
                &f.context,
                other,
                Fixture::key_input(member, "Invalid", "hash")
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    assert!(matches!(
        f.space_members
            .create(
                &f.context,
                f.space_id,
                CreateSpaceMember {
                    user_id: -1,
                    is_active: None,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    assert!(matches!(
        f.permissions
            .create(
                &f.context,
                -1,
                CreatePermission {
                    slug: "bad".into(),
                    name: "Bad".into(),
                    description: None,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    // Manually inserted cross-space assignments must never leak through joined reads.
    f.db.exec(
        &f.context,
        sqlx::query(
            "INSERT INTO role_permission (space_id, role_id, permission_id) VALUES ($1,$2,$3)",
        )
        .bind(other)
        .bind(role)
        .bind(permission),
    )
    .await
    .unwrap();
    assert!(
        f.role_permissions
            .read_by_role_id(&f.context, other, role)
            .await
            .unwrap()
            .is_empty()
    );
    f.permissions
        .delete_by_id(&f.context, f.space_id, permission, None)
        .await
        .unwrap();
    assert!(matches!(
        f.role_permissions
            .create(
                &f.context,
                f.space_id,
                CreateRolePermission {
                    role_id: role,
                    permission_id: permission,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    f.space_members
        .delete_by_id(&f.context, f.space_id, member, None)
        .await
        .unwrap();
    assert!(matches!(
        f.member_roles
            .create(
                &f.context,
                f.space_id,
                CreateMemberRole {
                    member_id: member,
                    role_id: role,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    assert!(matches!(
        f.api_keys
            .create(
                &f.context,
                f.space_id,
                Fixture::key_input(member, "Invalid", "hash")
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    f.users.delete_by_id(&f.context, user, None).await.unwrap();
    assert!(matches!(
        f.space_members
            .create(
                &f.context,
                f.space_id,
                CreateSpaceMember {
                    user_id: user,
                    is_active: None,
                    by: None
                }
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    f.spaces
        .delete_by_id(&f.context, f.space_id, None)
        .await
        .unwrap();
    assert!(matches!(
        f.roles.read_by_id(&f.context, f.space_id, role).await,
        Err(RepositoryError::RoleNotFound)
    ));
    assert!(matches!(
        f.roles
            .create(
                &f.context,
                f.space_id,
                CreateRole {
                    slug: "new".into(),
                    name: "New".into(),
                    description: None,
                    is_default: Some(true),
                    by: None
                }
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    f.close().await;
}
