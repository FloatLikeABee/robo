pub mod ai_status;
pub mod auth;
pub mod dashboards;
pub mod data_page_publish;
pub mod data_table_query;
pub mod data_tables;
pub mod databases;
pub mod docs_bridge;
pub mod docs_publish;
pub mod embed;
pub mod frontend;
pub mod mcp_tools;
pub mod metabase;
pub mod queries;
pub mod report_assistant;
pub mod report_assistant_llm;
pub mod report_builder;
pub mod schema;
pub mod settings;
pub mod setup;

use axum::{Json, http::StatusCode, response::IntoResponse};
use serde_json::json;

pub fn error_response(status: StatusCode, message: &str) -> impl IntoResponse {
    (status, Json(json!({"error": message})))
}
