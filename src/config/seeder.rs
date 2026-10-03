use crate::domain::models::SeedData;

pub fn load() -> Result<SeedData, serde_json::Error> {
    Ok(SeedData {
        permissions: serde_json::from_str(include_str!("../../database/seeder/permission.json"))?,
        roles: serde_json::from_str(include_str!("../../database/seeder/role.json"))?,
        users: serde_json::from_str(include_str!("../../database/seeder/user.json"))?,
    })
}
