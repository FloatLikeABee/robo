use sqlx::MySqlPool;

use crate::permissions::{
    default_app_permissions, effective_permissions, is_admin, normalize_app_permissions,
};

async fn fetch_user_permissions(pool: &MySqlPool, user_id: &str) -> Result<Vec<String>, sqlx::Error> {
    let row: Option<(String,)> =
        sqlx::query_as("SELECT permissions FROM plat_users WHERE id = ?")
            .bind(user_id)
            .fetch_optional(pool)
            .await?;
    Ok(match row {
        Some((json,)) => serde_json::from_str(&json).unwrap_or_else(|_| default_app_permissions()),
        None => vec![],
    })
}

/// Resolves permission names for a user from stored permissions; Admin role implies all app access.
pub async fn resolve_permission_names(
    pool: &MySqlPool,
    user_id: &str,
    roles: &[String],
) -> Vec<String> {
    if is_admin(roles) {
        return effective_permissions(roles, &[]);
    }
    match fetch_user_permissions(pool, user_id).await {
        Ok(perms) => effective_permissions(roles, &normalize_app_permissions(&perms)),
        Err(e) => {
            tracing::warn!(?e, user_id, "permission DB resolve failed; using defaults");
            effective_permissions(roles, &default_app_permissions())
        }
    }
}
