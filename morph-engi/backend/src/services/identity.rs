use crate::db::DbPool;
use crate::services::users_panel::map_engi_role;
use sqlx::Row;

pub struct Identity {
    pub user_id: i64,
    pub org_id: i64,
    pub role: String,
}

pub async fn ensure_platform_identity(
    pool: &DbPool,
    email: &str,
    display_name: &str,
    platform_roles: &[String],
) -> Result<Identity, sqlx::Error> {
    let email = email.trim();
    let existing = sqlx::query(
        "SELECT id, organization_id, role FROM users WHERE email = ? AND is_active = 1",
    )
    .bind(email)
    .fetch_optional(pool)
    .await?;

    if let Some(row) = existing {
        return Ok(Identity {
            user_id: row.get("id"),
            org_id: row.get("organization_id"),
            role: row.get("role"),
        });
    }

    let role = map_engi_role(platform_roles).to_string();
    let org_label = if display_name.is_empty() || display_name == email {
        "Engineering Organization".to_string()
    } else {
        format!("{display_name}'s Projects")
    };

    let mut tx = pool.begin().await?;
    let org_res = sqlx::query("INSERT INTO organizations (name) VALUES (?)")
        .bind(&org_label)
        .execute(&mut *tx)
        .await?;
    let org_id = org_res.last_insert_rowid();

    let user_res = sqlx::query(
        "INSERT INTO users (organization_id, name, email, role) VALUES (?, ?, ?, ?)",
    )
    .bind(org_id)
    .bind(display_name)
    .bind(email)
    .bind(&role)
    .execute(&mut *tx)
    .await?;
    let user_id = user_res.last_insert_rowid();
    tx.commit().await?;

    Ok(Identity {
        user_id,
        org_id,
        role,
    })
}
