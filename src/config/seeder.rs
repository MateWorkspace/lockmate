use std::collections::{HashMap, HashSet};

use serde::de::Error as _;

use crate::domain::models::SeedData;

pub fn load() -> Result<SeedData, serde_json::Error> {
    parse(
        include_str!("../../database/seeder/space.json"),
        include_str!("../../database/seeder/user.json"),
        include_str!("../../database/seeder/space_member.json"),
    )
}

/// Parses seed definitions and checks relationships; scalar validation is separate.
pub fn parse(
    spaces_json: &str,
    users_json: &str,
    space_members_json: &str,
) -> Result<SeedData, serde_json::Error> {
    let data = SeedData {
        spaces: serde_json::from_str(spaces_json)?,
        users: serde_json::from_str(users_json)?,
        space_members: serde_json::from_str(space_members_json)?,
    };
    validate_graph(&data)?;
    Ok(data)
}

fn validate_graph(data: &SeedData) -> Result<(), serde_json::Error> {
    let mut spaces = HashMap::new();
    for space in &data.spaces {
        if spaces.insert(space.slug.as_str(), space).is_some() {
            return Err(serde_json::Error::custom("duplicate seed space slug"));
        }

        let mut permissions = HashSet::new();
        for permission in &space.permissions {
            if !permissions.insert(permission.slug.as_str()) {
                return Err(serde_json::Error::custom(
                    "duplicate seed permission slug within a space",
                ));
            }
        }
        if space.roles.iter().filter(|role| role.is_default).count() != 1 {
            return Err(serde_json::Error::custom(
                "each seed space must have exactly one default role",
            ));
        }

        let mut roles = HashSet::new();
        for role in &space.roles {
            if !roles.insert(role.slug.as_str()) {
                return Err(serde_json::Error::custom(
                    "duplicate seed role slug within a space",
                ));
            }

            let mut assigned = HashSet::new();
            for permission in &role.permissions {
                if !assigned.insert(permission.as_str()) {
                    return Err(serde_json::Error::custom(
                        "duplicate seed role permission reference",
                    ));
                }
                if !permissions.contains(permission.as_str()) {
                    return Err(serde_json::Error::custom(
                        "seed role references an unknown permission in its space",
                    ));
                }
            }
        }
    }

    let mut users = HashSet::new();
    for user in &data.users {
        if !users.insert(user.username.as_str()) {
            return Err(serde_json::Error::custom("duplicate seed username"));
        }
    }

    let mut memberships = HashSet::new();
    for member in &data.space_members {
        if !memberships.insert((member.space_slug.as_str(), member.username.as_str())) {
            return Err(serde_json::Error::custom("duplicate seed space membership"));
        }
        let space = spaces.get(member.space_slug.as_str()).ok_or_else(|| {
            serde_json::Error::custom("seed membership references an unknown space")
        })?;
        if !users.contains(member.username.as_str()) {
            return Err(serde_json::Error::custom(
                "seed membership references an unknown user",
            ));
        }

        let mut assigned = HashSet::new();
        for role in &member.roles {
            if !assigned.insert(role.as_str()) {
                return Err(serde_json::Error::custom(
                    "duplicate seed membership role reference",
                ));
            }
            if !space.roles.iter().any(|existing| existing.slug == *role) {
                return Err(serde_json::Error::custom(
                    "seed membership references an unknown role in its space",
                ));
            }
        }
    }
    Ok(())
}
