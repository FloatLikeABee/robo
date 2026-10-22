use chrono::{Duration, Utc};
use jsonwebtoken::{DecodingKey, EncodingKey, Header, Validation, decode, encode};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Claims {
    pub uid: i64,
    pub org: i64,
    pub role: String,
    pub exp: i64,
}

#[derive(Clone)]
pub struct JwtService {
    secret: String,
    expiry_min: i64,
}

impl JwtService {
    pub fn new(secret: String, expiry_min: i64) -> Self {
        Self { secret, expiry_min }
    }

    pub fn issue(&self, user_id: i64, org_id: i64, role: &str) -> Result<String, jsonwebtoken::errors::Error> {
        let exp = (Utc::now() + Duration::minutes(self.expiry_min)).timestamp();
        let claims = Claims {
            uid: user_id,
            org: org_id,
            role: role.to_string(),
            exp,
        };
        encode(
            &Header::default(),
            &claims,
            &EncodingKey::from_secret(self.secret.as_bytes()),
        )
    }

    pub fn verify(&self, token: &str) -> Result<Claims, jsonwebtoken::errors::Error> {
        let data = decode::<Claims>(
            token,
            &DecodingKey::from_secret(self.secret.as_bytes()),
            &Validation::default(),
        )?;
        Ok(data.claims)
    }
}
