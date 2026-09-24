use axum::{Json, Router, routing::get};
use serde_json::{Value, json};

/// Local liveness. Does not call Morph or open a new database connection.
pub async fn health() -> Json<Value> {
    Json(json!({"status": "ok"}))
}

#[cfg(test)]
pub fn probe_router() -> Router {
    Router::new()
        .route("/health", get(health))
        .route("/ready", get(health))
}

#[cfg(test)]
mod tests {
    use super::probe_router;
    use axum::body::{Body, to_bytes};
    use axum::http::{Request, StatusCode};
    use tower::ServiceExt;

    #[tokio::test]
    async fn health_and_ready_are_ok_without_morph() {
        for path in ["/health", "/ready"] {
            let response = probe_router()
                .oneshot(
                    Request::builder()
                        .uri(path)
                        .body(Body::empty())
                        .unwrap(),
                )
                .await
                .unwrap();
            assert_eq!(response.status(), StatusCode::OK);
            let bytes = to_bytes(response.into_body(), 1024).await.unwrap();
            assert_eq!(bytes.as_ref(), br#"{"status":"ok"}"#);
        }
    }
}
