use std::collections::HashSet;

use lockmate::{config::seeder, domain::models::SeedUser};

#[test]
fn embedded_seed_data_has_resolvable_references_and_optional_user_fields() {
    let data = seeder::load().unwrap();
    assert_eq!(data.permissions.len(), 25);
    assert_eq!(data.roles.len(), 3);
    assert_eq!(data.users.len(), 3);
    let permissions: HashSet<_> = data
        .permissions
        .iter()
        .map(|value| value.name.as_str())
        .collect();
    let roles: HashSet<_> = data.roles.iter().map(|value| value.name.as_str()).collect();
    assert_eq!(permissions.len(), data.permissions.len());
    assert_eq!(roles.len(), data.roles.len());
    assert_eq!(data.roles.iter().filter(|role| role.is_default).count(), 1);
    for role in &data.roles {
        assert!(
            role.permissions
                .iter()
                .all(|permission| permissions.contains(permission.as_str()))
        );
    }
    for user in &data.users {
        assert!(roles.contains(user.role_name.as_str()));
        assert!(user.bio.is_none());
        assert!(user.avatar_path.is_none());
        assert!(!user.password.is_empty());
    }
}

#[test]
fn seed_user_accepts_an_object_storage_avatar_key() {
    let user: SeedUser = serde_json::from_str(
        r#"{
        "role_name":"user", "name":"User", "username":"user", "password":"test",
        "avatar_path":"avatars/users/42/avatar.webp"
    }"#,
    )
    .unwrap();
    assert_eq!(
        user.avatar_path.as_deref(),
        Some("avatars/users/42/avatar.webp")
    );
}
