use axum::{
    extract::State,
    http::StatusCode,
    Json,
};
use serde::Deserialize;
use std::sync::Arc;

use crate::api::extract::{json_err, AuthUser};
use crate::services::identity::ensure_platform_identity;
use crate::services::users_panel::UsersPanelError;
use crate::services::AppState;

#[derive(Deserialize)]
pub struct LoginBody {
    pub email: String,
    pub password: String,
}

#[derive(Deserialize)]
pub struct PlatformSessionBody {
    pub token: Option<String>,
}

fn token_response(state: &AppState, user_id: i64, org_id: i64, role: &str) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    let access = state
        .jwt
        .issue(user_id, org_id, role)
        .map_err(|_| json_err("token issue failed"))?;
    Ok(Json(serde_json::json!({
        "access_token": access,
        "token_type": "Bearer",
        "user_id": user_id,
        "organization_id": org_id,
        "role": role,
    })))
}

pub async fn login(
    State(state): State<Arc<AppState>>,
    Json(body): Json<LoginBody>,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    let session = state
        .users_panel
        .login(&body.email, &body.password)
        .await
        .map_err(|e| match e {
            UsersPanelError::Forbidden => (
                StatusCode::FORBIDDEN,
                Json(serde_json::json!({"error":"Morph Engi access is restricted by admin policy"})),
            ),
            UsersPanelError::InvalidCredentials => (
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({"error":"Invalid email or password"})),
            ),
            UsersPanelError::Unavailable(msg) => (
                StatusCode::SERVICE_UNAVAILABLE,
                Json(serde_json::json!({"error": msg})),
            ),
        })?;

    let display = if session.user.username.is_empty() {
        session.user.email.clone()
    } else {
        session.user.username.clone()
    };
    let id = ensure_platform_identity(
        &state.pool,
        &session.user.email,
        &display,
        &session.user.roles,
    )
    .await
    .map_err(|e| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({"error": e.to_string()})),
        )
    })?;

    token_response(&state, id.user_id, id.org_id, &id.role)
}

pub async fn platform_session(
    State(state): State<Arc<AppState>>,
    headers: axum::http::HeaderMap,
    Json(body): Json<PlatformSessionBody>,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    let token = body
        .token
        .filter(|t| !t.is_empty())
        .or_else(|| {
            headers
                .get(axum::http::header::AUTHORIZATION)
                .and_then(|v| v.to_str().ok())
                .and_then(|h| h.strip_prefix("Bearer "))
                .map(str::trim)
                .map(str::to_string)
        })
        .filter(|t| !t.is_empty())
        .ok_or(json_err("missing token"))?;

    let session = state.users_panel.session(&token).await.map_err(|e| match e {
        UsersPanelError::Forbidden => (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error":"Morph Engi access is restricted"})),
        ),
        UsersPanelError::Unavailable(msg) => (
            StatusCode::SERVICE_UNAVAILABLE,
            Json(serde_json::json!({"error": msg})),
        ),
        _ => (
            StatusCode::UNAUTHORIZED,
            Json(serde_json::json!({"error":"invalid platform session"})),
        ),
    })?;

    let display = if session.user.username.is_empty() {
        session.user.email.clone()
    } else {
        session.user.username.clone()
    };
    let id = ensure_platform_identity(
        &state.pool,
        &session.user.email,
        &display,
        &session.user.roles,
    )
    .await
    .map_err(|e| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({"error": e.to_string()})),
        )
    })?;

    token_response(&state, id.user_id, id.org_id, &id.role)
}

pub async fn dev_login(
    State(state): State<Arc<AppState>>,
    headers: axum::http::HeaderMap,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    if !state.settings.is_development() {
        return Err((
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error":"not found"})),
        ));
    }
    let cookie = headers
        .get(axum::http::header::COOKIE)
        .and_then(|v| v.to_str().ok())
        .unwrap_or("");
    let token = cookie
        .split(';')
        .find_map(|p| {
            let p = p.trim();
            p.strip_prefix("userspanel_session_token=")
                .or_else(|| p.strip_prefix("userspanel_session="))
        })
        .map(str::trim)
        .filter(|t| !t.is_empty());

    let token = token.ok_or(json_err(
        "missing userspanel_session_token cookie — sign in to Morph AI first (http://localhost:3031)",
    ))?
    .to_string();

    platform_session(
        State(state),
        axum::http::HeaderMap::new(),
        Json(PlatformSessionBody {
            token: Some(token),
        }),
    )
    .await
}

pub async fn me(AuthUser(auth): AuthUser) -> Json<serde_json::Value> {
    Json(serde_json::json!({
        "user_id": auth.user_id,
        "organization_id": auth.org_id,
        "role": auth.role,
    }))
}

pub async fn get_organization(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    let row = sqlx::query(
        "SELECT id, name, country, currency FROM organizations WHERE id = ?",
    )
    .bind(auth.org_id)
    .fetch_optional(&state.pool)
    .await
    .map_err(|e| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({"error": e.to_string()})),
        )
    })?;

    let Some(row) = row else {
        return Err(json_err("organization not found"));
    };

    Ok(Json(serde_json::json!({
        "id": row.get::<i64, _>("id"),
        "name": row.get::<String, _>("name"),
        "country": row.get::<String, _>("country"),
        "currency": row.get::<String, _>("currency"),
    })))
}

use sqlx::Row;

#[derive(Deserialize)]
pub struct PatchOrgBody {
    pub name: Option<String>,
    pub country: Option<String>,
    pub currency: Option<String>,
}

pub async fn patch_organization(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<PatchOrgBody>,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    if let Some(name) = &body.name {
        sqlx::query("UPDATE organizations SET name = ? WHERE id = ?")
            .bind(name)
            .bind(auth.org_id)
            .execute(&state.pool)
            .await
            .ok();
    }
    if let Some(country) = &body.country {
        sqlx::query("UPDATE organizations SET country = ? WHERE id = ?")
            .bind(country)
            .bind(auth.org_id)
            .execute(&state.pool)
            .await
            .ok();
    }
    if let Some(currency) = &body.currency {
        sqlx::query("UPDATE organizations SET currency = ? WHERE id = ?")
            .bind(currency)
            .bind(auth.org_id)
            .execute(&state.pool)
            .await
            .ok();
    }
    get_organization(State(state), AuthUser(auth)).await
}
