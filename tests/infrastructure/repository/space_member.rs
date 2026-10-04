use super::support::Fixture;
use lockmate::domain::{
    contracts::repository::{
        CreateRole, CreateSpaceMember, SpaceMemberFilter, UpdateRole, UpdateSpaceMember,
    },
    models::{RepositoryError, TransactorError},
};
use mate_pgdt::sqlx::{self, Row};
use serde_json::json;

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn membership_creation_requires_default_and_assigns_it_atomically() {
    let f = Fixture::new().await;
    let user = f.user("alice").await;
    let input = || CreateSpaceMember {
        user_id: user,
        is_active: None,
        by: Some(42),
    };
    assert!(matches!(
        f.space_members
            .create(&f.context, f.space_id, input())
            .await,
        Err(RepositoryError::RoleNotFound)
    ));
    let count: i64 =
        f.db.query_row(
            &f.context,
            sqlx::query("SELECT COUNT(*) AS total FROM space_members"),
        )
        .await
        .unwrap()
        .try_get("total")
        .unwrap();
    assert_eq!(count, 0);
    assert_eq!(f.logger.0.lock().unwrap().len(), 1);
    let role = f.role("member", true).await;
    let id = f.member(f.space_id, user).await;
    let member = f
        .space_members
        .read_by_user_id(&f.context, f.space_id, user)
        .await
        .unwrap();
    assert_eq!(member.id, id);
    assert_eq!(member.preferences, json!({}));
    assert!(member.is_active);
    let assigned = f
        .member_roles
        .read_by_member_id_and_role_id(&f.context, f.space_id, id, role)
        .await
        .unwrap();
    assert_eq!(assigned.member_role.audit.by, Some(42));
    assert!(matches!(
        f.space_members
            .create(&f.context, f.space_id, input())
            .await,
        Err(RepositoryError::SpaceMemberConflict)
    ));
    let other = f.space("other").await;
    let other_role = f
        .roles
        .create(
            &f.context,
            other,
            CreateRole {
                slug: "viewer".into(),
                name: "Viewer".into(),
                description: None,
                is_default: Some(true),
                by: None,
            },
        )
        .await
        .unwrap();
    let other_member = f.member(other, user).await;
    assert_eq!(
        f.member_roles
            .read_effective_roles_by_member_id(&f.context, other, other_member)
            .await
            .unwrap()[0]
            .id,
        other_role
    );
    assert_eq!(
        f.member_roles
            .read_effective_roles_by_member_id(&f.context, f.space_id, id)
            .await
            .unwrap()[0]
            .id,
        role
    );
    f.space_members
        .update_by_id(
            &f.context,
            f.space_id,
            id,
            UpdateSpaceMember {
                is_active: Some(false),
                preferences: Some(json!({"locale":"en"})),
                by: Some(43),
            },
        )
        .await
        .unwrap();
    let member = f
        .space_members
        .read_by_id(&f.context, f.space_id, id)
        .await
        .unwrap();
    assert!(!member.is_active);
    assert_eq!(member.audit.update.by, Some(43));
    assert!(matches!(
        f.space_members
            .create(&f.context, f.space_id, input())
            .await,
        Err(RepositoryError::SpaceMemberConflict)
    ));
    let (items, total) = f
        .space_members
        .read_by_filter(
            &f.context,
            f.space_id,
            SpaceMemberFilter {
                page: 1,
                limit: 10,
                search: Some("ALICE".into()),
                role_id: Some(role),
                user_id: Some(user),
                is_active: Some(false),
            },
        )
        .await
        .unwrap();
    assert_eq!(total, 1);
    assert_eq!(items[0].member.id, id);
    assert_eq!(items[0].user.username, "alice");
    f.space_members
        .delete_by_id(&f.context, f.space_id, id, Some(44))
        .await
        .unwrap();
    assert!(matches!(
        f.space_members
            .read_by_user_id(&f.context, f.space_id, user)
            .await,
        Err(RepositoryError::SpaceMemberNotFound)
    ));
    assert_ne!(f.member(f.space_id, user).await, id);
    // Removing project access preserves the shared account and other membership.
    assert!(f.users.read_by_id(&f.context, user).await.is_ok());
    assert!(
        f.space_members
            .read_by_id(&f.context, other, other_member)
            .await
            .is_ok()
    );
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn membership_join_reuses_transaction_and_rolls_back_failed_assignment() {
    let f = Fixture::new().await;
    f.role("default", true).await;
    let user = f.user("alice").await;
    // A failure in the second write must undo the first write and log only once.
    f.db.exec(
        &f.context,
        sqlx::query(
            "ALTER TABLE member_role ADD CONSTRAINT reject_assignment CHECK (created_by <> 99)",
        ),
    )
    .await
    .unwrap();
    assert!(matches!(
        f.space_members
            .create(
                &f.context,
                f.space_id,
                CreateSpaceMember {
                    user_id: user,
                    is_active: None,
                    by: Some(99)
                }
            )
            .await,
        Err(RepositoryError::BadArgs)
    ));
    assert!(matches!(
        f.space_members
            .read_by_user_id(&f.context, f.space_id, user)
            .await,
        Err(RepositoryError::SpaceMemberNotFound)
    ));
    let entries = f.logger.0.lock().unwrap().clone();
    assert_eq!(
        entries
            .iter()
            .filter(|e| e.tag.ends_with("/create"))
            .count(),
        1
    );
    let members = f.space_members.clone();
    let space_id = f.space_id;
    let outcome = f
        .transactor
        .with_tx(
            &f.context,
            Box::new(move |context| {
                Box::pin(async move {
                    members
                        .create(
                            &context,
                            space_id,
                            CreateSpaceMember {
                                user_id: user,
                                is_active: None,
                                by: None,
                            },
                        )
                        .await?;
                    Err(RepositoryError::BadState.into())
                })
            }),
        )
        .await;
    assert!(matches!(outcome, Err(TransactorError::Callback { .. })));
    assert!(matches!(
        f.space_members
            .read_by_user_id(&f.context, f.space_id, user)
            .await,
        Err(RepositoryError::SpaceMemberNotFound)
    ));
    let members = f.space_members.clone();
    f.transactor
        .with_tx(
            &f.context,
            Box::new(move |context| {
                Box::pin(async move {
                    members
                        .create(
                            &context,
                            space_id,
                            CreateSpaceMember {
                                user_id: user,
                                is_active: None,
                                by: None,
                            },
                        )
                        .await?;
                    Ok(())
                })
            }),
        )
        .await
        .unwrap();
    let id = f
        .space_members
        .read_by_user_id(&f.context, f.space_id, user)
        .await
        .unwrap()
        .id;
    assert_eq!(
        f.member_roles
            .read_by_member_id(&f.context, f.space_id, id)
            .await
            .unwrap()
            .len(),
        1
    );
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn concurrent_joins_and_default_replacement_assign_exactly_one_role() {
    let f = Fixture::new().await;
    let old = f.role("old", true).await;
    let new = f.role("new", false).await;
    let alice = f.user("alice").await;
    let bob = f.user("bob").await;
    let (a, b, update) = tokio::join!(
        f.space_members.create(
            &f.context,
            f.space_id,
            CreateSpaceMember {
                user_id: alice,
                is_active: None,
                by: None
            }
        ),
        f.space_members.create(
            &f.context,
            f.space_id,
            CreateSpaceMember {
                user_id: bob,
                is_active: Some(false),
                by: None
            }
        ),
        f.roles.update_by_id(
            &f.context,
            f.space_id,
            new,
            UpdateRole {
                is_default: Some(true),
                ..Default::default()
            }
        ),
    );
    update.unwrap();
    for member in [a.unwrap(), b.unwrap()] {
        let assigned = f
            .member_roles
            .read_by_member_id(&f.context, f.space_id, member)
            .await
            .unwrap();
        assert_eq!(assigned.len(), 1);
        assert!([old, new].contains(&assigned[0].role.id));
    }
    assert!(f.logger.0.lock().unwrap().is_empty());
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn default_role_lock_does_not_block_other_spaces_or_reassign_existing_members() {
    let f = Fixture::new().await;
    let old = f.role("old", true).await;
    let user = f.user("alice").await;
    let member = f.member(f.space_id, user).await;
    let new = f.role("new", true).await;
    assert_eq!(
        f.member_roles
            .read_effective_roles_by_member_id(&f.context, f.space_id, member)
            .await
            .unwrap()[0]
            .id,
        old
    );
    let other = f.space("other").await;
    let roles = f.roles.clone();
    let space_id = f.space_id;
    let (ready_send, ready_receive) = tokio::sync::oneshot::channel();
    let (release_send, release_receive) = tokio::sync::oneshot::channel();
    let hold = f.transactor.with_tx(
        &f.context,
        Box::new(move |context| {
            Box::pin(async move {
                roles
                    .update_by_id(
                        &context,
                        space_id,
                        new,
                        UpdateRole {
                            is_default: Some(true),
                            ..Default::default()
                        },
                    )
                    .await?;
                ready_send.send(()).unwrap();
                release_receive.await.unwrap();
                Ok(())
            })
        }),
    );
    let independent = async {
        ready_receive.await.unwrap();
        let outcome = tokio::time::timeout(
            std::time::Duration::from_secs(2),
            f.roles.create(
                &f.context,
                other,
                CreateRole {
                    slug: "other-default".into(),
                    name: "Default".into(),
                    description: None,
                    is_default: Some(true),
                    by: None,
                },
            ),
        )
        .await;
        // Release the held transaction before asserting, so a failure cannot hang the join.
        release_send.send(()).unwrap();
        outcome
            .expect("another space must not wait on this space's default lock")
            .unwrap()
    };
    let (held, other_default) = tokio::join!(hold, independent);
    held.unwrap();
    assert_eq!(
        f.roles.read_default(&f.context, other).await.unwrap().id,
        other_default
    );
    assert_eq!(
        f.roles
            .read_default(&f.context, f.space_id)
            .await
            .unwrap()
            .id,
        new
    );
    f.close().await;
}
