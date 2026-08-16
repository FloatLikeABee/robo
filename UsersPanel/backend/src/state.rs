use oauth2::basic::BasicClient;
use oauth2::{
    AuthUrl, ClientId, ClientSecret, PkceCodeVerifier, RedirectUrl, TokenUrl,
};
use sqlx::MySqlPool;
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

use crate::config::Config;
use crate::importcol::jobs::ImportJobStore;

pub struct OAuthPending {
    pub pkce_verifier: PkceCodeVerifier,
}

pub struct AppState {
    pub pool: MySqlPool,
    pub config: Config,
    pub oauth_client: Option<Arc<BasicClient>>,
    /// CSRF state -> pending PKCE (in-memory; use Redis in production)
    pub oauth_pending: Arc<RwLock<HashMap<String, OAuthPending>>>,
    pub import_jobs: ImportJobStore,
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        Self {
            pool: self.pool.clone(),
            config: self.config.clone(),
            oauth_client: self.oauth_client.clone(),
            oauth_pending: Arc::clone(&self.oauth_pending),
            import_jobs: self.import_jobs.clone(),
        }
    }
}

impl AppState {
    pub fn new(pool: MySqlPool, config: Config) -> Self {
        let oauth_client = match (
            config.google_client_id.as_ref(),
            config.google_client_secret.as_ref(),
        ) {
            (Some(cid), Some(cs)) => {
                let client = BasicClient::new(
                    ClientId::new(cid.clone()),
                    Some(ClientSecret::new(cs.clone())),
                    AuthUrl::new("https://accounts.google.com/o/oauth2/v2/auth".to_string())
                        .expect("auth url"),
                    Some(
                        TokenUrl::new("https://oauth2.googleapis.com/token".to_string())
                            .expect("token url"),
                    ),
                )
                .set_redirect_uri(
                    RedirectUrl::new(config.google_redirect_url.clone()).expect("redirect"),
                );
                Some(Arc::new(client))
            }
            _ => None,
        };

        Self {
            pool,
            config,
            oauth_client,
            oauth_pending: Arc::new(RwLock::new(HashMap::new())),
            import_jobs: ImportJobStore::new(),
        }
    }
}
