use std::{collections::HashSet, sync::Arc};

use lockmate::{
    config::seeder,
    domain::{
        contracts::utility::Validator,
        models::{AppContext, LoggerLevel, SeedUser},
    },
    infrastructure::utility::{logger::JsonLogger, validator::RegexValidator},
};
use serde_json::{Value, json};

fn definitions() -> (Value, Value, Value) {
    (
        serde_json::from_str(include_str!("../../database/seeder/space.json")).unwrap(),
        serde_json::from_str(include_str!("../../database/seeder/user.json")).unwrap(),
        serde_json::from_str(include_str!("../../database/seeder/space_member.json")).unwrap(),
    )
}

fn parse(
    spaces: &Value,
    users: &Value,
    members: &Value,
) -> Result<lockmate::domain::models::SeedData, serde_json::Error> {
    seeder::parse(
        &spaces.to_string(),
        &users.to_string(),
        &members.to_string(),
    )
}

#[test]
fn embedded_catalog_and_additional_membership_roles_are_consistent() {
    let data = seeder::load().unwrap();
    assert_eq!(data.spaces.len(), 1);
    assert_eq!(data.users.len(), 3);
    assert_eq!(data.space_members.len(), 3);
    let space = &data.spaces[0];
    assert_eq!(space.slug, "lockmate");
    assert_eq!(space.name, "Lockmate");
    assert!(space.is_active);
    assert_eq!(space.permissions.len(), 36);
    assert_eq!(space.roles.len(), 3);
    let permissions: HashSet<_> = space
        .permissions
        .iter()
        .map(|permission| permission.slug.as_str())
        .collect();
    assert_eq!(permissions.len(), 36);
    for slug in [
        "profile:get",
        "user_permission:get",
        "space:add",
        "space_member:set",
        "member_role:remove",
    ] {
        assert!(permissions.contains(slug));
    }
    for (slug, count) in [("super", 36), ("admin", 14), ("user", 4)] {
        let role = space.roles.iter().find(|role| role.slug == slug).unwrap();
        assert_eq!(role.permissions.len(), count);
        assert_eq!(role.is_default, slug == "user");
        assert!(
            role.permissions
                .iter()
                .all(|permission| permissions.contains(permission.as_str()))
        );
    }
    for member in &data.space_members {
        assert_eq!(member.space_slug, space.slug);
        assert!(member.is_active);
        assert!(
            data.users
                .iter()
                .any(|user| user.username == member.username)
        );
        if member.username == "user" {
            assert!(member.roles.is_empty());
        } else {
            assert_eq!(member.roles, vec![member.username.clone()]);
        }
    }
    for user in &data.users {
        assert!(user.bio.is_none());
        assert!(user.avatar_path.is_none());
        assert!(!user.password.is_empty());
    }
}

#[test]
fn every_embedded_scalar_passes_its_field_validator() {
    let data = seeder::load().unwrap();
    let validator = RegexValidator::new(Arc::new(JsonLogger::with_writer(
        std::io::sink(),
        LoggerLevel::None,
    )));
    let context = AppContext::default();
    for space in &data.spaces {
        validator.space_slug(&context, &space.slug).unwrap();
        validator.space_name(&context, &space.name).unwrap();
        validator.space_desc(&context, &space.description).unwrap();
        for permission in &space.permissions {
            validator
                .permission_slug(&context, &permission.slug)
                .unwrap();
            validator
                .permission_name(&context, &permission.name)
                .unwrap();
            validator
                .permission_desc(&context, &permission.description)
                .unwrap();
        }
        for role in &space.roles {
            validator.role_slug(&context, &role.slug).unwrap();
            validator.role_name(&context, &role.name).unwrap();
            validator.role_desc(&context, &role.description).unwrap();
            for permission in &role.permissions {
                validator.permission_slug(&context, permission).unwrap();
            }
        }
    }
    for user in &data.users {
        validator.user_name(&context, &user.name).unwrap();
        validator.user_username(&context, &user.username).unwrap();
        validator.user_password(&context, &user.password).unwrap();
        if let Some(bio) = &user.bio {
            validator.user_bio(&context, bio).unwrap();
        }
    }
    for member in &data.space_members {
        validator.space_slug(&context, &member.space_slug).unwrap();
        validator.user_username(&context, &member.username).unwrap();
        for role in &member.roles {
            validator.role_slug(&context, role).unwrap();
        }
    }
}

#[test]
fn graph_validation_rejects_duplicates_defaults_and_broken_references() {
    type Change = fn(&mut Value, &mut Value, &mut Value);
    let cases: [(&str, Change); 12] = [
        ("space slug", |s, _, _| {
            let copy = s[0].clone();
            s.as_array_mut().unwrap().push(copy);
        }),
        ("permission slug", |s, _, _| {
            let copy = s[0]["permissions"][0].clone();
            s[0]["permissions"].as_array_mut().unwrap().push(copy);
        }),
        ("role slug", |s, _, _| {
            let copy = s[0]["roles"][0].clone();
            s[0]["roles"].as_array_mut().unwrap().push(copy);
        }),
        ("exactly one default", |s, _, _| {
            s[0]["roles"][2]["is_default"] = false.into();
        }),
        ("exactly one default", |s, _, _| {
            s[0]["roles"][0]["is_default"] = true.into();
        }),
        ("role permission reference", |s, _, _| {
            s[0]["roles"][0]["permissions"]
                .as_array_mut()
                .unwrap()
                .push("profile:get".into());
        }),
        ("unknown permission", |s, _, _| {
            s[0]["roles"][0]["permissions"][0] = "unknown:get".into();
        }),
        ("username", |_, u, _| {
            let copy = u[0].clone();
            u.as_array_mut().unwrap().push(copy);
        }),
        ("space membership", |_, _, m| {
            let copy = m[0].clone();
            m.as_array_mut().unwrap().push(copy);
        }),
        ("unknown space", |_, _, m| {
            m[0]["space_slug"] = "other".into();
        }),
        ("unknown user", |_, _, m| {
            m[0]["username"] = "unknown".into();
        }),
        ("unknown role", |_, _, m| {
            m[0]["roles"][0] = "unknown".into();
        }),
    ];
    for (reason, change) in cases {
        let (mut spaces, mut users, mut members) = definitions();
        change(&mut spaces, &mut users, &mut members);
        let error = parse(&spaces, &users, &members)
            .err()
            .expect("invalid graph was accepted");
        assert!(error.is_data());
        assert!(error.to_string().contains(reason), "expected {reason}");
        assert!(!error.to_string().contains("12345678"));
    }
    let (spaces, users, mut members) = definitions();
    members[0]["roles"]
        .as_array_mut()
        .unwrap()
        .push("super".into());
    assert!(
        parse(&spaces, &users, &members)
            .err()
            .unwrap()
            .to_string()
            .contains("duplicate seed membership role")
    );
}

#[test]
fn repeated_catalog_slugs_are_valid_across_spaces_but_references_stay_local() {
    let (mut spaces, users, mut members) = definitions();
    let mut other = spaces[0].clone();
    other["slug"] = "other".into();
    spaces.as_array_mut().unwrap().push(other);
    members
        .as_array_mut()
        .unwrap()
        .push(json!({"space_slug":"other","username":"super","roles":["super"]}));
    assert!(parse(&spaces, &users, &members).is_ok());
    spaces[1]["permissions"]
        .as_array_mut()
        .unwrap()
        .push(json!({"slug":"only-other:get","name":"View other","description":""}));
    spaces[0]["roles"][0]["permissions"]
        .as_array_mut()
        .unwrap()
        .push("only-other:get".into());
    assert!(
        parse(&spaces, &users, &members)
            .err()
            .unwrap()
            .to_string()
            .contains("unknown permission in its space")
    );
    spaces[0]["roles"][0]["permissions"]
        .as_array_mut()
        .unwrap()
        .pop();
    spaces[1]["roles"].as_array_mut().unwrap().push(json!({"slug":"other-role","name":"Other role","description":"","is_default":false,"permissions":[]}));
    members[0]["roles"] = json!(["other-role"]);
    assert!(
        parse(&spaces, &users, &members)
            .err()
            .unwrap()
            .to_string()
            .contains("unknown role in its space")
    );
}

#[test]
fn typed_seed_parsing_rejects_malformed_json_unknown_fields_and_old_user_roles() {
    for position in 0..3 {
        let mut inputs = ["[]", "[]", "[]"];
        inputs[position] = "[";
        assert!(seeder::parse(inputs[0], inputs[1], inputs[2]).is_err());
    }
    for position in 0..5 {
        let (mut spaces, mut users, mut members) = definitions();
        let target = match position {
            0 => &mut spaces[0],
            1 => &mut spaces[0]["permissions"][0],
            2 => &mut spaces[0]["roles"][0],
            3 => &mut users[0],
            _ => &mut members[0],
        };
        target["unexpected"] = true.into();
        assert!(
            parse(&spaces, &users, &members)
                .err()
                .unwrap()
                .to_string()
                .contains("unknown field")
        );
    }
    let (spaces, mut users, members) = definitions();
    users[0]["role_name"] = "super".into();
    assert!(parse(&spaces, &users, &members).is_err());
}

#[test]
fn active_flags_additional_roles_and_user_profile_options_have_defined_defaults() {
    let (mut spaces, mut users, mut members) = definitions();
    spaces[0].as_object_mut().unwrap().remove("is_active");
    members[0].as_object_mut().unwrap().remove("is_active");
    members[0].as_object_mut().unwrap().remove("roles");
    users[0]["bio"] = "Profile".into();
    users[0]["avatar_path"] = "avatars/users/super.webp".into();
    let data = parse(&spaces, &users, &members).unwrap();
    assert!(data.spaces[0].is_active);
    assert!(data.space_members[0].is_active);
    assert!(data.space_members[0].roles.is_empty());
    assert_eq!(data.users[0].bio.as_deref(), Some("Profile"));
    assert_eq!(
        data.users[0].avatar_path.as_deref(),
        Some("avatars/users/super.webp")
    );
    spaces[0]["is_active"] = false.into();
    members[0]["is_active"] = false.into();
    let data = parse(&spaces, &users, &members).unwrap();
    assert!(!data.spaces[0].is_active);
    assert!(!data.space_members[0].is_active);
    let user: SeedUser = serde_json::from_str(r#"{"name":"User","username":"user","password":"test","avatar_path":"avatars/users/42/avatar.webp"}"#).unwrap();
    assert_eq!(
        user.avatar_path.as_deref(),
        Some("avatars/users/42/avatar.webp")
    );
}
