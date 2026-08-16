use chrono::Utc;
use sqlx::MySqlPool;
use uuid::Uuid;

use crate::permissions::{effective_permissions, PERM_INBOX_MESSAGE};

pub const PUBLIC_CHANNEL_THREAD_ID: &str = "00000000-0000-4000-8000-00000000c001";
pub const AI_SENDER_USER_ID: &str = "00000000-0000-4000-8000-00000000a001";

pub async fn ensure_platform_setup(pool: &MySqlPool) -> Result<(), sqlx::Error> {
    ensure_ai_sender_user(pool).await?;
    ensure_public_channel_thread(pool).await?;
    sync_public_channel_members(pool).await?;
    Ok(())
}

async fn ensure_ai_sender_user(pool: &MySqlPool) -> Result<(), sqlx::Error> {
    let exists: Option<(String,)> =
        sqlx::query_as("SELECT id FROM plat_users WHERE id = ?")
            .bind(AI_SENDER_USER_ID)
            .fetch_optional(pool)
            .await?;
    if exists.is_some() {
        return Ok(());
    }

    let now = Utc::now().to_rfc3339();
    let channel = format!("ch_{}", Uuid::new_v4().simple());
    let roles = "[]".to_string();
    let permissions = "[]".to_string();
    // Unusable password hash — this account is system-only.
    let password_hash = "$2b$10$platform.message.ai.nologin.hash.placeholder.value";

    sqlx::query(
        r#"INSERT INTO plat_users (
            id, email, username, password_hash, google_id, is_verified, roles, permissions,
            default_channel_id, verification_token, verification_expires_at,
            reset_token, reset_expires_at, created_at, updated_at
        ) VALUES (?, ?, ?, ?, NULL, 1, ?, ?, ?, NULL, NULL, NULL, NULL, ?, ?)"#,
    )
    .bind(AI_SENDER_USER_ID)
    .bind("message-ai@platform.internal")
    .bind("message-ai")
    .bind(password_hash)
    .bind(&roles)
    .bind(&permissions)
    .bind(&channel)
    .bind(&now)
    .bind(&now)
    .execute(pool)
    .await?;

    tracing::info!("Created platform Message AI system user");
    Ok(())
}

async fn ensure_public_channel_thread(pool: &MySqlPool) -> Result<(), sqlx::Error> {
    let exists: Option<(String,)> =
        sqlx::query_as("SELECT id FROM msg_threads WHERE id = ?")
            .bind(PUBLIC_CHANNEL_THREAD_ID)
            .fetch_optional(pool)
            .await?;
    if exists.is_some() {
        sqlx::query(
            "UPDATE msg_threads SET title = ?, is_group = 1, is_public_channel = 1 WHERE id = ?",
        )
        .bind("Public Channel")
        .bind(PUBLIC_CHANNEL_THREAD_ID)
        .execute(pool)
        .await?;
        return Ok(());
    }

    let now = Utc::now().to_rfc3339();
    sqlx::query(
        r#"INSERT INTO msg_threads (id, title, created_by, is_group, is_public_channel, created_at, updated_at)
           VALUES (?, ?, ?, 1, 1, ?, ?)"#,
    )
    .bind(PUBLIC_CHANNEL_THREAD_ID)
    .bind("Public Channel")
    .bind(AI_SENDER_USER_ID)
    .bind(&now)
    .bind(&now)
    .execute(pool)
    .await?;

    tracing::info!("Created platform public messaging channel");
    Ok(())
}

fn user_has_inbox_access(roles: &[String], permissions: &[String]) -> bool {
    effective_permissions(roles, permissions)
        .iter()
        .any(|p| p == PERM_INBOX_MESSAGE)
}

pub async fn sync_public_channel_members(pool: &MySqlPool) -> Result<(), sqlx::Error> {
    let users: Vec<(String, String, String)> =
        sqlx::query_as("SELECT id, roles, permissions FROM plat_users ORDER BY created_at ASC")
            .fetch_all(pool)
            .await?;

    let now = Utc::now().to_rfc3339();
    for (user_id, roles_json, permissions_json) in users {
        if user_id == AI_SENDER_USER_ID {
            continue;
        }
        let roles: Vec<String> = serde_json::from_str(&roles_json).unwrap_or_default();
        let permissions: Vec<String> = serde_json::from_str(&permissions_json).unwrap_or_default();
        if !user_has_inbox_access(&roles, &permissions) {
            continue;
        }
        sqlx::query(
            r#"INSERT IGNORE INTO msg_thread_members (thread_id, user_id, joined_at)
               VALUES (?, ?, ?)"#,
        )
        .bind(PUBLIC_CHANNEL_THREAD_ID)
        .bind(&user_id)
        .bind(&now)
        .execute(pool)
        .await?;
    }
    Ok(())
}

pub async fn sync_user_public_channel(
    pool: &MySqlPool,
    user_id: &str,
    roles: &[String],
    permissions: &[String],
) -> Result<(), sqlx::Error> {
    if user_id == AI_SENDER_USER_ID {
        return Ok(());
    }
    if !user_has_inbox_access(roles, permissions) {
        sqlx::query("DELETE FROM msg_thread_members WHERE thread_id = ? AND user_id = ?")
            .bind(PUBLIC_CHANNEL_THREAD_ID)
            .bind(user_id)
            .execute(pool)
            .await?;
        return Ok(());
    }
    let now = Utc::now().to_rfc3339();
    sqlx::query(
        r#"INSERT IGNORE INTO msg_thread_members (thread_id, user_id, joined_at)
           VALUES (?, ?, ?)"#,
    )
    .bind(PUBLIC_CHANNEL_THREAD_ID)
    .bind(user_id)
    .bind(&now)
    .execute(pool)
    .await?;
    Ok(())
}

pub async fn is_public_channel(pool: &MySqlPool, thread_id: &str) -> Result<bool, sqlx::Error> {
    let row: Option<(i8,)> = sqlx::query_as(
        "SELECT is_public_channel FROM msg_threads WHERE id = ?",
    )
    .bind(thread_id)
    .fetch_optional(pool)
    .await?;
    Ok(row.map(|(v,)| v == 1).unwrap_or(false))
}
