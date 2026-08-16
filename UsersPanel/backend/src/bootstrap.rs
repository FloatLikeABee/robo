use bcrypt::hash;
use chrono::Utc;
use sqlx::MySqlPool;
use uuid::Uuid;

use crate::config::Config;
use crate::permissions::roles_for_is_admin;

pub async fn ensure_bootstrap_admin(pool: &MySqlPool, config: &Config) -> Result<(), sqlx::Error> {
    let Some(ref email) = config.bootstrap_admin_email else {
        return Ok(());
    };
    let Some(ref password) = config.bootstrap_admin_password else {
        tracing::warn!("BOOTSTRAP_ADMIN_EMAIL set but BOOTSTRAP_ADMIN_PASSWORD missing");
        return Ok(());
    };

    let row: Option<(String,)> = sqlx::query_as("SELECT id FROM plat_users WHERE email = ?")
        .bind(email)
        .fetch_optional(pool)
        .await?;

    if row.is_some() {
        return Ok(());
    }

    let id = Uuid::new_v4().to_string();
    let username = "admin";
    let password_hash = hash(password, 10).expect("bcrypt");
    let roles = serde_json::to_string(&roles_for_is_admin(true)).unwrap();
    let permissions = serde_json::to_string(&Vec::<String>::new()).unwrap();
    let channel = format!("ch_{}", Uuid::new_v4().simple());
    let now = Utc::now().to_rfc3339();

    sqlx::query(
        r#"INSERT INTO plat_users (
            id, email, username, password_hash, google_id, is_verified, roles, permissions,
            default_channel_id, verification_token, verification_expires_at,
            reset_token, reset_expires_at, created_at, updated_at
        ) VALUES (?, ?, ?, ?, NULL, 1, ?, ?, ?, NULL, NULL, NULL, NULL, ?, ?)"#,
    )
    .bind(&id)
    .bind(email)
    .bind(username)
    .bind(&password_hash)
    .bind(&roles)
    .bind(&permissions)
    .bind(&channel)
    .bind(&now)
    .bind(&now)
    .execute(pool)
    .await?;

    tracing::info!(email = %email, "Bootstrap admin user created");
    Ok(())
}
