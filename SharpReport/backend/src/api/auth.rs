use axum::{
    Json,
    extract::State,
    http::{HeaderMap, StatusCode, header::AUTHORIZATION},
};
use serde::{Deserialize, Serialize};
use serde_json::json;

use crate::services::AppState;
use crate::services::users_panel::{UsersPanelError, resolve_role};

type AuthErrorResponse = (StatusCode, Json<serde_json::Value>);

fn bearer_token(headers: &HeaderMap) -> Option<String> {
    let v = headers.get(AUTHORIZATION)?.to_str().ok()?;
    let rest = v
        .strip_prefix("Bearer ")
        .or_else(|| v.strip_prefix("bearer "))?;
    let t = rest.trim();
    if t.is_empty() {
        return None;
    }
    Some(t.to_string())
}

#[derive(Debug, Deserialize)]
pub struct LoginRequest {
    pub email: String,
    pub password: String,
}

#[derive(Debug, Serialize)]
pub struct LoginResponse {
    pub token: String,
    pub user: UserResponse,
    pub permissions: Vec<String>,
}

#[derive(Debug, Serialize)]
pub struct UserResponse {
    pub id: String,
    pub email: String,
    pub name: String,
    pub role: String,
    pub avatar_url: Option<String>,
    pub roles: Vec<String>,
}

fn map_user(email: &str, username: &str, roles: &[String]) -> UserResponse {
    UserResponse {
        id: email.to_string(),
        email: email.to_string(),
        name: username.to_string(),
        role: resolve_role(roles),
        avatar_url: None,
        roles: roles.to_vec(),
    }
}

fn users_panel_status(err: UsersPanelError) -> AuthErrorResponse {
    let (status, msg) = match err {
        UsersPanelError::InvalidCredentials => {
            (StatusCode::UNAUTHORIZED, "Invalid email or password")
        }
        UsersPanelError::Forbidden => (
            StatusCode::FORBIDDEN,
            "Sharp Reports access is restricted by admin policy",
        ),
        UsersPanelError::Unavailable => (StatusCode::BAD_GATEWAY, "UsersPanel unavailable"),
        UsersPanelError::BadResponse => (
            StatusCode::BAD_GATEWAY,
            "Invalid auth response from UsersPanel",
        ),
    };
    (status, Json(json!({ "error": msg })))
}

pub async fn login(
    State(state): State<AppState>,
    Json(request): Json<LoginRequest>,
) -> Result<Json<LoginResponse>, AuthErrorResponse> {
    let session = match state
        .users_panel
        .login(&request.email, &request.password)
        .await
    {
        Ok(s) => s,
        Err(e) => return Err(users_panel_status(e)),
    };

    Ok(Json(LoginResponse {
        token: session.token,
        user: map_user(
            &session.user.email,
            &session.user.username,
            &session.user.roles,
        ),
        permissions: session.permissions,
    }))
}

pub async fn me(
    State(state): State<AppState>,
    headers: HeaderMap,
) -> Result<Json<UserResponse>, AuthErrorResponse> {
    let token = bearer_token(&headers).ok_or((
        StatusCode::UNAUTHORIZED,
        Json(json!({ "error": "Missing Authorization bearer token" })),
    ))?;

    let session = match state.users_panel.session(&token).await {
        Ok(s) => s,
        Err(e) => return Err(users_panel_status(e)),
    };

    Ok(Json(map_user(
        &session.user.email,
        &session.user.username,
        &session.user.roles,
    )))
}

pub async fn logout() -> Result<(), (StatusCode, String)> {
    Ok(())
}
