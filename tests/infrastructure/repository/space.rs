use super::support::Fixture;
use lockmate::domain::{
    contracts::repository::{CreateSpace, SpaceFilter, UpdateSpace},
    models::RepositoryError,
};
use serde_json::json;

#[tokio::test]
#[ignore = "requires isolated PostgreSQL via LOCKMATE_TEST_DATABASE_URL"]
async fn space_crud_slugs_display_names_activity_and_audits() {
    let f = Fixture::new().await;
    let id = f.space("taskmate").await;
    let space = f.spaces.read_by_slug(&f.context, "taskmate").await.unwrap();
    assert_eq!(space.id, id);
    assert!(space.is_active);
    assert_eq!(space.preferences, json!({}));
    assert_eq!(space.audit.create.by, Some(42));
    f.spaces
        .update_by_id(
            &f.context,
            id,
            UpdateSpace {
                name: Some("Task management".into()),
                description: Some("Project".into()),
                is_active: Some(false),
                preferences: Some(json!({"theme":"dark"})),
                by: Some(43),
            },
        )
        .await
        .unwrap();
    let space = f.spaces.read_by_id(&f.context, id).await.unwrap();
    assert_eq!(space.slug, "taskmate");
    assert_eq!(space.name, "Task management");
    assert!(!space.is_active);
    assert_eq!(space.audit.update.by, Some(43));
    let (spaces, total) = f
        .spaces
        .read_by_filter(
            &f.context,
            SpaceFilter {
                page: 1,
                limit: 10,
                search: Some("TASKMATE".into()),
                is_active: Some(false),
            },
        )
        .await
        .unwrap();
    assert_eq!(total, 1);
    assert_eq!(spaces[0].id, id);
    let duplicate = || CreateSpace {
        slug: "taskmate".into(),
        name: "Another label".into(),
        description: None,
        is_active: None,
        by: None,
    };
    assert!(matches!(
        f.spaces.create(&f.context, duplicate()).await,
        Err(RepositoryError::SpaceSlugConflict)
    ));
    f.spaces
        .delete_by_id(&f.context, id, Some(44))
        .await
        .unwrap();
    assert!(matches!(
        f.spaces.read_by_slug(&f.context, "taskmate").await,
        Err(RepositoryError::SpaceNotFound)
    ));
    assert!(matches!(
        f.spaces
            .update_by_id(&f.context, id, UpdateSpace::default())
            .await,
        Err(RepositoryError::SpaceNotFound)
    ));
    assert!(matches!(
        f.spaces.delete_by_id(&f.context, id, None).await,
        Err(RepositoryError::SpaceNotFound)
    ));
    assert!(matches!(
        f.spaces.create(&f.context, duplicate()).await,
        Err(RepositoryError::SpaceSlugConflict)
    ));
    f.close().await;
}
