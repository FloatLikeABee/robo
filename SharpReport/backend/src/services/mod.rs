pub mod auth_service;
pub mod database_service;
pub mod report_service;
pub mod users_panel;

use crate::config::Settings;
use crate::db::Database;
use crate::metabase::orchestrator::Orchestrator;
use crate::services::users_panel::UsersPanelClient;
use std::sync::Arc;

#[derive(Clone)]
pub struct AppState {
    pub db_pool: Database,
    pub metabase_handle: Arc<Orchestrator>,
    pub settings: Settings,
    pub users_panel: UsersPanelClient,
}

impl AppState {
    pub fn new(db_pool: Database, metabase_handle: Arc<Orchestrator>, settings: Settings) -> Self {
        let users_panel = UsersPanelClient::new(settings.users_panel.base_url.clone());
        Self {
            db_pool,
            metabase_handle,
            settings,
            users_panel,
        }
    }
}
