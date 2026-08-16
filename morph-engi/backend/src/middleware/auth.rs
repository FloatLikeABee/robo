use axum::{
    body::Body,
    extract::State,
    http::{Request, StatusCode},
    middleware::Next,
    response::{IntoResponse, Response},
};
use std::sync::Arc;

use crate::services::AppState;

#[derive(Clone, Debug)]
pub struct AuthContext {
    pub user_id: i64,
    pub org_id: i64,
    pub role: String,
    #[allow(dead_code)]
    pub bearer: String,
}

fn bearer_token(header: &str) -> Option<String> {
    let h = header.trim();
    let mut parts = h.splitn(2, ' ');
    let scheme = parts.next()?;
    if !scheme.eq_ignore_ascii_case("bearer") {
        return None;
    }
    let token = parts.next()?.trim();
    if token.is_empty() {
        return None;
    }
    Some(token.to_string())
}

fn is_public(path: &str, method: &str) -> bool {
    if path == "/auth/login" && method == "POST" {
        return true;
    }
    if path == "/auth/platform-session" && method == "POST" {
        return true;
    }
    if path == "/auth/dev-login" && method == "POST" {
        return true;
    }
    false
}

pub async fn auth_layer(
    State(state): State<Arc<AppState>>,
    mut request: Request<Body>,
    next: Next,
) -> Response {
    let path = request.uri().path().to_string();
    let method = request.method().as_str().to_string();

    // Nested under /api/v1 — paths here are e.g. /dashboard, not /api/v1/dashboard.
    if is_public(&path, &method) {
        return next.run(request).await;
    }

    let token = request
        .headers()
        .get(axum::http::header::AUTHORIZATION)
        .and_then(|v| v.to_str().ok())
        .and_then(bearer_token);

    let Some(token) = token else {
        return (StatusCode::UNAUTHORIZED, axum::Json(serde_json::json!({"error":"missing token"})))
            .into_response();
    };

    match state.jwt.verify(&token) {
        Ok(claims) => {
            request.extensions_mut().insert(AuthContext {
                user_id: claims.uid,
                org_id: claims.org,
                role: claims.role,
                bearer: token,
            });
            next.run(request).await
        }
        Err(_) => (
            StatusCode::UNAUTHORIZED,
            axum::Json(serde_json::json!({"error":"invalid token"})),
        )
            .into_response(),
    }
}