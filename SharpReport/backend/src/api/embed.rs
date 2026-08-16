use crate::services::AppState;
use axum::{
    Json,
    extract::{Path, State},
    http::StatusCode,
};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct EmbedRequest {
    pub params: Option<serde_json::Value>,
}

pub async fn dashboard(
    State(_state): State<AppState>,
    Path(id): Path<String>,
    Json(_request): Json<EmbedRequest>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    Ok(Json(serde_json::json!({
        "embed_url": format!("/embed/dashboard/{}", id),
        "token": "placeholder-token",
        "expires_in": 600
    })))
}

pub async fn card(
    State(_state): State<AppState>,
    Path(id): Path<String>,
    Json(_request): Json<EmbedRequest>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    Ok(Json(serde_json::json!({
        "embed_url": format!("/embed/card/{}", id),
        "token": "placeholder-token",
        "expires_in": 600
    })))
}
