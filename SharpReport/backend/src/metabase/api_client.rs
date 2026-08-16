use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone)]
pub struct MetabaseApiClient {
    client: Client,
    base_url: String,
    session_token: Option<String>,
}

impl MetabaseApiClient {
    pub fn new(base_url: &str) -> Self {
        Self {
            client: Client::new(),
            base_url: base_url.trim_end_matches('/').to_string(),
            session_token: None,
        }
    }

    pub fn with_session_token(base_url: &str, session_token: &str) -> Self {
        Self {
            client: Client::new(),
            base_url: base_url.trim_end_matches('/').to_string(),
            session_token: Some(session_token.to_string()),
        }
    }

    pub async fn login(&mut self, email: &str, password: &str) -> Result<(), MetabaseApiError> {
        let url = format!("{}/api/session", self.base_url);
        let params = HashMap::from([("username", email), ("password", password)]);

        let response = self.client.post(&url).json(&params).send().await?;

        if !response.status().is_success() {
            return Err(MetabaseApiError::AuthenticationFailed);
        }

        let body: serde_json::Value = response.json().await?;
        if let Some(token) = body["id"].as_str() {
            self.session_token = Some(token.to_string());
            Ok(())
        } else {
            Err(MetabaseApiError::InvalidResponse)
        }
    }

    pub async fn get_session_token(&self) -> Option<&str> {
        self.session_token.as_deref()
    }

    async fn request<T: Serialize, R: for<'de> Deserialize<'de>>(
        &self,
        method: reqwest::Method,
        path: &str,
        body: Option<T>,
    ) -> Result<R, MetabaseApiError> {
        let url = format!("{}{}", self.base_url, path);
        let mut request = self.client.request(method, &url);

        if let Some(token) = &self.session_token {
            request = request.header("X-Metabase-Session", token);
        }

        if let Some(body) = body {
            request = request.json(&body);
        }

        let response = request.send().await?;

        if !response.status().is_success() {
            let status = response.status();
            let error_body = response.text().await?;
            return Err(MetabaseApiError::ApiError(status, error_body));
        }

        let result = response.json::<R>().await?;
        Ok(result)
    }

    pub async fn get_databases(&self) -> Result<Vec<Database>, MetabaseApiError> {
        self.request(reqwest::Method::GET, "/api/database", None::<()>)
            .await
    }

    pub async fn create_database(&self, database: &Database) -> Result<Database, MetabaseApiError> {
        self.request(reqwest::Method::POST, "/api/database", Some(database))
            .await
    }

    pub async fn get_database(&self, id: u64) -> Result<Database, MetabaseApiError> {
        self.request(
            reqwest::Method::GET,
            &format!("/api/database/{}", id),
            None::<()>,
        )
        .await
    }

    pub async fn update_database(
        &self,
        id: u64,
        database: &Database,
    ) -> Result<Database, MetabaseApiError> {
        self.request(
            reqwest::Method::PUT,
            &format!("/api/database/{}", id),
            Some(database),
        )
        .await
    }

    pub async fn delete_database(&self, id: u64) -> Result<(), MetabaseApiError> {
        self.request(
            reqwest::Method::DELETE,
            &format!("/api/database/{}", id),
            None::<()>,
        )
        .await
    }

    pub async fn get_dashboards(&self) -> Result<Vec<Dashboard>, MetabaseApiError> {
        self.request(reqwest::Method::GET, "/api/dashboard", None::<()>)
            .await
    }

    pub async fn get_dashboard(&self, id: u64) -> Result<Dashboard, MetabaseApiError> {
        self.request(
            reqwest::Method::GET,
            &format!("/api/dashboard/{}", id),
            None::<()>,
        )
        .await
    }

    pub async fn get_cards(&self) -> Result<Vec<Card>, MetabaseApiError> {
        self.request(reqwest::Method::GET, "/api/card", None::<()>)
            .await
    }

    pub async fn get_card(&self, id: u64) -> Result<Card, MetabaseApiError> {
        self.request(
            reqwest::Method::GET,
            &format!("/api/card/{}", id),
            None::<()>,
        )
        .await
    }

    pub async fn get_embed_url(
        &self,
        resource: &EmbedResource,
    ) -> Result<String, MetabaseApiError> {
        let resource_type = match resource {
            EmbedResource::Dashboard(id) => format!("dashboard/{}", id),
            EmbedResource::Card(id) => format!("card/{}", id),
        };

        let url = format!(
            "{}/embed/{}#bordered=true&titled=true",
            self.base_url, resource_type
        );
        Ok(url)
    }

    pub async fn get_signed_embed_token(
        &self,
        resource: &EmbedResource,
        params: &EmbedParams,
    ) -> Result<String, MetabaseApiError> {
        // This would require the Metabase secret key
        // For now, we'll return a placeholder
        Ok("placeholder-token".to_string())
    }
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Database {
    pub id: Option<u64>,
    pub name: String,
    pub engine: String,
    pub details: HashMap<String, serde_json::Value>,
    pub is_full_sync: bool,
    pub is_sample: bool,
    pub cache_field_values_schedule: Option<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Dashboard {
    pub id: u64,
    pub name: String,
    pub description: Option<String>,
    pub collection_id: Option<u64>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Card {
    pub id: u64,
    pub name: String,
    pub description: Option<String>,
    pub dataset_query: serde_json::Value,
    pub display: String,
    pub visualization_settings: serde_json::Value,
    pub collection_id: Option<u64>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug)]
pub enum EmbedResource {
    Dashboard(u64),
    Card(u64),
}

#[derive(Debug, Serialize)]
pub struct EmbedParams {
    pub bordered: Option<bool>,
    pub titled: Option<bool>,
    pub theme: Option<String>,
    pub params: Option<HashMap<String, String>>,
}

#[derive(Debug, thiserror::Error)]
pub enum MetabaseApiError {
    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),

    #[error("JSON error: {0}")]
    Json(#[from] serde_json::Error),

    #[error("Authentication failed")]
    AuthenticationFailed,

    #[error("Invalid API response")]
    InvalidResponse,

    #[error("API error: {0} - {1}")]
    ApiError(reqwest::StatusCode, String),
}
