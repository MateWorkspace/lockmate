use super::support::Fixture;
use lockmate::domain::{
    contracts::repository::{UpdateUser, UserFilter},
    models::RepositoryError,
};
use serde_json::json;

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn user_crud_defaults_contacts_verification_and_avatar() {
    let f = Fixture::new().await;
    let id = f.user("user").await;
    let item = f.users.read_by_id(&f.context, id).await.unwrap();
    assert_eq!(item.bio, "");
    assert!(item.email.is_none());
    assert!(item.phone.is_none());
    assert!(item.avatar_path.is_none());
    assert!(!item.is_email_verified);
    assert!(!item.is_phone_verified);
    assert_eq!(item.preferences, json!({}));
    assert_eq!(
        f.users
            .read_by_username(&f.context, "user")
            .await
            .unwrap()
            .id,
        id
    );
    f.users
        .update_by_id(
            &f.context,
            id,
            UpdateUser {
                email: Some(Some("first@example.com".into())),
                phone: Some(Some("+123456789".into())),
                avatar_path: Some(Some("avatars/user.webp".into())),
                is_email_verified: Some(true),
                is_phone_verified: Some(true),
                bio: Some("bio".into()),
                preferences: Some(json!({"nested":{"ok":true}})),
                by: Some(43),
                ..Default::default()
            },
        )
        .await
        .unwrap();
    assert_eq!(
        f.users
            .read_by_email(&f.context, "first@example.com")
            .await
            .unwrap()
            .id,
        id
    );
    assert_eq!(
        f.users
            .read_by_phone(&f.context, "+123456789")
            .await
            .unwrap()
            .id,
        id
    );
    f.users
        .update_by_id(
            &f.context,
            id,
            UpdateUser {
                email: Some(Some("first@example.com".into())),
                ..Default::default()
            },
        )
        .await
        .unwrap();
    assert!(
        f.users
            .read_by_id(&f.context, id)
            .await
            .unwrap()
            .is_email_verified
    );
    f.users
        .update_by_id(
            &f.context,
            id,
            UpdateUser {
                email: Some(Some("second@example.com".into())),
                phone: Some(None),
                ..Default::default()
            },
        )
        .await
        .unwrap();
    let item = f.users.read_by_id(&f.context, id).await.unwrap();
    assert!(!item.is_email_verified);
    assert!(!item.is_phone_verified);
    assert!(item.phone.is_none());
    assert_eq!(item.avatar_path.as_deref(), Some("avatars/user.webp"));
    assert_eq!(item.preferences, json!({"nested":{"ok":true}}));
    f.users
        .update_by_id(
            &f.context,
            id,
            UpdateUser {
                email: Some(None),
                is_email_verified: Some(true),
                avatar_path: Some(None),
                preferences: Some(json!(null)),
                ..Default::default()
            },
        )
        .await
        .unwrap();
    let item = f.users.read_by_id(&f.context, id).await.unwrap();
    assert!(item.email.is_none());
    assert!(item.is_email_verified);
    assert!(item.avatar_path.is_none());
    assert_eq!(item.preferences, json!(null));
    f.users
        .delete_by_id(&f.context, id, Some(44))
        .await
        .unwrap();
    assert!(matches!(
        f.users.read_by_id(&f.context, id).await,
        Err(RepositoryError::UserNotFound)
    ));
    assert!(matches!(
        f.users.read_by_username(&f.context, "user").await,
        Err(RepositoryError::UserNotFound)
    ));
    assert!(matches!(
        f.users
            .update_by_id(&f.context, id, UpdateUser::default())
            .await,
        Err(RepositoryError::UserNotFound)
    ));
    assert!(matches!(
        f.users.delete_by_id(&f.context, id, None).await,
        Err(RepositoryError::UserNotFound)
    ));
    assert_ne!(f.user("user").await, id);
    f.close().await;
}

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn user_conflicts_filters_and_creation_options() {
    let f = Fixture::new().await;
    let mut input = Fixture::user_input("first");
    input.email = Some("first@example.com".into());
    input.phone = Some("+123456789".into());
    input.avatar_path = Some("avatars/first.webp".into());
    input.bio = Some("bio".into());
    input.is_email_verified = Some(true);
    input.is_phone_verified = Some(true);
    let id = f.users.create(&f.context, input.clone()).await.unwrap();
    assert!(matches!(
        f.users.create(&f.context, input.clone()).await,
        Err(RepositoryError::UserUsernameConflict)
    ));
    input.username = "second".into();
    assert!(matches!(
        f.users.create(&f.context, input.clone()).await,
        Err(RepositoryError::UserEmailConflict)
    ));
    input.email = Some("second@example.com".into());
    assert!(matches!(
        f.users.create(&f.context, input.clone()).await,
        Err(RepositoryError::UserPhoneConflict)
    ));
    input.phone = Some("+987654321".into());
    let second = f.users.create(&f.context, input).await.unwrap();
    assert!(matches!(
        f.users
            .update_by_id(
                &f.context,
                second,
                UpdateUser {
                    username: Some("first".into()),
                    ..Default::default()
                }
            )
            .await,
        Err(RepositoryError::UserUsernameConflict)
    ));
    let (items, total) = f
        .users
        .read_by_filter(
            &f.context,
            UserFilter {
                page: 1,
                limit: 10,
                search: Some("FIRST@".into()),
                is_email_verified: Some(true),
                is_phone_verified: Some(true),
            },
        )
        .await
        .unwrap();
    assert_eq!(total, 1);
    assert_eq!(items[0].id, id);
    assert_eq!(items[0].avatar_path.as_deref(), Some("avatars/first.webp"));
    assert_eq!(items[0].bio, "bio");
    f.users.delete_by_id(&f.context, id, None).await.unwrap();
    assert!(matches!(
        f.users.read_by_email(&f.context, "first@example.com").await,
        Err(RepositoryError::UserNotFound)
    ));
    assert!(matches!(
        f.users.read_by_phone(&f.context, "+123456789").await,
        Err(RepositoryError::UserNotFound)
    ));
    f.close().await;
}
