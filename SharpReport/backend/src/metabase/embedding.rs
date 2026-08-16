use chrono::{Duration, Utc};
use jsonwebtoken::{DecodingKey, EncodingKey, Header, Validation, decode, encode};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Serialize, Deserialize)]
pub struct EmbedClaims {
    pub resource: Resource,
    pub params: HashMap<String, String>,
    pub exp: i64,
}

#[derive(Debug, Serialize, Deserialize)]
pub enum Resource {
    #[serde(rename = "dashboard")]
    Dashboard(u64),
    #[serde(rename = "card")]
    Card(u64),
}

#[derive(Debug)]
pub struct EmbedTokenGenerator {
    secret: String,
}

impl EmbedTokenGenerator {
    pub fn new(secret: &str) -> Self {
        Self {
            secret: secret.to_string(),
        }
    }

    pub fn generate_token(
        &self,
        resource: Resource,
        params: HashMap<String, String>,
        expiry_minutes: i64,
    ) -> Result<String, EmbeddingError> {
        let expiration = Utc::now() + Duration::minutes(expiry_minutes);
        let claims = EmbedClaims {
            resource,
            params,
            exp: expiration.timestamp(),
        };

        let encoding_key = EncodingKey::from_secret(self.secret.as_bytes());
        encode(&Header::default(), &claims, &encoding_key)
            .map_err(|e| EmbeddingError::TokenGeneration(e.to_string()))
    }

    pub fn validate_token(&self, token: &str) -> Result<EmbedClaims, EmbeddingError> {
        let decoding_key = DecodingKey::from_secret(self.secret.as_bytes());
        let validation = Validation::default();

        decode::<EmbedClaims>(token, &decoding_key, &validation)
            .map(|token_data| token_data.claims)
            .map_err(|e| EmbeddingError::TokenValidation(e.to_string()))
    }
}

#[derive(Debug, thiserror::Error)]
pub enum EmbeddingError {
    #[error("Token generation error: {0}")]
    TokenGeneration(String),

    #[error("Token validation error: {0}")]
    TokenValidation(String),

    #[error("Invalid resource type")]
    InvalidResource,
}
