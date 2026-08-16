use axum::{
    body::Body,
    extract::State,
    http::{HeaderValue, Method, Request, StatusCode, header},
    middleware::Next,
    response::Response,
};

use crate::services::AppState;

/// Handle CORS, including OPTIONS preflight. Echoes `Origin` when it is in `settings.cors.allowed_origins`
/// (required when the browser enforces a specific allowlist instead of `*`).
pub async fn cors_middleware(
    State(state): State<AppState>,
    request: Request<Body>,
    next: Next,
) -> Response {
    let allowed = &state.settings.cors.allowed_origins;
    let origin_header = request
        .headers()
        .get(header::ORIGIN)
        .and_then(|v| v.to_str().ok())
        .map(|s| s.to_string());

    let allow_origin = origin_header.as_ref().and_then(|o| {
        allowed
            .iter()
            .find(|a| a.as_str() == o.as_str())
            .map(|_| o.as_str())
    });

    let req_headers = request
        .headers()
        .get(header::ACCESS_CONTROL_REQUEST_HEADERS)
        .and_then(|v| v.to_str().ok())
        .unwrap_or("");

    if request.method() == Method::OPTIONS {
        let mut res = Response::builder().status(StatusCode::OK);

        if let Some(origin) = allow_origin {
            if let Ok(val) = HeaderValue::from_str(origin) {
                res = res.header(header::ACCESS_CONTROL_ALLOW_ORIGIN, val);
            }
        }

        res = res
            .header(
                header::ACCESS_CONTROL_ALLOW_METHODS,
                "GET, POST, PUT, PATCH, DELETE, OPTIONS",
            )
            .header(header::ACCESS_CONTROL_MAX_AGE, "86400");

        let allow_headers = if req_headers.is_empty() {
            "*"
        } else {
            req_headers
        };
        if let Ok(val) = HeaderValue::from_str(allow_headers) {
            res = res.header(header::ACCESS_CONTROL_ALLOW_HEADERS, val);
        } else {
            res = res.header(header::ACCESS_CONTROL_ALLOW_HEADERS, "*");
        }

        return res.body(Body::empty()).unwrap();
    }

    let mut response = next.run(request).await;

    if let Some(origin) = allow_origin {
        if let Ok(val) = HeaderValue::from_str(origin) {
            response
                .headers_mut()
                .insert(header::ACCESS_CONTROL_ALLOW_ORIGIN, val);
        }
    }

    let h = response.headers_mut();
    h.insert(
        header::ACCESS_CONTROL_ALLOW_METHODS,
        HeaderValue::from_static("GET, POST, PUT, PATCH, DELETE, OPTIONS"),
    );
    h.insert(
        header::ACCESS_CONTROL_ALLOW_HEADERS,
        HeaderValue::from_static("*"),
    );

    response
}
