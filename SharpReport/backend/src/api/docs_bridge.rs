//! Reverse proxy to Academi Docs APIs (`ACADEMI_API_BASE_URL` / `/api/v1/...`).

use axum::{
    extract::{Path, Request, State},
    http::{HeaderMap, HeaderName, HeaderValue, StatusCode, header},
    response::{IntoResponse, Response},
};
use tracing::warn;

use crate::services::AppState;

const HOP_BY_HOP: &[&str] = &[
    "connection",
    "keep-alive",
    "proxy-authenticate",
    "proxy-authorization",
    "te",
    "trailers",
    "transfer-encoding",
    "upgrade",
    "host",
    "content-length",
];

fn academi_base(state: &AppState) -> String {
    state
        .settings
        .academi
        .base_url
        .trim_end_matches('/')
        .to_string()
}

/// Content document extensions allowed on Docs upload paths.
pub fn is_docs_content_filename(name: &str) -> bool {
    let lower = name.to_ascii_lowercase();
    lower.ends_with(".pdf")
        || lower.ends_with(".txt")
        || lower.ends_with(".md")
        || lower.ends_with(".markdown")
        || lower.ends_with(".png")
        || lower.ends_with(".jpg")
        || lower.ends_with(".jpeg")
        || lower.ends_with(".webp")
        || lower.ends_with(".gif")
        || lower.ends_with(".heic")
        || lower.ends_with(".heif")
}

/// Tabular data extensions for DataX import (not Docs content).
pub fn is_data_table_filename(name: &str) -> bool {
    let lower = name.to_ascii_lowercase();
    lower.ends_with(".csv")
        || lower.ends_with(".tsv")
        || lower.ends_with(".json")
        || lower.ends_with(".xlsx")
        || lower.ends_with(".xls")
}

fn filename_from_multipart(body: &[u8]) -> Option<String> {
    let text = String::from_utf8_lossy(body);
    for line in text.lines().take(40) {
        if let Some(idx) = line.to_ascii_lowercase().find("filename=") {
            let rest = &line[idx + "filename=".len()..];
            let rest = rest.trim();
            let name = rest.trim_matches('"').trim_matches('\'').trim();
            if !name.is_empty() {
                return Some(name.to_string());
            }
        }
    }
    None
}

/// Proxy that preserves query string from the original request URI.
pub async fn proxy_with_query(
    State(state): State<AppState>,
    Path(path): Path<String>,
    req: Request,
) -> Response {
    let method = req.method().clone();
    let headers = req.headers().clone();
    let query = req.uri().query().unwrap_or("").to_string();
    let body = match axum::body::to_bytes(req.into_body(), 32 * 1024 * 1024).await {
        Ok(b) => b,
        Err(e) => {
            return (StatusCode::BAD_REQUEST, format!("Invalid body: {e}")).into_response();
        }
    };

    let base = academi_base(&state);
    if base.is_empty() {
        return (
            StatusCode::SERVICE_UNAVAILABLE,
            "ACADEMI_API_BASE_URL is not configured",
        )
            .into_response();
    }

    let path = path.trim_start_matches('/');
    if path == "docs/upload" || path.ends_with("/docs/upload") {
        if let Some(name) = filename_from_multipart(&body) {
            if is_data_table_filename(&name) && !name.to_ascii_lowercase().ends_with(".txt") {
                return (
                    StatusCode::BAD_REQUEST,
                    "Docs accepts content files (PDF, TXT, Markdown, images). CSV/JSON belong in Data tables import.",
                )
                    .into_response();
            }
            if !is_docs_content_filename(&name) {
                return (
                    StatusCode::BAD_REQUEST,
                    "Unsupported Docs content type. Use PDF or TXT (Markdown/images also allowed).",
                )
                    .into_response();
            }
        }
    }

    let url = if query.is_empty() {
        format!("{base}/api/v1/{path}")
    } else {
        format!("{base}/api/v1/{path}?{query}")
    };

    let client = reqwest::Client::new();
    let mut rb = client.request(method, &url);
    for (name, value) in headers.iter() {
        let key = name.as_str();
        if HOP_BY_HOP.iter().any(|h| key.eq_ignore_ascii_case(h)) {
            continue;
        }
        if let Ok(v) = HeaderValue::from_bytes(value.as_bytes()) {
            rb = rb.header(name.clone(), v);
        }
    }
    if !body.is_empty() {
        rb = rb.body(body.to_vec());
    }

    match rb.send().await {
        Ok(upstream) => {
            let status =
                StatusCode::from_u16(upstream.status().as_u16()).unwrap_or(StatusCode::BAD_GATEWAY);
            let mut response_headers = HeaderMap::new();
            for (name, value) in upstream.headers().iter() {
                let key = name.as_str();
                if HOP_BY_HOP.iter().any(|h| key.eq_ignore_ascii_case(h)) {
                    continue;
                }
                if let (Ok(n), Ok(v)) = (
                    HeaderName::from_bytes(name.as_str().as_bytes()),
                    HeaderValue::from_bytes(value.as_bytes()),
                ) {
                    response_headers.insert(n, v);
                }
            }
            match upstream.bytes().await {
                Ok(bytes) => {
                    let mut res = Response::new(bytes.into());
                    *res.status_mut() = status;
                    *res.headers_mut() = response_headers;
                    if !res.headers().contains_key(header::CONTENT_TYPE) {
                        res.headers_mut().insert(
                            header::CONTENT_TYPE,
                            HeaderValue::from_static("application/octet-stream"),
                        );
                    }
                    res
                }
                Err(e) => {
                    warn!("docs bridge body error: {e}");
                    (StatusCode::BAD_GATEWAY, format!("Academi read error: {e}")).into_response()
                }
            }
        }
        Err(e) => {
            warn!("docs bridge proxy error: {e}");
            (
                StatusCode::BAD_GATEWAY,
                format!("Cannot reach Academi Docs API at {base}: {e}"),
            )
                .into_response()
        }
    }
}
