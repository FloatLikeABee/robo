use axum::{
    body::Body,
    http::StatusCode,
    response::{IntoResponse, Redirect, Response},
};

/// `GET /` only. In a debug build, the UI runs on Vite; redirect the browser so opening the API
/// port (e.g. :3050) is not a dead end. `DATAPULSE_VITE_DEV_URL` overrides the default below.
const DEFAULT_VITE_DEV_URL: &str = "http://localhost:5178";

pub async fn root() -> impl IntoResponse {
    if cfg!(debug_assertions) {
        let url = std::env::var("DATAPULSE_VITE_DEV_URL")
            .ok()
            .filter(|s| !s.is_empty())
            .unwrap_or_else(|| DEFAULT_VITE_DEV_URL.to_string());
        return Redirect::temporary(&url).into_response();
    }
    serve().await
}

pub async fn serve() -> Response {
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
