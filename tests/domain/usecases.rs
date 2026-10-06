use std::{error::Error, io};

use lockmate::domain::{
    contracts::repository,
    models::{
        ApiKey, ApiKeyError, CachedApiKey, CachedSpaceMemberWithUser, CachedUser, CachingError,
        PasswordError, RepositoryError, SpaceMember, TokenError, TransactorError, UsecaseError,
        User, ValidatorError,
    },
    usecases::{
        UsecaseFuture, auth, management, profile,
        shared::{ApiKeyView, Page, SpaceMemberWithUser, UserView},
    },
};
use serde_json::{Value, json};

#[test]
fn usecases_are_object_safe_send_and_sync_and_return_send_futures() {
    fn contract<T: ?Sized + Send + Sync>() {}
    fn future<T: Send>() {}

    contract::<dyn auth::Session>();
    contract::<dyn management::Space>();
    contract::<dyn management::User>();
    contract::<dyn management::Permission>();
    contract::<dyn management::Role>();
    contract::<dyn management::SpaceMember>();
    contract::<dyn management::RolePermission>();
    contract::<dyn management::MemberRole>();
    contract::<dyn management::ApiKey>();
    contract::<dyn profile::Account>();
    contract::<dyn profile::Security>();
    contract::<dyn profile::ApiKey>();

    future::<UsecaseFuture<'static, auth::session::SessionResult>>();
    future::<UsecaseFuture<'static, Page<UserView>>>();
    future::<UsecaseFuture<'static, ()>>();
}

fn user() -> User {
    serde_json::from_value(json!({
        "id": 1,
        "name": "Test User",
        "bio": "Profile bio",
        "username": "test_user",
        "email": "user@example.com",
        "phone": "+1234567890",
        "password_hash": "PRIVATE_PASSWORD_HASH",
        "is_email_verified": true,
        "is_phone_verified": false,
        "avatar_path": "avatars/1.png",
        "preferences": {"theme": "dark"},
        "created_at": "2026-10-06T00:00:00Z",
        "created_by": 1
    }))
    .unwrap()
}

fn api_key() -> ApiKey {
    serde_json::from_value(json!({
        "id": 2,
        "space_id": 3,
        "member_id": 4,
        "name": "Test key",
        "description": "Key description",
        "hash": "PRIVATE_API_KEY_HASH",
        "redacted": "abcd",
        "preferences": {"purpose": "test"},
        "created_at": "2026-10-06T00:00:00Z",
        "created_by": 1
    }))
    .unwrap()
}

fn member() -> SpaceMember {
    serde_json::from_value(json!({
        "id": 4,
        "space_id": 3,
        "user_id": 1,
        "is_active": true,
        "preferences": {},
        "created_at": "2026-10-06T00:00:00Z"
    }))
    .unwrap()
}

fn assert_safe(value: Value) {
    fn walk(value: &Value) {
        match value {
            Value::Object(fields) => {
                for (key, value) in fields {
                    assert!(!matches!(key.as_str(), "password_hash" | "hash"));
                    walk(value);
                }
            }
            Value::Array(values) => values.iter().for_each(walk),
            _ => {}
        }
    }
    walk(&value);
    let text = value.to_string();
    assert!(!text.contains("PRIVATE_PASSWORD_HASH"));
    assert!(!text.contains("PRIVATE_API_KEY_HASH"));
}

#[test]
fn public_views_preserve_data_but_discard_credentials_from_entities_and_caches() {
    let user_view = UserView::from(user());
    assert_eq!(user_view, UserView::from(CachedUser::from(user())));
    assert_eq!(user_view.avatar_path.as_deref(), Some("avatars/1.png"));
    assert_eq!(user_view.email.as_deref(), Some("user@example.com"));
    assert!(user_view.is_email_verified);
    assert_eq!(user_view.preferences, json!({"theme": "dark"}));
    assert_safe(serde_json::to_value(&user_view).unwrap());

    let key_view = ApiKeyView::from(api_key());
    assert_eq!(key_view, ApiKeyView::from(CachedApiKey::from(api_key())));
    assert_eq!(key_view.redacted, "abcd");
    assert_safe(serde_json::to_value(&key_view).unwrap());

    let joined = SpaceMemberWithUser::from(repository::SpaceMemberWithUser {
        member: member(),
        user: user(),
    });
    let cached = SpaceMemberWithUser::from(CachedSpaceMemberWithUser {
        member: member(),
        user: user().into(),
    });
    assert_eq!(joined, cached);
    let page = Page::from((vec![joined], 10));
    assert_eq!(page.total, 10);
    assert_safe(serde_json::to_value(&page).unwrap());
    assert_safe(serde_json::to_value(Page::from((vec![key_view], 1))).unwrap());
}

#[test]
fn nullable_updates_distinguish_omission_set_and_clear() {
    let mut request = profile::account::UpdateProfileRequest::default();
    assert_eq!(request.email, None);
    assert_eq!(request.phone, None);
    assert_eq!(request.avatar_path, None);

    request.email = Some(Some("new@example.com".into()));
    request.phone = Some(Some("+1234567890".into()));
    request.avatar_path = Some(Some("avatars/new.png".into()));
    assert_eq!(
        request.email.as_ref().unwrap().as_deref(),
        Some("new@example.com")
    );
    assert_eq!(
        request.phone.as_ref().unwrap().as_deref(),
        Some("+1234567890")
    );
    assert_eq!(
        request.avatar_path.as_ref().unwrap().as_deref(),
        Some("avatars/new.png")
    );

    request.email = Some(None);
    request.phone = Some(None);
    request.avatar_path = Some(None);
    assert_eq!(request.email, Some(None));
    assert_eq!(request.phone, Some(None));
    assert_eq!(request.avatar_path, Some(None));

    let management = management::user::UpdateRequest {
        email: request.email,
        phone: request.phone,
        avatar_path: request.avatar_path,
        ..Default::default()
    };
    assert_eq!(management.email, Some(None));
    assert_eq!(management.phone, Some(None));
    assert_eq!(management.avatar_path, Some(None));
}

#[test]
fn usecase_errors_keep_codes_and_sources_without_displaying_private_causes() {
    let cases: Vec<(UsecaseError, &str)> = vec![
        (UsecaseError::BadArgs, "BAD_ARGS"),
        (UsecaseError::BadState, "BAD_STATE"),
        (UsecaseError::Forbidden, "FORBIDDEN"),
        (UsecaseError::Unauthorized { source: None }, "UNAUTHORIZED"),
        (
            ValidatorError::UserEmailInvalid.into(),
            "USER_EMAIL_INVALID",
        ),
        (
            RepositoryError::UserEmailConflict.into(),
            "USER_EMAIL_CONFLICT",
        ),
        (PasswordError::Mismatch.into(), "UNAUTHORIZED"),
        (TokenError::Invalid.into(), "TOKEN_INVALID"),
        (TokenError::Expired.into(), "TOKEN_EXPIRED"),
        (ApiKeyError::Invalid.into(), "API_KEY_INVALID"),
    ];
    for (error, code) in cases {
        assert_eq!(error.code(), code);
    }

    let unauthorized = UsecaseError::Unauthorized {
        source: Some(Box::new(RepositoryError::UserNotFound)),
    };
    assert_eq!(unauthorized.to_string(), "authentication failed");
    assert!(unauthorized.source().unwrap().is::<RepositoryError>());

    let failure: UsecaseError = RepositoryError::Failure {
        source: Box::new(io::Error::other("PRIVATE_DATABASE_DETAIL")),
    }
    .into();
    assert_eq!(failure.code(), "FAILURE");
    assert_eq!(failure.to_string(), "repository operation failed");
    assert!(failure.source().unwrap().is::<RepositoryError>());
    assert_eq!(
        failure.source().unwrap().source().unwrap().to_string(),
        "PRIVATE_DATABASE_DETAIL"
    );

    let timeout: UsecaseError = TransactorError::Timeout {
        source: Box::new(io::Error::new(
            io::ErrorKind::TimedOut,
            "PRIVATE_TIMEOUT_DETAIL",
        )),
    }
    .into();
    assert_eq!(timeout.code(), "TIMEOUT");
    assert!(!timeout.to_string().contains("PRIVATE_TIMEOUT_DETAIL"));
    assert!(timeout.source().unwrap().is::<TransactorError>());

    let callback: UsecaseError = TransactorError::Callback {
        source: Box::new(UsecaseError::Forbidden),
    }
    .into();
    assert!(
        callback
            .source()
            .unwrap()
            .source()
            .unwrap()
            .is::<UsecaseError>()
    );

    // Cache failures remain distinct from the public usecase failure model.
    assert_eq!(CachingError::BadState.code(), "BAD_STATE");
}
