use axum::extract::{Query, State};
use axum::response::{IntoResponse, Redirect, Response};
use axum::Json;
use chrono::{Duration, Utc};
use oauth2::{AuthorizationCode, CsrfToken, PkceCodeChallenge, Scope, TokenResponse};
use rand::RngCore;
use serde::Deserialize;
use serde_json::json;
use uuid::Uuid;

use crate::error::{AppError, AppResult};
use crate::extractors::JwtClaims;
use crate::jwt::encode_token;
use crate::models::{
    ForgotPasswordBody, LoginBody, RegisterBody, ResetPasswordBody, UserRow,
};
use crate::permissions_resolve::resolve_permission_names;
use crate::state::AppState;

#[derive(Deserialize)]
pub struct VerifyEmailQuery {
    pub token: String,
}

#[derive(Deserialize)]
pub struct GoogleCallbackQuery {
    pub code: Option<String>,
    pub state: Option<String>,
    pub error: Option<String>,
}

fn random_token_bytes() -> String {
    let mut buf = [0u8; 32];
    rand::thread_rng().fill_bytes(&mut buf);
    base64::Engine::encode(&base64::engine::general_purpose::URL_SAFE_NO_PAD, buf)
}

pub async fn register(
    State(state): State<AppState>,
    Json(body): Json<RegisterBody>,
) -> AppResult<Json<serde_json::Value>> {
    if body.email.is_empty() || body.username.is_empty() || body.password.len() < 8 {
        return Err(AppError::BadRequest(
            "email, username required; password min 8 chars".into(),
        ));
    }

    let exists: Option<(String,)> = sqlx::query_as("SELECT id FROM plat_users WHERE email = ? OR username = ?")
        .bind(&body.email)
        .bind(&body.username)
        .fetch_optional(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;

    if exists.is_some() {
        return Err(AppError::Conflict("email or username already registered".into()));
    }

    let id = Uuid::new_v4().to_string();
    let password_hash =
        bcrypt::hash(&body.password, 10).map_err(|_| AppError::Internal)?;
    let roles = serde_json::to_string(&Vec::<String>::new()).unwrap();
    let permissions = serde_json::to_string(&crate::permissions::default_app_permissions()).unwrap();
    let channel = format!("ch_{}", Uuid::new_v4().simple());
    let verify_token = random_token_bytes();
    let verify_exp = (Utc::now() + Duration::hours(24)).to_rfc3339();
    let now = Utc::now().to_rfc3339();

    sqlx::query(
        r#"INSERT INTO plat_users (
            id, email, username, password_hash, google_id, is_verified, roles, permissions,
            default_channel_id, verification_token, verification_expires_at,
            reset_token, reset_expires_at, created_at, updated_at
        ) VALUES (?, ?, ?, ?, NULL, 0, ?, ?, ?, ?, ?, NULL, NULL, ?, ?)"#,
    )
    .bind(&id)
    .bind(&body.email)
    .bind(&body.username)
    .bind(&password_hash)
    .bind(&roles)
    .bind(&permissions)
    .bind(&channel)
    .bind(&verify_token)
    .bind(&verify_exp)
    .bind(&now)
    .bind(&now)
    .execute(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;

    let verify_url = format!(
        "{}/api/auth/verify-email?token={}",
        state.config.api_public_url.trim_end_matches('/'),
        urlencoding::encode(&verify_token)
    );
    tracing::info!(%verify_url, "User registered; verification URL (dev)");

    Ok(Json(json!({
        "message": "User registered successfully. Please check your email for verification."
    })))
}

pub async fn verify_email(
    State(state): State<AppState>,
    Query(q): Query<VerifyEmailQuery>,
) -> AppResult<Json<serde_json::Value>> {
    let row: Option<UserRow> = sqlx::query_as(
        "SELECT * FROM plat_users WHERE verification_token = ?",
    )
    .bind(&q.token)
    .fetch_optional(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;

    let Some(u) = row else {
        return Err(AppError::BadRequest("invalid or expired token".into()));
    };

    if let Some(exp) = &u.verification_expires_at {
        if let Ok(dt) = chrono::DateTime::parse_from_rfc3339(exp) {
            if Utc::now() > dt.with_timezone(&Utc) {
                return Err(AppError::BadRequest("verification token expired".into()));
            }
        }
    }

    let now = Utc::now().to_rfc3339();
    sqlx::query(
        "UPDATE plat_users SET is_verified = 1, verification_token = NULL, verification_expires_at = NULL, updated_at = ? WHERE id = ?",
    )
    .bind(&now)
    .bind(&u.id)
    .execute(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;

    Ok(Json(json!({ "message": "Email verified successfully." })))
}

pub async fn login(
    State(state): State<AppState>,
    Json(body): Json<LoginBody>,
) -> AppResult<Json<serde_json::Value>> {
    let row: Option<UserRow> = sqlx::query_as("SELECT * FROM plat_users WHERE email = ?")
        .bind(&body.email)
        .fetch_optional(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;

    let Some(u) = row else {
        return Err(AppError::Unauthorized);
    };

    let Some(ref hash) = u.password_hash else {
        return Err(AppError::BadRequest("use Google sign-in for this account".into()));
    };

    if !bcrypt::verify(&body.password, hash).unwrap_or(false) {
        return Err(AppError::Unauthorized);
    }

    if !u.is_verified {
        return Err(AppError::BadRequest("email not verified".into()));
    }

    let roles: Vec<String> = u.roles_vec().map_err(|_| AppError::Internal)?;
    let token = encode_token(
        &state.config,
        &u.id,
        &u.email,
        &u.username,
        &roles,
        &u.default_channel_id,
    )
    .map_err(|_| AppError::Internal)?;

    let user = u.to_public().map_err(|_| AppError::Internal)?;

    Ok(Json(json!({
        "token": token,
        "user": user
    })))
}

pub async fn google_start(State(state): State<AppState>) -> Result<Response, AppError> {
    let Some(client) = state.oauth_client.as_deref() else {
        return Err(AppError::BadRequest(
            "Google OAuth not configured".into(),
        ));
    };

    let (pkce_challenge, pkce_verifier) = PkceCodeChallenge::new_random_sha256();
    let (auth_url, csrf_token) = client
        .authorize_url(CsrfToken::new_random)
        .add_scope(Scope::new("openid".to_string()))
        .add_scope(Scope::new(
            "https://www.googleapis.com/auth/userinfo.email".to_string(),
        ))
        .add_scope(Scope::new(
            "https://www.googleapis.com/auth/userinfo.profile".to_string(),
        ))
        .set_pkce_challenge(pkce_challenge)
        .url();

    {
        let mut pending = state.oauth_pending.write().await;
        pending.insert(
            csrf_token.secret().clone(),
            crate::state::OAuthPending { pkce_verifier },
        );
    }

    let loc = auth_url.to_string();
    Ok(Redirect::temporary(loc.as_str()).into_response())
}

pub async fn google_callback(
    State(state): State<AppState>,
    Query(q): Query<GoogleCallbackQuery>,
) -> Result<Response, AppError> {
    let fe = state.config.frontend_origin.trim_end_matches('/');
    if q.error.is_some() {
        let loc = format!("{fe}#/login?error=oauth");
        return Ok(Redirect::to(loc.as_str()).into_response());
    }

    let Some(code) = q.code else {
        let loc = format!("{fe}#/login?error=missing_code");
        return Ok(Redirect::to(loc.as_str()).into_response());
    };
    let Some(csrf) = q.state else {
        let loc = format!("{fe}#/login?error=missing_state");
        return Ok(Redirect::to(loc.as_str()).into_response());
    };

    let pkce_verifier = {
        let mut pending = state.oauth_pending.write().await;
        pending
            .remove(&csrf)
            .map(|p| p.pkce_verifier)
    };

    let Some(pkce_verifier) = pkce_verifier else {
        let loc = format!("{fe}#/login?error=invalid_state");
        return Ok(Redirect::to(loc.as_str()).into_response());
    };

    let client = state.oauth_client.as_deref().ok_or_else(|| {
        AppError::BadRequest("Google OAuth not configured".into())
    })?;

    let token = client
        .exchange_code(AuthorizationCode::new(code))
        .set_pkce_verifier(pkce_verifier)
        .request_async(oauth2::reqwest::async_http_client)
        .await
        .map_err(|e| {
            tracing::error!(?e, "token exchange failed");
            AppError::BadRequest("oauth token exchange failed".into())
        })?;

    let access = token.access_token().secret();
    let userinfo: serde_json::Value = reqwest::Client::new()
        .get("https://www.googleapis.com/oauth2/v3/userinfo")
        .bearer_auth(access)
        .send()
        .await
        .map_err(|_| AppError::Internal)?
        .json()
        .await
        .map_err(|_| AppError::Internal)?;

    let email = userinfo
        .get("email")
        .and_then(|v| v.as_str())
        .ok_or_else(|| AppError::BadRequest("no email from Google".into()))?
        .to_string();
    let google_id = userinfo
        .get("sub")
        .and_then(|v| v.as_str())
        .ok_or_else(|| AppError::BadRequest("no sub from Google".into()))?
        .to_string();
    let name = userinfo
        .get("name")
        .and_then(|v| v.as_str())
        .unwrap_or("user")
        .to_string();

    let existing: Option<UserRow> = sqlx::query_as("SELECT * FROM plat_users WHERE google_id = ?")
        .bind(&google_id)
        .fetch_optional(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;

    let user_row = if let Some(u) = existing {
        u
    } else {
        let by_email: Option<UserRow> = sqlx::query_as("SELECT * FROM plat_users WHERE email = ?")
            .bind(&email)
            .fetch_optional(&state.pool)
            .await
            .map_err(|_| AppError::Internal)?;

        if let Some(mut u) = by_email {
            let now = Utc::now().to_rfc3339();
            sqlx::query("UPDATE plat_users SET google_id = ?, is_verified = 1, updated_at = ? WHERE id = ?")
                .bind(&google_id)
                .bind(&now)
                .bind(&u.id)
                .execute(&state.pool)
                .await
                .map_err(|_| AppError::Internal)?;
            u.google_id = Some(google_id.clone());
            u.is_verified = true;
            u
        } else {
            let id = Uuid::new_v4().to_string();
            let roles = serde_json::to_string(&Vec::<String>::new()).unwrap();
            let permissions =
                serde_json::to_string(&crate::permissions::default_app_permissions()).unwrap();
            let channel = format!("ch_{}", Uuid::new_v4().simple());
            let now = Utc::now().to_rfc3339();
            sqlx::query(
                r#"INSERT INTO plat_users (
                    id, email, username, password_hash, google_id, is_verified, roles, permissions,
                    default_channel_id, verification_token, verification_expires_at,
                    reset_token, reset_expires_at, created_at, updated_at
                ) VALUES (?, ?, ?, NULL, ?, 1, ?, ?, ?, NULL, NULL, NULL, NULL, ?, ?)"#,
            )
            .bind(&id)
            .bind(&email)
            .bind(&name)
            .bind(&google_id)
            .bind(&roles)
            .bind(&permissions)
            .bind(&channel)
            .bind(&now)
            .bind(&now)
            .execute(&state.pool)
            .await
            .map_err(|_| AppError::Internal)?;

            sqlx::query_as("SELECT * FROM plat_users WHERE id = ?")
                .bind(&id)
                .fetch_one(&state.pool)
                .await
                .map_err(|_| AppError::Internal)?
        }
    };

    let roles: Vec<String> = user_row.roles_vec().map_err(|_| AppError::Internal)?;
    let jwt = encode_token(
        &state.config,
        &user_row.id,
        &user_row.email,
        &user_row.username,
        &roles,
        &user_row.default_channel_id,
    )
    .map_err(|_| AppError::Internal)?;

    let redirect = format!("{fe}#/auth/callback?token={}", urlencoding::encode(&jwt));
    Ok(Redirect::temporary(redirect.as_str()).into_response())
}

pub async fn forgot_password(
    State(state): State<AppState>,
    Json(body): Json<ForgotPasswordBody>,
) -> AppResult<Json<serde_json::Value>> {
    let row: Option<(String, String)> =
        sqlx::query_as("SELECT id, email FROM plat_users WHERE email = ?")
            .bind(&body.email)
            .fetch_optional(&state.pool)
            .await
            .map_err(|_| AppError::Internal)?;

    if let Some((id, email)) = row {
        let reset_token = random_token_bytes();
        let exp = (Utc::now() + Duration::hours(1)).to_rfc3339();
        let now = Utc::now().to_rfc3339();
        sqlx::query(
            "UPDATE plat_users SET reset_token = ?, reset_expires_at = ?, updated_at = ? WHERE id = ?",
        )
        .bind(&reset_token)
        .bind(&exp)
        .bind(&now)
        .bind(&id)
        .execute(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;

        let fe = state.config.frontend_origin.trim_end_matches('/');
        let reset_url = format!("{fe}#/reset-password?token={}", urlencoding::encode(&reset_token));
        tracing::info!(%email, %reset_url, "Password reset requested");
    }

    Ok(Json(json!({
        "message": "Password reset email sent."
    })))
}

pub async fn reset_password(
    State(state): State<AppState>,
    Json(body): Json<ResetPasswordBody>,
) -> AppResult<Json<serde_json::Value>> {
    if body.password.len() < 8 {
        return Err(AppError::BadRequest("password min 8 chars".into()));
    }

    let row: Option<UserRow> = sqlx::query_as("SELECT * FROM plat_users WHERE reset_token = ?")
        .bind(&body.token)
        .fetch_optional(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;

    let Some(u) = row else {
        return Err(AppError::BadRequest("invalid or expired token".into()));
    };

    if let Some(exp) = &u.reset_expires_at {
        if let Ok(dt) = chrono::DateTime::parse_from_rfc3339(exp) {
            if Utc::now() > dt.with_timezone(&Utc) {
                return Err(AppError::BadRequest("reset token expired".into()));
            }
        }
    }

    let hash = bcrypt::hash(&body.password, 10).map_err(|_| AppError::Internal)?;
    let now = Utc::now().to_rfc3339();
    sqlx::query(
        "UPDATE plat_users SET password_hash = ?, reset_token = NULL, reset_expires_at = NULL, updated_at = ? WHERE id = ?",
    )
    .bind(&hash)
    .bind(&now)
    .bind(&u.id)
    .execute(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;

    Ok(Json(json!({ "message": "Password reset successfully." })))
}

pub async fn get_roles(
    JwtClaims(claims): JwtClaims,
) -> AppResult<Json<serde_json::Value>> {
    Ok(Json(json!({ "roles": claims.roles })))
}

pub async fn get_permissions(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<serde_json::Value>> {
    let perms = resolve_permission_names(&state.pool, &claims.sub, &claims.roles).await;
    Ok(Json(json!({ "permissions": perms })))
}

pub async fn get_user(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<serde_json::Value>> {
    let row: UserRow = sqlx::query_as("SELECT * FROM plat_users WHERE id = ?")
        .bind(&claims.sub)
        .fetch_optional(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?
        .ok_or(AppError::NotFound)?;

    let user = row.to_public().map_err(|_| AppError::Internal)?;
    Ok(Json(json!({ "user": user })))
}
