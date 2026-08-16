//! MorphAI configuration probe (does not expose secrets).

use axum::Json;
use morphai::Config;
use serde_json::json;

use crate::config::load_env_files;

pub async fn status() -> Json<serde_json::Value> {
    load_env_files();
    let cfg = Config::from_env();
    Json(json!({
        "configured": cfg.configured(),
        "model": cfg.model,
        "base_url": cfg.base_url,
        "uses_openai_compatible": cfg.uses_openai_compatible(),
        "chat_url": if cfg.uses_openai_compatible() {
            cfg.chat_completions_url()
        } else {
            cfg.api_url.clone()
        },
        "env_hint": if cfg.configured() {
            "MorphAI env vars are loaded in this process."
        } else {
            "MORPH_AI_API_KEY missing — set it in SharpReport/.env and restart sharpreport-api."
        }
    }))
}
