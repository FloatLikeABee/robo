use axum::{Json, extract::State};
use serde::Serialize;

use crate::services::AppState;

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct MetabaseInfo {
    pub port: u16,
    pub health_check_interval: u64,
    pub restart_on_failure: bool,
    pub max_restarts: u32,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SystemInfo {
    pub version: &'static str,
    pub environment: String,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SettingsResponse {
    pub metabase: MetabaseInfo,
    pub system: SystemInfo,
}

pub async fn get(State(state): State<AppState>) -> Json<SettingsResponse> {
    let m = &state.settings.metabase;
    let environment = std::env::var("RUN_ENV").unwrap_or_else(|_| "development".into());
    Json(SettingsResponse {
        metabase: MetabaseInfo {
            port: m.port,
            health_check_interval: m.health_check_interval,
            restart_on_failure: m.restart_on_failure,
            max_restarts: m.max_restarts,
        },
        system: SystemInfo {
            version: env!("CARGO_PKG_VERSION"),
            environment,
        },
    })
}
