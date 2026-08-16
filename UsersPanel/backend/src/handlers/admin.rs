use axum::extract::{Path, State};
use axum::Json;
use chrono::Utc;
use serde::Deserialize;
use serde_json::json;
use uuid::Uuid;

use crate::error::{AppError, AppResult};
use crate::extractors::JwtClaims;
use crate::models::{CreateAdminUserBody, UpdateAdminUserBody, UpdateUserAccessBody, UserRow};
use crate::permissions::{
    default_app_permissions, is_admin, normalize_app_permissions, require_role, roles_for_is_admin,
    ROLE_ADMIN,
};
use crate::state::AppState;

pub async fn list_users(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<serde_json::Value>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }

    let rows: Vec<UserRow> = sqlx::query_as("SELECT * FROM plat_users ORDER BY created_at DESC")
        .fetch_all(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;

    let mut list = Vec::new();
    for r in rows {
        let roles: Vec<String> = r.roles_vec().map_err(|_| AppError::Internal)?;
        let permissions: Vec<String> = r.permissions_vec().map_err(|_| AppError::Internal)?;
        list.push(json!({
            "id": r.id,
            "email": r.email,
            "username": r.username,
            "isVerified": r.is_verified,
            "isAdmin": is_admin(&roles),
            "roles": roles,
            "permissions": permissions,
            "defaultChannelId": r.default_channel_id,
            "createdAt": r.created_at,
        }));
    }

    Ok(Json(json!({ "users": list })))
}

#[derive(Debug)]
pub struct CreateUserResult {
    pub user_id: String,
    pub morph_user_id: u64,
}

pub async fn create_admin_user(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Json(body): Json<CreateAdminUserBody>,
) -> AppResult<Json<serde_json::Value>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }

    let created = create_user_and_morph_profile(&state, body).await?;

    Ok(Json(json!({
        "message": "user created",
        "userId": created.user_id,
        "morphUserId": created.morph_user_id,
    })))
}

pub async fn update_user_access(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(user_id): Path<String>,
    Json(body): Json<UpdateUserAccessBody>,
) -> AppResult<Json<serde_json::Value>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }

    let roles = roles_for_is_admin(body.is_admin);
    let permissions = if body.is_admin {
        vec![]
    } else {
        let perms = if body.permissions.is_empty() {
            default_app_permissions()
        } else {
            normalize_app_permissions(&body.permissions)
        };
        if perms.is_empty() {
            return Err(AppError::BadRequest(
                "non-admin users need at least one app permission".into(),
            ));
        }
        perms
    };

    let roles_json = serde_json::to_string(&roles).map_err(|_| AppError::Internal)?;
    let permissions_json = serde_json::to_string(&permissions).map_err(|_| AppError::Internal)?;
    let now = Utc::now().to_rfc3339();
    let res = sqlx::query(
        "UPDATE plat_users SET roles = ?, permissions = ?, updated_at = ? WHERE id = ?",
    )
    .bind(&roles_json)
    .bind(&permissions_json)
    .bind(&now)
    .bind(&user_id)
    .execute(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;

    if res.rows_affected() == 0 {
        return Err(AppError::NotFound);
    }

    crate::public_channel::sync_user_public_channel(&state.pool, &user_id, &roles, &permissions)
        .await
        .map_err(|_| AppError::Internal)?;

    Ok(Json(json!({ "message": "access updated" })))
}

pub async fn create_user_and_morph_profile(
    state: &AppState,
    body: CreateAdminUserBody,
) -> AppResult<CreateUserResult> {
    let email = normalize_required(&body.email, "email")?;
    let username = normalize_required(&body.username, "username")?;
    if body.password.trim().len() < 8 {
        return Err(AppError::BadRequest("password min 8 chars".into()));
    }

    let last_name = normalize_required(
        body.last_name.as_deref().unwrap_or_default(),
        "last_name",
    )?;
    let first_name = normalize_optional(body.first_name.as_deref());
    let login_id = normalize_optional(body.login_id.as_deref()).or_else(|| Some(username.clone()));
    let phone = normalize_optional(body.phone.as_deref());
    let is_verified = body.is_verified.unwrap_or(true);
    let administrator = if body.administrator.unwrap_or(false) { 1 } else { 0 };

    let is_admin_flag = body.is_admin.unwrap_or(false);
    let roles = roles_for_is_admin(is_admin_flag);
    let permissions = if is_admin_flag {
        vec![]
    } else {
        let perms = body
            .permissions
            .as_ref()
            .map(|p| normalize_app_permissions(p))
            .filter(|p| !p.is_empty())
            .unwrap_or_else(default_app_permissions);
        if perms.is_empty() {
            return Err(AppError::BadRequest(
                "non-admin users need at least one app permission".into(),
            ));
        }
        perms
    };

    let dup: Option<(String,)> =
        sqlx::query_as("SELECT id FROM plat_users WHERE email = ? OR username = ?")
            .bind(&email)
            .bind(&username)
            .fetch_optional(&state.pool)
            .await
            .map_err(|_| AppError::Internal)?;
    if dup.is_some() {
        return Err(AppError::Conflict("email or username already registered".into()));
    }

    let mut tx = state.pool.begin().await.map_err(|_| AppError::Internal)?;

    let user_id = Uuid::new_v4().to_string();
    let password_hash = bcrypt::hash(body.password.trim(), 10).map_err(|_| AppError::Internal)?;
    let roles_json = serde_json::to_string(&roles).map_err(|_| AppError::Internal)?;
    let permissions_json = serde_json::to_string(&permissions).map_err(|_| AppError::Internal)?;
    let channel = format!("ch_{}", Uuid::new_v4().simple());
    let now = Utc::now().to_rfc3339();
    let verify_token = if is_verified {
        None
    } else {
        Some(Uuid::new_v4().simple().to_string())
    };
    let verify_exp = if is_verified {
        None
    } else {
        Some((Utc::now() + chrono::Duration::hours(24)).to_rfc3339())
    };

    sqlx::query(
        r#"INSERT INTO plat_users (
            id, email, username, password_hash, google_id, is_verified, roles, permissions,
            default_channel_id, verification_token, verification_expires_at,
            reset_token, reset_expires_at, created_at, updated_at
        ) VALUES (?, ?, ?, ?, NULL, ?, ?, ?, ?, ?, ?, NULL, NULL, ?, ?)"#,
    )
    .bind(&user_id)
    .bind(&email)
    .bind(&username)
    .bind(&password_hash)
    .bind(is_verified)
    .bind(&roles_json)
    .bind(&permissions_json)
    .bind(&channel)
    .bind(verify_token)
    .bind(verify_exp)
    .bind(&now)
    .bind(&now)
    .execute(&mut *tx)
    .await
    .map_err(|_| AppError::Internal)?;

    let morph_insert = sqlx::query(
        "INSERT INTO `User` (Administrator, LoginID, FirstName, LastName, Email, Phone, Title, Deactivated) VALUES (?, ?, ?, ?, ?, ?, ?, 0)",
    )
    .bind(administrator)
    .bind(login_id)
    .bind(first_name)
    .bind(last_name)
    .bind(Some(email))
    .bind(phone)
    .bind(None::<String>)
    .execute(&mut *tx)
    .await
    .map_err(|_| AppError::Internal)?;

    tx.commit().await.map_err(|_| AppError::Internal)?;

    crate::public_channel::sync_user_public_channel(&state.pool, &user_id, &roles, &permissions)
        .await
        .map_err(|_| AppError::Internal)?;

    Ok(CreateUserResult {
        user_id,
        morph_user_id: morph_insert.last_insert_id(),
    })
}

fn normalize_required(value: &str, field: &str) -> AppResult<String> {
    let trimmed = value.trim();
    if trimmed.is_empty() {
        return Err(AppError::BadRequest(format!("{field} is required")));
    }
    Ok(trimmed.to_string())
}

fn normalize_optional(value: Option<&str>) -> Option<String> {
    value.and_then(|v| {
        let trimmed = v.trim();
        if trimmed.is_empty() {
            None
        } else {
            Some(trimmed.to_string())
        }
    })
}

pub async fn update_admin_user(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(user_id): Path<String>,
    Json(body): Json<UpdateAdminUserBody>,
) -> AppResult<Json<serde_json::Value>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }
    if body.email.is_none() && body.username.is_none() {
        return Err(AppError::BadRequest(
            "provide at least one of email, username".into(),
        ));
    }

    let row: UserRow = sqlx::query_as("SELECT * FROM plat_users WHERE id = ?")
        .bind(&user_id)
        .fetch_optional(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?
        .ok_or(AppError::NotFound)?;

    let email = body.email.clone().unwrap_or_else(|| row.email.clone());
    let username = body.username.clone().unwrap_or_else(|| row.username.clone());

    if body.email.is_some() {
        let dup: Option<(String,)> =
            sqlx::query_as("SELECT id FROM plat_users WHERE email = ? AND id != ?")
                .bind(&email)
                .bind(&user_id)
                .fetch_optional(&state.pool)
                .await
                .map_err(|_| AppError::Internal)?;
        if dup.is_some() {
            return Err(AppError::Conflict("email already in use".into()));
        }
    }
    if body.username.is_some() {
        let dup: Option<(String,)> =
            sqlx::query_as("SELECT id FROM plat_users WHERE username = ? AND id != ?")
                .bind(&username)
                .bind(&user_id)
                .fetch_optional(&state.pool)
                .await
                .map_err(|_| AppError::Internal)?;
        if dup.is_some() {
            return Err(AppError::Conflict("username already in use".into()));
        }
    }

    let now = Utc::now().to_rfc3339();
    sqlx::query("UPDATE plat_users SET email = ?, username = ?, updated_at = ? WHERE id = ?")
        .bind(&email)
        .bind(&username)
        .bind(&now)
        .bind(&user_id)
        .execute(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;

    Ok(Json(json!({ "message": "user updated" })))
}

pub async fn delete_admin_user(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(user_id): Path<String>,
) -> AppResult<Json<serde_json::Value>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }
    if claims.sub == user_id {
        return Err(AppError::BadRequest(
            "cannot delete your own account".into(),
        ));
    }

    let res = sqlx::query("DELETE FROM plat_users WHERE id = ?")
        .bind(&user_id)
        .execute(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;
    if res.rows_affected() == 0 {
        return Err(AppError::NotFound);
    }
    Ok(Json(json!({ "message": "user deleted" })))
}

#[derive(Debug, sqlx::FromRow)]
struct MorphUserDbRow {
    user_id: i32,
    login_id: Option<String>,
    first_name: Option<String>,
    last_name: String,
    email: Option<String>,
    phone: Option<String>,
    administrator: i8,
}

pub async fn list_morph_users(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<serde_json::Value>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }

    let rows: Vec<MorphUserDbRow> = sqlx::query_as(
        "SELECT UserID AS user_id, LoginID AS login_id, FirstName AS first_name, LastName AS last_name, Email AS email, Phone AS phone, Administrator AS administrator FROM `User` WHERE Deactivated = 0 ORDER BY LastName, FirstName LIMIT 500",
    )
    .fetch_all(&state.pool)
    .await
    .map_err(|e| {
        tracing::error!("list_morph_users query failed: {e}");
        AppError::Internal
    })?;

    let users: Vec<serde_json::Value> = rows
        .into_iter()
        .map(|r| {
            json!({
                "id": r.user_id,
                "loginId": r.login_id,
                "firstName": r.first_name,
                "lastName": r.last_name,
                "email": r.email,
                "phone": r.phone,
                "administrator": r.administrator != 0,
            })
        })
        .collect();

    Ok(Json(json!({ "users": users })))
}

#[derive(Debug, Deserialize)]
pub struct UpdateMorphUserBody {
    pub first_name: Option<String>,
    pub last_name: Option<String>,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub administrator: Option<bool>,
}

pub async fn update_morph_user(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(user_id): Path<i32>,
    Json(body): Json<UpdateMorphUserBody>,
) -> AppResult<Json<serde_json::Value>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }

    if body.first_name.is_none()
        && body.last_name.is_none()
        && body.email.is_none()
        && body.phone.is_none()
        && body.administrator.is_none()
    {
        return Err(AppError::BadRequest("no fields to update".into()));
    }

    if let Some(ref ln) = body.last_name {
        if ln.trim().is_empty() {
            return Err(AppError::BadRequest("last_name cannot be empty".into()));
        }
    }

    let mut sets: Vec<String> = Vec::new();
    if body.first_name.is_some() {
        sets.push("FirstName = ?".into());
    }
    if body.last_name.is_some() {
        sets.push("LastName = ?".into());
    }
    if body.email.is_some() {
        sets.push("Email = ?".into());
    }
    if body.phone.is_some() {
        sets.push("Phone = ?".into());
    }
    if body.administrator.is_some() {
        sets.push("Administrator = ?".into());
    }

    let q = format!(
        "UPDATE `User` SET {} WHERE UserID = ? AND Deactivated = 0",
        sets.join(", ")
    );
    let mut query = sqlx::query(&q);
    if let Some(v) = &body.first_name {
        query = query.bind(normalize_optional(Some(v.as_str())));
    }
    if let Some(v) = &body.last_name {
        query = query.bind(v.trim());
    }
    if let Some(v) = &body.email {
        query = query.bind(normalize_optional(Some(v.as_str())));
    }
    if let Some(v) = &body.phone {
        query = query.bind(normalize_optional(Some(v.as_str())));
    }
    if let Some(v) = body.administrator {
        query = query.bind(if v { 1i8 } else { 0i8 });
    }
    query = query.bind(user_id);

    let res = query.execute(&state.pool).await.map_err(|_| AppError::Internal)?;
    if res.rows_affected() == 0 {
        return Err(AppError::NotFound);
    }

    Ok(Json(json!({ "message": "profile updated" })))
}

pub async fn deactivate_morph_user(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(user_id): Path<i32>,
) -> AppResult<Json<serde_json::Value>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }

    let now = Utc::now().format("%Y-%m-%d %H:%M:%S").to_string();
    let res = sqlx::query(
        "UPDATE `User` SET Deactivated = 1, DeactivatedDate = ? WHERE UserID = ? AND Deactivated = 0",
    )
    .bind(&now)
    .bind(user_id)
    .execute(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;

    if res.rows_affected() == 0 {
        return Err(AppError::NotFound);
    }

    Ok(Json(json!({ "message": "user deactivated", "id": user_id })))
}
