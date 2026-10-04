use super::support::Fixture;
use lockmate::domain::{
    contracts::repository::{CreateMemberRole, UpdateSpace, UpdateSpaceMember},
    models::RepositoryError,
};

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn member_role_crud_joined_reads_and_bulk_or_deletion() {
    let f = Fixture::new().await;
    let default = f.role("default", true).await;
    let admin = f.role("admin", false).await;
    let alice = f.user("alice").await;
    let bob = f.user("bob").await;
    let a = f.member(f.space_id, alice).await;
    let b = f.member(f.space_id, bob).await;
    let input = || CreateMemberRole {
        member_id: a,
        role_id: admin,
        by: Some(43),
    };
    let id = f
        .member_roles
        .create(&f.context, f.space_id, input())
        .await
        .unwrap();
    let details = f
        .member_roles
        .read_by_id(&f.context, f.space_id, id)
        .await
        .unwrap();
    assert_eq!(details.member.id, a);
    assert_eq!(details.role.id, admin);
    assert_eq!(details.member_role.audit.by, Some(43));
    assert_eq!(
        f.member_roles
            .read_by_member_id_and_role_id(&f.context, f.space_id, a, admin)
            .await
            .unwrap()
            .member_role
            .id,
        id
    );
    assert!(matches!(
        f.member_roles.create(&f.context, f.space_id, input()).await,
        Err(RepositoryError::MemberRoleConflict)
    ));
    assert_eq!(
        f.member_roles
            .read_by_role_id(&f.context, f.space_id, default)
            .await
            .unwrap()
            .len(),
        2
    );
    assert_eq!(
        f.member_roles
            .read_by_member_id(&f.context, f.space_id, a)
            .await
            .unwrap()
            .len(),
        2
    );
    assert!(matches!(
        f.member_roles
            .delete_by_member_id_or_role_id(&f.context, f.space_id, None, None)
            .await,
        Err(RepositoryError::BadArgs)
    ));
    // Delete all assignments for a, plus the default role for b; no AND semantics.
    f.member_roles
        .delete_by_member_id_or_role_id(&f.context, f.space_id, Some(a), Some(default))
        .await
        .unwrap();
    assert!(
        f.member_roles
            .read_by_member_id(&f.context, f.space_id, a)
            .await
            .unwrap()
            .is_empty()
    );
    assert!(
        f.member_roles
            .read_by_member_id(&f.context, f.space_id, b)
            .await
            .unwrap()
            .is_empty()
    );
    assert!(matches!(
        f.member_roles.read_by_id(&f.context, f.space_id, id).await,
        Err(RepositoryError::MemberRoleNotFound)
    ));
    assert!(matches!(
        f.member_roles
            .delete_by_id(&f.context, f.space_id, id)
            .await,
        Err(RepositoryError::MemberRoleNotFound)
    ));
    let id = f
        .member_roles
        .create(&f.context, f.space_id, input())
        .await
        .unwrap();
    f.roles
        .delete_by_id(&f.context, f.space_id, admin, None)
        .await
        .unwrap();
    // Cleanup remains possible when a role has been soft-deleted.
    f.member_roles
        .delete_by_id(&f.context, f.space_id, id)
        .await
        .unwrap();
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn effective_access_unions_permissions_and_distinguishes_management_reads() {
    let f = Fixture::new().await;
    let basic = f.role("basic", true).await;
    let admin = f.role("admin", false).await;
    let p1 = f.permission("settings.read").await;
    let p2 = f.permission("settings.write").await;
    f.pivot(basic, p1).await;
    f.pivot(admin, p1).await;
    f.pivot(admin, p2).await;
    let user = f.user("alice").await;
    let member = f.member(f.space_id, user).await;
    let assignment = f
        .member_roles
        .create(
            &f.context,
            f.space_id,
            CreateMemberRole {
                member_id: member,
                role_id: admin,
                by: None,
            },
        )
        .await
        .unwrap();
    let roles = f
        .member_roles
        .read_effective_roles_by_member_id(&f.context, f.space_id, member)
        .await
        .unwrap();
    assert_eq!(
        roles.iter().map(|r| r.slug.as_str()).collect::<Vec<_>>(),
        vec!["admin", "basic"]
    );
    let permissions = f
        .member_roles
        .read_effective_permissions_by_member_id(&f.context, f.space_id, member)
        .await
        .unwrap();
    assert_eq!(
        permissions.iter().map(|p| p.id).collect::<Vec<_>>(),
        vec![p1, p2]
    );
    let key = f
        .api_keys
        .create(
            &f.context,
            f.space_id,
            Fixture::key_input(member, "Key", "private-hash"),
        )
        .await
        .unwrap();
    let details = f
        .api_keys
        .read_active_by_hash(&f.context, f.space_id, "private-hash")
        .await
        .unwrap();
    assert_eq!(details.api_key.id, key);
    assert_eq!(details.member.id, member);
    assert_eq!(details.user.id, user);
    assert_eq!(details.space.id, f.space_id);
    f.space_members
        .update_by_id(
            &f.context,
            f.space_id,
            member,
            UpdateSpaceMember {
                is_active: Some(false),
                ..Default::default()
            },
        )
        .await
        .unwrap();
    assert!(matches!(
        f.member_roles
            .read_effective_roles_by_member_id(&f.context, f.space_id, member)
            .await,
        Err(RepositoryError::BadState)
    ));
    assert!(matches!(
        f.api_keys
            .read_active_by_hash(&f.context, f.space_id, "private-hash")
            .await,
        Err(RepositoryError::ApiKeyNotFound)
    ));
    assert!(
        f.api_keys
            .read_by_hash(&f.context, f.space_id, "private-hash")
            .await
            .is_ok()
    );
    assert_eq!(
        f.member_roles
            .read_by_member_id(&f.context, f.space_id, member)
            .await
            .unwrap()
            .len(),
        2
    );
    f.space_members
        .update_by_id(
            &f.context,
            f.space_id,
            member,
            UpdateSpaceMember {
                is_active: Some(true),
                ..Default::default()
            },
        )
        .await
        .unwrap();
    f.spaces
        .update_by_id(
            &f.context,
            f.space_id,
            UpdateSpace {
                is_active: Some(false),
                ..Default::default()
            },
        )
        .await
        .unwrap();
    assert!(matches!(
        f.member_roles
            .read_effective_permissions_by_member_id(&f.context, f.space_id, member)
            .await,
        Err(RepositoryError::BadState)
    ));
    assert!(matches!(
        f.api_keys
            .read_active_by_hash(&f.context, f.space_id, "private-hash")
            .await,
        Err(RepositoryError::ApiKeyNotFound)
    ));
    assert!(
        f.roles
            .read_by_id(&f.context, f.space_id, admin)
            .await
            .is_ok()
    );
    f.spaces
        .update_by_id(
            &f.context,
            f.space_id,
            UpdateSpace {
                is_active: Some(true),
                ..Default::default()
            },
        )
        .await
        .unwrap();
    assert!(
        f.api_keys
            .read_active_by_hash(&f.context, f.space_id, "private-hash")
            .await
            .is_ok()
    );
    f.member_roles
        .delete_by_id(&f.context, f.space_id, assignment)
        .await
        .unwrap();
    assert_eq!(
        f.member_roles
            .read_effective_permissions_by_member_id(&f.context, f.space_id, member)
            .await
            .unwrap()
            .len(),
        1
    );
    f.permissions
        .delete_by_id(&f.context, f.space_id, p1, None)
        .await
        .unwrap();
    assert!(
        f.member_roles
            .read_effective_permissions_by_member_id(&f.context, f.space_id, member)
            .await
            .unwrap()
            .is_empty()
    );
    f.roles
        .delete_by_id(&f.context, f.space_id, basic, None)
        .await
        .unwrap();
    assert!(
        f.member_roles
            .read_effective_roles_by_member_id(&f.context, f.space_id, member)
            .await
            .unwrap()
            .is_empty()
    );
    // A new default is required before a replacement membership can join.
    f.role("replacement", true).await;
    f.space_members
        .delete_by_id(&f.context, f.space_id, member, None)
        .await
        .unwrap();
    let replacement = f.member(f.space_id, user).await;
    assert_ne!(replacement, member);
    assert!(matches!(
        f.api_keys
            .read_active_by_hash(&f.context, f.space_id, "private-hash")
            .await,
        Err(RepositoryError::ApiKeyNotFound)
    ));
    assert!(matches!(
        f.member_roles
            .read_effective_roles_by_member_id(&f.context, f.space_id, member)
            .await,
        Err(RepositoryError::SpaceMemberNotFound)
    ));
    f.users.delete_by_id(&f.context, user, None).await.unwrap();
    assert!(matches!(
        f.member_roles
            .read_effective_roles_by_member_id(&f.context, f.space_id, replacement)
            .await,
        Err(RepositoryError::SpaceMemberNotFound)
    ));
    f.close().await;
}
