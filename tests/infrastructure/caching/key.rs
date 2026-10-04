use super::support;
use lockmate::{
    domain::{
        contracts::{caching::KeyBuilder, repository as repo},
        models::*,
    },
    infrastructure::caching::Sha256KeyBuilder,
};
use std::{sync::Arc, time::Duration};

#[test]
fn canonical_keys_match_independent_golden_vectors_for_every_query() {
    let builder = Sha256KeyBuilder::new("test").unwrap();
    let context = AppContext::default();
    let cases = vec![
        (
            CacheEntryKey {
                query: CacheQuery::SpaceById { id: 2 },
                revisions: vec![CacheRevision {
                    family: CacheFamily::Spaces,
                    token: "generation".into(),
                }],
            },
            "test:cache:v1:global:space:read_by_id:038966de9f6b9a901b20b4c6ca8b2a46009feebe031babc842d43690c0bc222b:5eb40dd7d43e48e73cede453758015bf0c21d39c5c42d0bcbc7ad97e0001900c",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::SpaceBySlug {
                    slug: "private-selector".into(),
                },
                revisions: vec![CacheRevision {
                    family: CacheFamily::Spaces,
                    token: "generation".into(),
                }],
            },
            "test:cache:v1:global:space:read_by_slug:f904dee38cbdcd0c4342c71381d702547d3a479eb1793bb5a740eb6ab73c1210:5eb40dd7d43e48e73cede453758015bf0c21d39c5c42d0bcbc7ad97e0001900c",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::SpaceByFilter {
                    filter: repo::SpaceFilter {
                        page: 2,
                        limit: 10,
                        search: Some("private-selector".into()),
                        is_active: Some(true),
                    },
                },
                revisions: vec![CacheRevision {
                    family: CacheFamily::Spaces,
                    token: "generation".into(),
                }],
            },
            "test:cache:v1:global:space:read_by_filter:af6301c9e68733dcfa63e734a87a3c0c9c30ceebe643d9e7021c7516604d876e:5eb40dd7d43e48e73cede453758015bf0c21d39c5c42d0bcbc7ad97e0001900c",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::UserById { id: 2 },
                revisions: vec![CacheRevision {
                    family: CacheFamily::Users,
                    token: "generation".into(),
                }],
            },
            "test:cache:v1:global:user:read_by_id:038966de9f6b9a901b20b4c6ca8b2a46009feebe031babc842d43690c0bc222b:e4f8819355632bfaba3d24017002d161926e9585689669cb4d0d00956daabc99",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::UserByUsername {
                    username: "private-selector".into(),
                },
                revisions: vec![CacheRevision {
                    family: CacheFamily::Users,
                    token: "generation".into(),
                }],
            },
            "test:cache:v1:global:user:read_by_username:f904dee38cbdcd0c4342c71381d702547d3a479eb1793bb5a740eb6ab73c1210:e4f8819355632bfaba3d24017002d161926e9585689669cb4d0d00956daabc99",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::UserByEmail {
                    email: "private-selector".into(),
                },
                revisions: vec![CacheRevision {
                    family: CacheFamily::Users,
                    token: "generation".into(),
                }],
            },
            "test:cache:v1:global:user:read_by_email:f904dee38cbdcd0c4342c71381d702547d3a479eb1793bb5a740eb6ab73c1210:e4f8819355632bfaba3d24017002d161926e9585689669cb4d0d00956daabc99",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::UserByPhone {
                    phone: "private-selector".into(),
                },
                revisions: vec![CacheRevision {
                    family: CacheFamily::Users,
                    token: "generation".into(),
                }],
            },
            "test:cache:v1:global:user:read_by_phone:f904dee38cbdcd0c4342c71381d702547d3a479eb1793bb5a740eb6ab73c1210:e4f8819355632bfaba3d24017002d161926e9585689669cb4d0d00956daabc99",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::UserByFilter {
                    filter: repo::UserFilter {
                        page: 2,
                        limit: 10,
                        search: Some("private-selector".into()),
                        is_email_verified: Some(true),
                        is_phone_verified: Some(true),
                    },
                },
                revisions: vec![CacheRevision {
                    family: CacheFamily::Users,
                    token: "generation".into(),
                }],
            },
            "test:cache:v1:global:user:read_by_filter:4bec3de6c24451fb82cef3f4fc741c36951fbb88e9e33e1c0de7f14d9d4abf6e:e4f8819355632bfaba3d24017002d161926e9585689669cb4d0d00956daabc99",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::PermissionById { space_id: 1, id: 2 },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Permissions { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:permission:read_by_id:49a64717d5d4cb19952e6eac2946415cf6879adacf9908e7d872332d32c6e684:abed7ab91ee1e56c59ea3084816dc046ddbae3dad9a04177d6d3f878e3b018bb",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::PermissionBySlug {
                    space_id: 1,
                    slug: "private-selector".into(),
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Permissions { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:permission:read_by_slug:e02f7b78d86259032e2012713535bda88fb4b570f5fb0a30f885100ffe00fae3:abed7ab91ee1e56c59ea3084816dc046ddbae3dad9a04177d6d3f878e3b018bb",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::PermissionByFilter {
                    space_id: 1,
                    filter: repo::PermissionFilter {
                        page: 2,
                        limit: 10,
                        search: Some("private-selector".into()),
                    },
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Permissions { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:permission:read_by_filter:4ab5edb26f44bad0efb76717815548540c747f960af6be9a2f77a80df2eb0338:abed7ab91ee1e56c59ea3084816dc046ddbae3dad9a04177d6d3f878e3b018bb",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::RoleById { space_id: 1, id: 2 },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:role:read_by_id:49a64717d5d4cb19952e6eac2946415cf6879adacf9908e7d872332d32c6e684:6a6c28d23b9acb9c1aeda70371d374f8e29db97834538060b82d1f01b2a73990",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::RoleBySlug {
                    space_id: 1,
                    slug: "private-selector".into(),
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:role:read_by_slug:e02f7b78d86259032e2012713535bda88fb4b570f5fb0a30f885100ffe00fae3:6a6c28d23b9acb9c1aeda70371d374f8e29db97834538060b82d1f01b2a73990",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::RoleDefault { space_id: 1 },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:role:read_default:080a9ed428559ef602668b4c00f114f1a11c3f6b02a435f0bdc154578e4d7f22:6a6c28d23b9acb9c1aeda70371d374f8e29db97834538060b82d1f01b2a73990",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::RoleByFilter {
                    space_id: 1,
                    filter: repo::RoleFilter {
                        page: 2,
                        limit: 10,
                        search: Some("private-selector".into()),
                        is_default: Some(true),
                    },
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:role:read_by_filter:b45dec1b847e418388d792541007fc1bd1188b0f709c44251cdc4a761975ac39:6a6c28d23b9acb9c1aeda70371d374f8e29db97834538060b82d1f01b2a73990",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::SpaceMemberById { space_id: 1, id: 2 },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Users,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::SpaceMembers { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::MemberRoles { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:space_member:read_by_id:49a64717d5d4cb19952e6eac2946415cf6879adacf9908e7d872332d32c6e684:a5abe5690d7fd99250e47df5efcf41c1a26ccdd5b2234fd91ebc9ff74c6f6ca3",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::SpaceMemberByUserId {
                    space_id: 1,
                    user_id: 2,
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Users,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::SpaceMembers { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::MemberRoles { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:space_member:read_by_user_id:49a64717d5d4cb19952e6eac2946415cf6879adacf9908e7d872332d32c6e684:a5abe5690d7fd99250e47df5efcf41c1a26ccdd5b2234fd91ebc9ff74c6f6ca3",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::SpaceMemberByFilter {
                    space_id: 1,
                    filter: repo::SpaceMemberFilter {
                        page: 2,
                        limit: 10,
                        search: Some("private-selector".into()),
                        user_id: Some(2),
                        role_id: Some(2),
                        is_active: Some(true),
                    },
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Users,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::SpaceMembers { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::MemberRoles { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:space_member:read_by_filter:10f11695b8d7c162e8abe8a423e9b216e12cd11d39d2c43516700cf27897f784:a5abe5690d7fd99250e47df5efcf41c1a26ccdd5b2234fd91ebc9ff74c6f6ca3",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::RolePermissionById { space_id: 1, id: 2 },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Permissions { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::RolePermissions { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:role_permission:read_by_id:49a64717d5d4cb19952e6eac2946415cf6879adacf9908e7d872332d32c6e684:f067db6490958337eacf2b52ebe5f437217389c7cf074b299a66639f8252acdd",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::RolePermissionByPair {
                    space_id: 1,
                    role_id: 2,
                    permission_id: 2,
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Permissions { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::RolePermissions { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:role_permission:read_by_role_id_and_permission_id:c87a0b507cd70dc26bd56b10f736aea70c0643ffd8067f543b23cd0079f2ece9:f067db6490958337eacf2b52ebe5f437217389c7cf074b299a66639f8252acdd",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::RolePermissionByRoleId {
                    space_id: 1,
                    role_id: 2,
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Permissions { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::RolePermissions { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:role_permission:read_by_role_id:49a64717d5d4cb19952e6eac2946415cf6879adacf9908e7d872332d32c6e684:f067db6490958337eacf2b52ebe5f437217389c7cf074b299a66639f8252acdd",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::RolePermissionByPermissionId {
                    space_id: 1,
                    permission_id: 2,
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Permissions { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::RolePermissions { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:role_permission:read_by_permission_id:49a64717d5d4cb19952e6eac2946415cf6879adacf9908e7d872332d32c6e684:f067db6490958337eacf2b52ebe5f437217389c7cf074b299a66639f8252acdd",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::MemberRoleById { space_id: 1, id: 2 },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Users,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::SpaceMembers { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::MemberRoles { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:member_role:read_by_id:49a64717d5d4cb19952e6eac2946415cf6879adacf9908e7d872332d32c6e684:a5abe5690d7fd99250e47df5efcf41c1a26ccdd5b2234fd91ebc9ff74c6f6ca3",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::MemberRoleByPair {
                    space_id: 1,
                    member_id: 2,
                    role_id: 2,
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Users,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::SpaceMembers { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::MemberRoles { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:member_role:read_by_member_id_and_role_id:c87a0b507cd70dc26bd56b10f736aea70c0643ffd8067f543b23cd0079f2ece9:a5abe5690d7fd99250e47df5efcf41c1a26ccdd5b2234fd91ebc9ff74c6f6ca3",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::MemberRoleByMemberId {
                    space_id: 1,
                    member_id: 2,
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Users,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::SpaceMembers { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::MemberRoles { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:member_role:read_by_member_id:49a64717d5d4cb19952e6eac2946415cf6879adacf9908e7d872332d32c6e684:a5abe5690d7fd99250e47df5efcf41c1a26ccdd5b2234fd91ebc9ff74c6f6ca3",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::MemberRoleByRoleId {
                    space_id: 1,
                    role_id: 2,
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Users,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Roles { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::SpaceMembers { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::MemberRoles { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:member_role:read_by_role_id:49a64717d5d4cb19952e6eac2946415cf6879adacf9908e7d872332d32c6e684:a5abe5690d7fd99250e47df5efcf41c1a26ccdd5b2234fd91ebc9ff74c6f6ca3",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::ApiKeyById { space_id: 1, id: 2 },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Users,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::SpaceMembers { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::ApiKeys { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:api_key:read_by_id:49a64717d5d4cb19952e6eac2946415cf6879adacf9908e7d872332d32c6e684:7c6389607512284ff56adae7e01918359bbaf76a62cd327b215827fc7d0515f4",
        ),
        (
            CacheEntryKey {
                query: CacheQuery::ApiKeyByFilter {
                    space_id: 1,
                    filter: repo::ApiKeyFilter {
                        page: 2,
                        limit: 10,
                        search: Some("private-selector".into()),
                        member_id: Some(2),
                        user_id: Some(2),
                    },
                },
                revisions: vec![
                    CacheRevision {
                        family: CacheFamily::Users,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::Spaces,
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::SpaceMembers { space_id: 1 },
                        token: "generation".into(),
                    },
                    CacheRevision {
                        family: CacheFamily::ApiKeys { space_id: 1 },
                        token: "generation".into(),
                    },
                ],
            },
            "test:cache:v1:space:1:api_key:read_by_filter:19fc125471005eddc7da9037a97fe2e4d8dea45466500fb25db2a27243ce784d:7c6389607512284ff56adae7e01918359bbaf76a62cd327b215827fc7d0515f4",
        ),
    ];
    for (mut key, expected) in cases {
        let actual = builder.entry(&context, &key).unwrap();
        assert_eq!(actual, expected);
        assert!(!actual.contains("private-selector"));
        key.revisions.reverse();
        assert_eq!(builder.entry(&context, &key).unwrap(), expected);
        let changed_context = AppContext {
            actor: Some("another actor".into()),
            trace_id: Some(uuid::Uuid::from_u128(123)),
            transaction: None,
        };
        assert_eq!(builder.entry(&changed_context, &key).unwrap(), expected);
    }
}

#[test]
fn key_validation_rejects_invalid_dependencies_ids_namespace_and_transactions() {
    assert!(matches!(
        Sha256KeyBuilder::new(""),
        Err(CachingError::BadArgs)
    ));
    assert!(matches!(
        Sha256KeyBuilder::new("invalid:namespace"),
        Err(CachingError::BadArgs)
    ));
    let builder = Sha256KeyBuilder::new("test").unwrap();
    let context = AppContext::default();
    let mut key = CacheEntryKey {
        query: CacheQuery::UserById { id: 2 },
        revisions: vec![CacheRevision {
            family: CacheFamily::Users,
            token: "generation".into(),
        }],
    };
    let first = builder.entry(&context, &key).unwrap();
    key.query = CacheQuery::UserById { id: 3 };
    assert_ne!(builder.entry(&context, &key).unwrap(), first);
    key.query = CacheQuery::UserById { id: 0 };
    assert!(matches!(
        builder.entry(&context, &key),
        Err(CachingError::BadArgs)
    ));
    key.query = CacheQuery::UserById { id: 2 };
    key.revisions.push(key.revisions[0].clone());
    assert!(matches!(
        builder.entry(&context, &key),
        Err(CachingError::BadArgs)
    ));
    key.revisions.clear();
    assert!(matches!(
        builder.entry(&context, &key),
        Err(CachingError::BadArgs)
    ));
    key.revisions.push(CacheRevision {
        family: CacheFamily::Spaces,
        token: "generation".into(),
    });
    assert!(matches!(
        builder.entry(&context, &key),
        Err(CachingError::BadArgs)
    ));
    key.revisions[0].family = CacheFamily::Users;
    key.revisions[0].token.clear();
    assert!(matches!(
        builder.entry(&context, &key),
        Err(CachingError::BadArgs)
    ));
    key.revisions[0].token = "generation".into();
    assert_ne!(
        Sha256KeyBuilder::new("other")
            .unwrap()
            .entry(&context, &key)
            .unwrap(),
        first
    );
    assert!(matches!(
        builder.revision(&context, CacheFamily::Roles { space_id: 0 }),
        Err(CachingError::BadArgs)
    ));
    let transaction = AppContext {
        transaction: Some(AppTransaction::new(())),
        ..context
    };
    assert!(matches!(
        builder.entry(&transaction, &key),
        Err(CachingError::BadState)
    ));
    assert!(matches!(
        builder.revision(&transaction, CacheFamily::Users),
        Err(CachingError::BadState)
    ));
}

#[tokio::test]
async fn initial_connection_is_bounded_and_logs_no_connection_details() {
    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let mut config = support::env();
    config.redis_port = listener.local_addr().unwrap().port();
    config.redis_password = "private-password".into();
    config.redis_connect_timeout = Duration::from_millis(30);
    let logger = Arc::new(support::RecordingLogger::default());
    let keys = Arc::new(Sha256KeyBuilder::new("test").unwrap());
    let context = AppContext::default();
    let result = lockmate::infrastructure::caching::RedisBackend::connect(
        &context,
        &config,
        keys,
        logger.clone(),
    )
    .await;
    assert!(matches!(result, Err(CachingError::Timeout { .. })));
    let entries = logger.0.lock().unwrap();
    assert_eq!(entries.len(), 1);
    assert_eq!(entries[0].level, LoggerLevel::Warn);
    assert!(!entries[0].message.contains("private-password"));
    assert!(!format!("{:?}", entries[0].meta).contains("private-password"));
}
