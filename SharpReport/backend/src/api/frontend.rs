use std::path::{Path, PathBuf};

use axum::{
    body::Body,
    http::{StatusCode, Uri, header},
    response::{IntoResponse, Redirect, Response},
};

/// Directory of the production UI (`index.html` plus assets). Unset in local debug.
pub fn ui_dir() -> Option<PathBuf> {
    let raw = std::env::var("SHARPREPORT_UI_DIR").ok()?;
    let dir = PathBuf::from(raw.trim());
    if dir.join("index.html").is_file() {
        Some(dir)
    } else {
        None
    }
}

/// File to send for a UI path. Navigation routes use `index.html`. A missing
/// asset (a segment with a `.`) is a miss so it is not served as HTML.
pub fn resolve_ui_path(dir: &Path, request_path: &str) -> Option<PathBuf> {
    let rel = request_path.trim_start_matches('/');
    if rel.is_empty() {
        return Some(dir.join("index.html"));
    }
    if rel.split('/').any(|part| part.is_empty() || part == "." || part == "..") {
        return None;
    }
    let full = dir.join(rel);
    if full.is_file() {
        return Some(full);
    }
    if Path::new(rel)
        .file_name()
        .and_then(|name| name.to_str())
        .is_some_and(|name| name.contains('.'))
    {
        return None;
    }
    Some(dir.join("index.html"))
}

fn content_type(path: &Path) -> &'static str {
    match path.extension().and_then(|ext| ext.to_str()) {
        Some("html") => "text/html; charset=utf-8",
        Some("js") => "text/javascript; charset=utf-8",
        Some("css") => "text/css; charset=utf-8",
        Some("json") | Some("map") => "application/json",
        Some("svg") => "image/svg+xml",
        Some("png") => "image/png",
        Some("webp") => "image/webp",
        Some("woff2") => "font/woff2",
        _ => "application/octet-stream",
    }
}

fn file_response(path: &Path) -> Response {
    match std::fs::read(path) {
        Ok(bytes) => Response::builder()
            .status(StatusCode::OK)
            .header(header::CONTENT_TYPE, content_type(path))
            .body(Body::from(bytes))
            .unwrap(),
        Err(_) => StatusCode::NOT_FOUND.into_response(),
    }
}

/// `GET /` only. In a debug build, the UI runs on Vite; redirect the browser so opening the API
/// port (e.g. :3050) is not a dead end. `DATAPULSE_VITE_DEV_URL` overrides the default below.
const DEFAULT_VITE_DEV_URL: &str = "http://localhost:5178";

pub async fn root() -> Response {
    if ui_dir().is_none() && cfg!(debug_assertions) {
        let url = std::env::var("DATAPULSE_VITE_DEV_URL")
            .ok()
            .filter(|s| !s.is_empty())
            .unwrap_or_else(|| DEFAULT_VITE_DEV_URL.to_string());
        return Redirect::temporary(&url).into_response();
    }
    serve_path("/")
}

pub async fn serve(uri: Uri) -> Response {
    let path = uri.path();
    if path.starts_with("/api")
        || path.starts_with("/public")
        || path.starts_with("/metabase")
    {
        return StatusCode::NOT_FOUND.into_response();
    }
    serve_path(path)
}

fn serve_path(path: &str) -> Response {
    if let Some(dir) = ui_dir() {
        if let Some(file) = resolve_ui_path(&dir, path) {
            return file_response(&file);
        }
        return StatusCode::NOT_FOUND.into_response();
    }
    stub_html()
}

fn stub_html() -> Response {
    // SPA fallback when no route matches. In production, replace with the built Svelte `index.html` + assets.
    // Do not use `import.meta` here: it only works in JS modules, and `import.meta.env` is a Vite compile-time
    // transform — it does not exist in a raw script served from the API.
    Response::builder()
        .status(StatusCode::OK)
        .header("Content-Type", "text/html")
        .body(Body::from(
            r#"<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8" />
    <title>DataPulse</title>
</head>
<body>
    <div id="app">DataPulse API is running.</div>
    <p style="font-family: system-ui, sans-serif; max-width: 32rem; line-height: 1.4;">
      In local development, open the web app on the <strong>Vite</strong> port (e.g.
      <a href="http://127.0.0.1:5178">http://127.0.0.1:5178</a> — see <code>frontend/vite.config.ts</code>).
    </p>
</body>
</html>"#
        ))
        .unwrap()
}

#[cfg(test)]
mod tests {
    use super::resolve_ui_path;
    use std::fs;

    #[test]
    fn navigation_uses_index_and_missing_assets_do_not() {
        let dir = tempfile::tempdir().unwrap();
        fs::write(dir.path().join("index.html"), "<!doctype html>").unwrap();
        fs::create_dir(dir.path().join("_app")).unwrap();
        fs::write(dir.path().join("_app/start.js"), "console.log(1)").unwrap();

        let index = dir.path().join("index.html");
        assert_eq!(resolve_ui_path(dir.path(), "/"), Some(index.clone()));
        assert_eq!(resolve_ui_path(dir.path(), "/data-tables"), Some(index));
        assert_eq!(
            resolve_ui_path(dir.path(), "/_app/start.js"),
            Some(dir.path().join("_app/start.js"))
        );
        assert_eq!(resolve_ui_path(dir.path(), "/_app/missing.js"), None);
        assert_eq!(resolve_ui_path(dir.path(), "/../secret"), None);
    }
}
