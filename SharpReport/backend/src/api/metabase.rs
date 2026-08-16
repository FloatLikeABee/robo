use crate::services::AppState;
use axum::{
    body::Body,
    extract::{Path, State},
    http::{Request, StatusCode},
    response::Response,
};
use reqwest::Client;
use std::convert::Infallible;

pub async fn proxy(
    State(state): State<AppState>,
    Path(path): Path<String>,
    request: Request<Body>,
) -> Result<Response, Response> {
    // Forward to Metabase: path is the segment after `/metabase/` (e.g. `api/health`, `dashboard/1`, `embed/dashboard/TOKEN`).
    let metabase_url = format!(
        "http://{}:{}/{}",
        state.metabase_handle.get_host(),
        state.metabase_handle.get_port(),
        path.trim_start_matches('/')
    );

    // Forward the request to Metabase
    let client = Client::new();

    // Convert Axum request to reqwest request
    let method = request.method().clone();
    let headers = request.headers().clone();
    let body_bytes = axum::body::to_bytes(request.into_body(), usize::MAX)
        .await
        .map_err(|e| {
            Response::builder()
                .status(StatusCode::BAD_REQUEST)
                .body(Body::from(format!("Failed to read request body: {}", e)))
                .unwrap()
        })?;

    let mut reqwest_request = client.request(method, &metabase_url);

    // Copy headers
    for (name, value) in headers.iter() {
        reqwest_request = reqwest_request.header(name, value);
    }

    // Set body
    let reqwest_request = reqwest_request.body(body_bytes.to_vec());

    // Send request to Metabase
    let response = reqwest_request.send().await.map_err(|e| {
        Response::builder()
            .status(StatusCode::BAD_GATEWAY)
            .body(Body::from(format!("Metabase request failed: {}", e)))
            .unwrap()
    })?;

    // Convert reqwest response to Axum response
    let status = response.status();
    let response_headers = response.headers().clone();
    let response_body = response.bytes().await.map_err(|e| {
        Response::builder()
            .status(StatusCode::BAD_GATEWAY)
            .body(Body::from(format!("Failed to read response body: {}", e)))
            .unwrap()
    })?;

    let mut axum_response = Response::builder().status(status);

    // Copy headers; drop frame-blocking so Metabase can render inside app iframes (modal + builder tab).
    for (name, value) in response_headers.iter() {
        if name.as_str().eq_ignore_ascii_case("x-frame-options") {
            continue;
        }
        axum_response = axum_response.header(name, value);
    }

    Ok(axum_response.body(Body::from(response_body)).unwrap())
}
