use axum::extract::State;
use axum::Json;
use serde_json::json;

use crate::error::{AppError, AppResult};
use crate::extractors::JwtClaims;
use crate::permissions::{require_any_permission, PERM_MORPH_UTIL};
use crate::permissions_resolve::resolve_permission_names;
use crate::state::AppState;

async fn require_morph_util(state: &AppState, claims: &crate::jwt::Claims) -> AppResult<()> {
    let perms = resolve_permission_names(&state.pool, &claims.sub, &claims.roles).await;
    let stored: Vec<String> = if crate::permissions::is_admin(&claims.roles) {
        vec![]
    } else {
        perms.clone()
    };
    if require_any_permission(&claims.roles, &stored, &[PERM_MORPH_UTIL]) {
        Ok(())
    } else {
        Err(AppError::Forbidden)
    }
}

pub async fn main_panel(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<serde_json::Value>> {
    require_morph_util(&state, &claims).await?;
    Ok(Json(json!({
        "message": "Morph Util access granted",
        "user": claims.username
    })))
}

pub async fn forms_create(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<serde_json::Value>> {
    require_morph_util(&state, &claims).await?;
    Ok(Json(json!({ "message": "Forms endpoint stub — create form" })))
}

pub async fn email_compose(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<serde_json::Value>> {
    require_morph_util(&state, &claims).await?;
    Ok(Json(json!({ "message": "Email compose stub" })))
}

pub async fn reports(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<serde_json::Value>> {
    require_morph_util(&state, &claims).await?;
    Ok(Json(json!({ "message": "Reports stub — PDF/Excel" })))
}
