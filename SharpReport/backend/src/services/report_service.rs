use crate::db::Database;
use crate::metabase::api_client::MetabaseApiClient;

#[derive(Debug)]
pub struct ReportService {
    db: Database,
    metabase_client: MetabaseApiClient,
}

impl ReportService {
    pub fn new(db: Database, metabase_url: &str) -> Self {
        Self {
            db,
            metabase_client: MetabaseApiClient::new(metabase_url),
        }
    }

    pub async fn get_dashboards(
        &self,
    ) -> Result<
        Vec<crate::metabase::api_client::Dashboard>,
        crate::metabase::api_client::MetabaseApiError,
    > {
        self.metabase_client.get_dashboards().await
    }

    pub async fn get_dashboard(
        &self,
        id: u64,
    ) -> Result<crate::metabase::api_client::Dashboard, crate::metabase::api_client::MetabaseApiError>
    {
        self.metabase_client.get_dashboard(id).await
    }

    pub async fn get_embed_url(
        &self,
        resource_type: &str,
        id: u64,
    ) -> Result<String, crate::metabase::api_client::MetabaseApiError> {
        match resource_type {
            "dashboard" => {
                let resource = crate::metabase::api_client::EmbedResource::Dashboard(id);
                self.metabase_client.get_embed_url(&resource).await
            }
            "card" => {
                let resource = crate::metabase::api_client::EmbedResource::Card(id);
                self.metabase_client.get_embed_url(&resource).await
            }
            _ => Err(crate::metabase::api_client::MetabaseApiError::InvalidResponse),
        }
    }

    pub async fn get_signed_embed_token(
        &self,
        resource_type: &str,
        id: u64,
        params: std::collections::HashMap<String, String>,
    ) -> Result<String, crate::metabase::api_client::MetabaseApiError> {
        match resource_type {
            "dashboard" => {
                let resource = crate::metabase::api_client::EmbedResource::Dashboard(id);
                let embed_params = crate::metabase::api_client::EmbedParams {
                    bordered: Some(true),
                    titled: Some(true),
                    theme: None,
                    params: Some(params),
                };
                self.metabase_client
                    .get_signed_embed_token(&resource, &embed_params)
                    .await
            }
            "card" => {
                let resource = crate::metabase::api_client::EmbedResource::Card(id);
                let embed_params = crate::metabase::api_client::EmbedParams {
                    bordered: Some(true),
                    titled: Some(true),
                    theme: None,
                    params: Some(params),
                };
                self.metabase_client
                    .get_signed_embed_token(&resource, &embed_params)
                    .await
            }
            _ => Err(crate::metabase::api_client::MetabaseApiError::InvalidResponse),
        }
    }
}
