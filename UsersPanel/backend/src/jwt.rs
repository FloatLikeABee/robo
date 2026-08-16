use chrono::{Duration, Utc};
use jsonwebtoken::{decode, encode, DecodingKey, EncodingKey, Header, Validation};
use serde::{Deserialize, Serialize};

use crate::config::Config;

#[derive(Debug, Serialize, Deserialize)]
pub struct Claims {
    pub sub: String,
    pub email: String,
    pub username: String,
    pub roles: Vec<String>,
    pub default_channel_id: String,
    pub exp: i64,
}

pub fn encode_token(
    cfg: &Config,
    user_id: &str,
    email: &str,
    username: &str,
    roles: &[String],
    default_channel_id: &str,
) -> Result<String, jsonwebtoken::errors::Error> {
    let exp = Utc::now() + Duration::hours(cfg.jwt_expiry_hours);
    let claims = Claims {
        sub: user_id.to_string(),
        email: email.to_string(),
        username: username.to_string(),
        roles: roles.to_vec(),
        default_channel_id: default_channel_id.to_string(),
        exp: exp.timestamp(),
    };
    encode(
        &Header::default(),
        &claims,
        &EncodingKey::from_secret(cfg.jwt_secret.as_bytes()),
    )
}

pub fn decode_token(cfg: &Config, token: &str) -> Result<Claims, jsonwebtoken::errors::Error> {
    decode::<Claims>(
        token,
        &DecodingKey::from_secret(cfg.jwt_secret.as_bytes()),
        &Validation::default(),
    )
    .map(|d| d.claims)
}
