use axum::{
    body::Body,
    extract::State,
    http::{Request, StatusCode},
    middleware::Next,
    response::{IntoResponse, Response},
};

use crate::services::AppState;
use crate::services::users_panel::UsersPanelError;

fn bearer_token(header: &str) -> Option<String> {
    let h = header.trim();
    if h.is_empty() {
        return None;
    }
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

fn is_public_api_path(path: &str, method: &str) -> bool {
    if path == "/api/v1/health" || path == "/api/v1/ai/status" {
        return true;
    }
    if path.starts_with("/api/v1/setup") {
        return true;
    }
    if path == "/api/v1/auth/login" && method == "POST" {
        return true;
    }
    false
}

pub async fn auth_middleware(
    State(state): State<AppState>,
    request: Request<Body>,
    next: Next,
) -> Response {
    let method = request.method().as_str().to_string();
    let path = request.uri().path().to_string();

    if !path.starts_with("/api/v1/") || is_public_api_path(&path, &method) {
        return next.run(request).await;
    }

    if path == "/api/v1/auth/me" && method == "GET" {
        return next.run(request).await;
    }

    if path == "/api/v1/auth/logout" && method == "POST" {
        return next.run(request).await;
    }

    let token = request
        .headers()
        .get(axum::http::header::AUTHORIZATION)
        .and_then(|v| v.to_str().ok())
        .and_then(bearer_token);

    let Some(token) = token else {
        return (
            StatusCode::UNAUTHORIZED,
            "Missing Authorization bearer token",
        )
            .into_response();
    };

    match state.users_panel.session(&token).await {
        Ok(session) => {
            let mut req = request;
            if let Ok(v) = session.user.email.parse() {
                req.headers_mut().insert("x-user-email", v);
            }
            if let Ok(v) = crate::services::users_panel::resolve_role(&session.user.roles).parse() {
                req.headers_mut().insert("x-user-role", v);
            }
            if let Ok(v) = session.user.roles.join(",").parse() {
                req.headers_mut().insert("x-user-roles", v);
            }
            if let Ok(v) = session.permissions.join(",").parse() {
                req.headers_mut().insert("x-user-permissions", v);
            }
            next.run(req).await
        }
        Err(UsersPanelError::InvalidCredentials) => {
            (StatusCode::UNAUTHORIZED, "Invalid or expired token").into_response()
        }
        Err(UsersPanelError::Forbidden) => (
            StatusCode::FORBIDDEN,
            "Sharp Reports access is restricted by admin policy",
        )
            .into_response(),
        Err(_) => (StatusCode::BAD_GATEWAY, "UsersPanel unavailable").into_response(),
    }
}
