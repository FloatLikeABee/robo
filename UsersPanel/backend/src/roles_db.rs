use chrono::Utc;
use sqlx::MySqlPool;

pub async fn fetch_role_names(pool: &MySqlPool) -> Result<Vec<String>, sqlx::Error> {
    sqlx::query_scalar::<_, String>("SELECT name FROM plat_roles ORDER BY name ASC")
        .fetch_all(pool)
        .await
}

pub async fn remove_role_from_all_users(pool: &MySqlPool, role_name: &str) -> Result<(), sqlx::Error> {
    let users: Vec<(String, String)> =
        sqlx::query_as("SELECT id, roles FROM plat_users").fetch_all(pool).await?;
    let now = Utc::now().to_rfc3339();
    for (id, roles_json) in users {
        let mut roles: Vec<String> = serde_json::from_str(&roles_json).unwrap_or_default();
        let before = roles.len();
        roles.retain(|r| r != role_name);
        if roles.len() == before {
            continue;
        }
        let j = serde_json::to_string(&roles).unwrap();
        sqlx::query("UPDATE plat_users SET roles = ?, updated_at = ? WHERE id = ?")
            .bind(&j)
            .bind(&now)
            .bind(&id)
            .execute(pool)
            .await?;
    }
    Ok(())
}

pub async fn rename_role_in_all_users(
    pool: &MySqlPool,
    old_name: &str,
    new_name: &str,
) -> Result<(), sqlx::Error> {
    let users: Vec<(String, String)> =
        sqlx::query_as("SELECT id, roles FROM plat_users").fetch_all(pool).await?;
    let now = Utc::now().to_rfc3339();
    for (id, roles_json) in users {
        let mut roles: Vec<String> = serde_json::from_str(&roles_json).unwrap_or_default();
        let mut changed = false;
        for r in &mut roles {
            if r == old_name {
                *r = new_name.to_string();
                changed = true;
            }
        }
        if !changed {
            continue;
        }
        let j = serde_json::to_string(&roles).unwrap();
        sqlx::query("UPDATE plat_users SET roles = ?, updated_at = ? WHERE id = ?")
            .bind(&j)
            .bind(&now)
            .bind(&id)
            .execute(pool)
            .await?;
    }
    Ok(())
}
