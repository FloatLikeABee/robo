use axum::extract::{Path, State};
use axum::Json;
use chrono::Utc;
use serde_json::{json, Value};
use uuid::Uuid;

use crate::error::{AppError, AppResult};
use crate::extractors::JwtClaims;
use crate::jwt::Claims;
use crate::models::{CreateRoleBody, RoleRow, UpdateRoleBody};
use crate::permissions::{require_role, ROLE_ADMIN};
use crate::roles_db::{remove_role_from_all_users, rename_role_in_all_users};
use crate::state::AppState;

fn require_admin(claims: &Claims) -> AppResult<()> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }
    Ok(())
}

fn valid_role_display_name(name: &str) -> bool {
    let t = name.trim();
    !t.is_empty() && t.len() <= 128
}

pub async fn insert_role(
    pool: &sqlx::MySqlPool,
    name: &str,
    description: Option<&str>,
) -> AppResult<(String, String)> {
    let name = name.trim().to_string();
    if !valid_role_display_name(&name) {
        return Err(AppError::BadRequest(
            "name must be 1–128 characters (trimmed)".into(),
        ));
    }
    let id = Uuid::new_v4().to_string();
    let now = Utc::now().to_rfc3339();
    sqlx::query(
        r#"INSERT INTO plat_roles (id, name, description, created_at, updated_at)
           VALUES (?, ?, ?, ?, ?)"#,
    )
    .bind(&id)
    .bind(&name)
    .bind(description)
    .bind(&now)
    .bind(&now)
    .execute(pool)
    .await
    .map_err(|e: sqlx::Error| {
        let msg = e.to_string();
        if msg.contains("Duplicate") || msg.contains("duplicate") {
            return AppError::Conflict("role name already exists".into());
        }
        tracing::error!(?e, "insert role");
        AppError::Internal
    })?;
    Ok((id, name))
}

pub async fn list_roles(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<Value>> {
    require_admin(&claims)?;
    let rows: Vec<RoleRow> = sqlx::query_as("SELECT * FROM plat_roles ORDER BY name ASC")
        .fetch_all(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;
    let list: Vec<Value> = rows
        .into_iter()
        .map(|r| {
            json!({
                "id": r.id,
                "name": r.name,
                "description": r.description,
                "createdAt": r.created_at,
                "updatedAt": r.updated_at,
            })
        })
        .collect();
    Ok(Json(json!({ "roles": list })))
}

pub async fn create_role(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Json(body): Json<CreateRoleBody>,
) -> AppResult<Json<Value>> {
    require_admin(&claims)?;
    let name = body.name.trim().to_string();
    let (id, name) = insert_role(&state.pool, &name, body.description.as_deref()).await?;

    Ok(Json(json!({
        "id": id,
        "name": name,
        "description": body.description,
        "createdAt": Utc::now().to_rfc3339(),
        "updatedAt": Utc::now().to_rfc3339(),
    })))
}

pub async fn update_role(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(body): Json<UpdateRoleBody>,
) -> AppResult<Json<Value>> {
    require_admin(&claims)?;
    if body.name.is_none() && body.description.is_none() {
        return Err(AppError::BadRequest("no fields to update".into()));
    }

    let row: RoleRow = sqlx::query_as("SELECT * FROM plat_roles WHERE id = ?")
        .bind(&id)
        .fetch_optional(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?
        .ok_or(AppError::NotFound)?;

    let old_name = row.name.clone();
    let now = Utc::now().to_rfc3339();

    let final_name = if let Some(ref n) = body.name {
        let t = n.trim().to_string();
        if !valid_role_display_name(&t) {
            return Err(AppError::BadRequest(
                "name must be 1–128 characters (trimmed)".into(),
            ));
        }
        t
    } else {
        old_name.clone()
    };

    let final_desc = match body.description {
        None => row.description.clone(),
        Some(ref d) => Some(d.clone()),
    };

    if final_name != old_name {
        rename_role_in_all_users(&state.pool, &old_name, &final_name)
            .await
            .map_err(|_| AppError::Internal)?;
    }

    sqlx::query("UPDATE plat_roles SET name = ?, description = ?, updated_at = ? WHERE id = ?")
        .bind(&final_name)
        .bind(&final_desc)
        .bind(&now)
        .bind(&id)
        .execute(&state.pool)
        .await
        .map_err(|e: sqlx::Error| {
            let msg = e.to_string();
            if msg.contains("Duplicate") || msg.contains("duplicate") {
                return AppError::Conflict("role name already exists".into());
            }
            tracing::error!(?e, "update role");
            AppError::Internal
        })?;

    Ok(Json(json!({ "message": "updated" })))
}

pub async fn delete_role(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> AppResult<Json<Value>> {
    require_admin(&claims)?;
    let row: Option<(String,)> = sqlx::query_as("SELECT name FROM plat_roles WHERE id = ?")
        .bind(&id)
        .fetch_optional(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;
    let Some((role_name,)) = row else {
        return Err(AppError::NotFound);
    };

    remove_role_from_all_users(&state.pool, &role_name)
        .await
        .map_err(|_| AppError::Internal)?;

    sqlx::query("DELETE FROM plat_roles WHERE id = ?")
        .bind(&id)
        .execute(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;

    Ok(Json(json!({ "message": "deleted" })))
}
