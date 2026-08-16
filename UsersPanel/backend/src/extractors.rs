use async_trait::async_trait;
use axum::extract::FromRequestParts;
use axum::http::header::AUTHORIZATION;
use axum::http::request::Parts;

use crate::error::AppError;
use crate::jwt::{decode_token, Claims};
use crate::state::AppState;

pub struct JwtClaims(pub Claims);

#[async_trait]
impl FromRequestParts<AppState> for JwtClaims {
    type Rejection = AppError;

    async fn from_request_parts(
        parts: &mut Parts,
        state: &AppState,
    ) -> Result<Self, Self::Rejection> {
        let auth = parts
            .headers
            .get(AUTHORIZATION)
            .and_then(|v| v.to_str().ok())
            .ok_or(AppError::Unauthorized)?;

        const PREFIX: &str = "Bearer ";
        if !auth.starts_with(PREFIX) {
            return Err(AppError::Unauthorized);
        }
        let token = &auth[PREFIX.len()..];
        let claims = decode_token(&state.config, token).map_err(|_| AppError::Unauthorized)?;
        Ok(JwtClaims(claims))
    }
}
