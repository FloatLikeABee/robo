use crate::services::AppState;
use axum::{
    Json,
    extract::{Path, State},
    http::StatusCode,
};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct Dashboard {
    pub id: String,
    pub name: String,
    pub description: Option<String>,
    pub metabase_id: Option<u64>,
    pub database_id: String,
    pub is_public: bool,
}

pub async fn list(
    State(_state): State<AppState>,
) -> Result<Json<Vec<Dashboard>>, (StatusCode, String)> {
    Ok(Json(vec![Dashboard {
        id: "1".to_string(),
        name: "Sales Overview".to_string(),
        description: Some("Monthly sales performance".to_string()),
        metabase_id: Some(1),
        database_id: "1".to_string(),
        is_public: false,
    }]))
}

pub async fn create(
    State(_state): State<AppState>,
    Json(_request): Json<Dashboard>,
) -> Result<Json<Dashboard>, (StatusCode, String)> {
    Ok(Json(Dashboard {
        id: "new-id".to_string(),
        name: "New Dashboard".to_string(),
        description: None,
        metabase_id: None,
        database_id: "1".to_string(),
        is_public: false,
    }))
}

pub async fn get(
    State(_state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<Dashboard>, (StatusCode, String)> {
    Ok(Json(Dashboard {
        id,
        name: "Dashboard".to_string(),
        description: None,
        metabase_id: Some(1),
        database_id: "1".to_string(),
        is_public: false,
    }))
}

pub async fn get_embed_url(
    State(_state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    Ok(Json(serde_json::json!({
        "embed_url": format!("/embed/dashboard/{}", id),
        "token": "placeholder-token"
    })))
}
