pub mod doc_text;
pub mod identity;
pub mod jwt;
pub mod users_panel;

use crate::config::Settings;
use crate::db::DbPool;
use morphai::{Client, Config};
use std::sync::Arc;

#[derive(Clone)]
pub struct AppState {
    pub pool: DbPool,
    pub settings: Settings,
    pub users_panel: users_panel::UsersPanelClient,
    pub jwt: jwt::JwtService,
    pub ai: Arc<Option<Client>>,
}

impl AppState {
    pub fn new(pool: DbPool, settings: Settings) -> Self {
        let morph_cfg = Config::from_env();
        let ai = if morph_cfg.configured() {
            Some(Client::new(morph_cfg))
        } else {
            None
        };
        Self {
            pool,
            jwt: jwt::JwtService::new(
                settings.jwt_secret.clone(),
                settings.jwt_access_expiry_min,
            ),
            users_panel: users_panel::UsersPanelClient::new(settings.users_panel_base_url.clone()),
            settings,
            ai: Arc::new(ai),
        }
    }
}