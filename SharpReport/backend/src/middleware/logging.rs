use axum::{body::Body, http::Request, middleware::Next, response::Response};
use std::time::Instant;
use tracing::info;
use uuid::Uuid;

pub async fn logging_middleware(request: Request<Body>, next: Next) -> Response {
    let request_id = Uuid::new_v4();
    let method = request.method().clone();
    let uri = request.uri().clone();
    let start = Instant::now();

    let response = next.run(request).await;
    let status = response.status();
    let latency = start.elapsed().as_millis();

    info!(
        request_id = %request_id,
        method = %method,
        uri = %uri,
        status = %status,
        latency_ms = %latency,
        "HTTP request"
    );

    response
}
