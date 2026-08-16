use axum::extract::{Path, State};
use axum::Json;
use chrono::Utc;
use serde_json::{json, Value};
use uuid::Uuid;

use crate::error::{AppError, AppResult};
use crate::extractors::JwtClaims;
use crate::jwt::Claims;
use crate::models::{
    CreatePermissionBody, PermissionRow, PutRolePermissionsBody, UpdatePermissionBody,
};
use crate::permissions::{require_role, ROLE_ADMIN};
use crate::roles_db::fetch_role_names;
use crate::state::AppState;

fn require_admin(claims: &Claims) -> AppResult<()> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }
    Ok(())
}

fn valid_permission_slug(name: &str) -> bool {
    let mut chars = name.chars();
    match chars.next() {
        Some(c) if c.is_ascii_lowercase() => {}
        _ => return false,
    }
    name.len() >= 2
        && name.len() <= 128
        && name.chars().all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '_')
}

pub async fn list_permissions(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<Value>> {
    require_admin(&claims)?;
    let rows: Vec<PermissionRow> = sqlx::query_as("SELECT * FROM plat_permissions ORDER BY name ASC")
        .fetch_all(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;
    let list: Vec<Value> = rows
        .into_iter()
        .map(|p| {
            json!({
                "id": p.id,
                "name": p.name,
                "description": p.description,
                "createdAt": p.created_at,
                "updatedAt": p.updated_at,
            })
        })
        .collect();
    Ok(Json(json!({ "permissions": list })))
}

pub async fn permissions_overview(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<Value>> {
    require_admin(&claims)?;
    let permissions: Vec<PermissionRow> =
        sqlx::query_as("SELECT * FROM plat_permissions ORDER BY name ASC")
            .fetch_all(&state.pool)
            .await
            .map_err(|_| AppError::Internal)?;

    let rows: Vec<(String, String)> = sqlx::query_as(
        "SELECT role_name, permission_id FROM plat_role_permissions ORDER BY role_name, permission_id",
    )
    .fetch_all(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;

    let role_names = fetch_role_names(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;
    let mut assignments: std::collections::HashMap<String, Vec<String>> =
        std::collections::HashMap::new();
    for r in &role_names {
        assignments.insert(r.clone(), vec![]);
    }
    for (role_name, perm_id) in rows {
        assignments.entry(role_name).or_default().push(perm_id);
    }

    let perm_json: Vec<Value> = permissions
        .into_iter()
        .map(|p| {
            json!({
                "id": p.id,
                "name": p.name,
                "description": p.description,
                "createdAt": p.created_at,
                "updatedAt": p.updated_at,
            })
        })
        .collect();

    Ok(Json(json!({
        "permissions": perm_json,
        "roleNames": role_names,
        "assignments": assignments,
    })))
}

pub async fn create_permission(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Json(body): Json<CreatePermissionBody>,
) -> AppResult<Json<Value>> {
    require_admin(&claims)?;
    if !valid_permission_slug(&body.name) {
        return Err(AppError::BadRequest(
            "name must be 2–128 chars: start with a letter, then [a-z0-9_]".into(),
        ));
    }
    let id = Uuid::new_v4().to_string();
    let now = Utc::now().to_rfc3339();
    let desc = body.description.as_deref();
    sqlx::query(
        r#"INSERT INTO plat_permissions (id, name, description, created_at, updated_at)
           VALUES (?, ?, ?, ?, ?)"#,
    )
    .bind(&id)
    .bind(&body.name)
    .bind(desc)
    .bind(&now)
    .bind(&now)
    .execute(&state.pool)
    .await
    .map_err(|e: sqlx::Error| {
        let msg = e.to_string();
        if msg.contains("Duplicate") || msg.contains("duplicate") {
            return AppError::Conflict("permission name already exists".into());
        }
        tracing::error!(?e, "insert permission");
        AppError::Internal
    })?;

    Ok(Json(json!({
        "id": id,
        "name": body.name,
        "description": body.description,
        "createdAt": now,
        "updatedAt": now,
    })))
}

pub async fn update_permission(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(body): Json<UpdatePermissionBody>,
) -> AppResult<Json<Value>> {
    require_admin(&claims)?;
    let now = Utc::now().to_rfc3339();
    let row = sqlx::query(
        "UPDATE plat_permissions SET description = ?, updated_at = ? WHERE id = ?",
    )
    .bind(&body.description)
    .bind(&now)
    .bind(&id)
    .execute(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;

    if row.rows_affected() == 0 {
        return Err(AppError::NotFound);
    }

    Ok(Json(json!({ "message": "updated" })))
}

pub async fn delete_permission(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> AppResult<Json<Value>> {
    require_admin(&claims)?;
    let res = sqlx::query("DELETE FROM plat_permissions WHERE id = ?")
        .bind(&id)
        .execute(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;
    if res.rows_affected() == 0 {
        return Err(AppError::NotFound);
    }
    Ok(Json(json!({ "message": "deleted" })))
}

pub async fn put_role_permissions(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(role_name): Path<String>,
    Json(body): Json<PutRolePermissionsBody>,
) -> AppResult<Json<Value>> {
    require_admin(&claims)?;
    let allowed: std::collections::HashSet<String> = fetch_role_names(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?
        .into_iter()
        .collect();
    if !allowed.contains(&role_name) {
        return Err(AppError::BadRequest("unknown role".into()));
    }

    for pid in &body.permission_ids {
        let exists: Option<(String,)> = sqlx::query_as("SELECT id FROM plat_permissions WHERE id = ?")
            .bind(pid)
            .fetch_optional(&state.pool)
            .await
            .map_err(|_| AppError::Internal)?;
        if exists.is_none() {
            return Err(AppError::BadRequest(format!("unknown permission id: {pid}")));
        }
    }

    let mut tx = state.pool.begin().await.map_err(|_| AppError::Internal)?;
    sqlx::query("DELETE FROM plat_role_permissions WHERE role_name = ?")
        .bind(&role_name)
        .execute(&mut *tx)
        .await
        .map_err(|_| AppError::Internal)?;

    for pid in &body.permission_ids {
        sqlx::query(
            "INSERT INTO plat_role_permissions (role_name, permission_id) VALUES (?, ?)",
        )
        .bind(&role_name)
        .bind(pid)
        .execute(&mut *tx)
        .await
        .map_err(|_| AppError::Internal)?;
    }

    tx.commit().await.map_err(|_| AppError::Internal)?;
    Ok(Json(json!({ "message": "role permissions updated" })))
}
