use crate::db::models::User;
use crate::db::repositories::UserRepository;
use argon2::{
    self, Argon2,
    password_hash::{PasswordHash, PasswordHasher, PasswordVerifier, SaltString},
};
use chrono::{Duration, Utc};
use jsonwebtoken::{DecodingKey, EncodingKey, Header, Validation, decode, encode};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Serialize, Deserialize)]
pub struct Claims {
    pub sub: String,
    pub exp: i64,
    pub role: String,
}

#[derive(Debug)]
pub struct AuthService {
    users: UserRepository,
    jwt_secret: String,
    jwt_expiry_hours: i64,
}

impl AuthService {
    pub fn new(db: crate::db::Database, jwt_secret: &str, jwt_expiry_hours: i64) -> Self {
        Self {
            users: UserRepository::new(db),
            jwt_secret: jwt_secret.to_string(),
            jwt_expiry_hours,
        }
    }

    pub async fn register_user(
        &self,
        email: &str,
        password: &str,
        name: &str,
    ) -> Result<User, AuthError> {
        self.register_user_with_role(email, password, name, "viewer")
            .await
    }

    /// First-time setup: create the primary admin account.
    pub async fn create_admin_user(
        &self,
        email: &str,
        password: &str,
        name: &str,
    ) -> Result<User, AuthError> {
        self.register_user_with_role(email, password, name, "admin")
            .await
    }

    async fn register_user_with_role(
        &self,
        email: &str,
        password: &str,
        name: &str,
        role: &str,
    ) -> Result<User, AuthError> {
        let existing_user = self.get_user_by_email(email).await?;
        if existing_user.is_some() {
            return Err(AuthError::UserAlreadyExists);
        }

        let password_hash = self.hash_password(password)?;

        let user = User {
            id: Uuid::new_v4(),
            email: email.to_string(),
            password_hash,
            name: name.to_string(),
            role: role.to_string(),
            avatar_url: None,
            metabase_user_id: None,
            created_at: Utc::now(),
            updated_at: Utc::now(),
        };

        self.users.insert(&user).await?;

        Ok(user)
    }

    pub async fn login(&self, email: &str, password: &str) -> Result<String, AuthError> {
        let user = self.get_user_by_email(email).await?;

        let user = match user {
            Some(u) => u,
            None => return Err(AuthError::InvalidCredentials),
        };

        self.verify_password(&user.password_hash, password)
            .map_err(|_| AuthError::InvalidCredentials)?;

        // Generate JWT token
        self.generate_token(&user)
    }

    pub async fn get_user_by_email(&self, email: &str) -> Result<Option<User>, AuthError> {
        Ok(self.users.find_by_email(email).await?)
    }

    pub async fn get_user_by_id(&self, id: Uuid) -> Result<Option<User>, AuthError> {
        Ok(self.users.find_by_id(id).await?)
    }

    pub async fn validate_token(&self, token: &str) -> Result<Claims, AuthError> {
        let decoding_key = DecodingKey::from_secret(self.jwt_secret.as_bytes());
        let validation = Validation::default();

        let token_data = decode::<Claims>(token, &decoding_key, &validation)
            .map_err(|_| AuthError::InvalidToken)?;

        Ok(token_data.claims)
    }

    fn generate_token(&self, user: &User) -> Result<String, AuthError> {
        let expiration = Utc::now() + Duration::hours(self.jwt_expiry_hours);
        let claims = Claims {
            sub: user.id.to_string(),
            exp: expiration.timestamp(),
            role: user.role.clone(),
        };

        let encoding_key = EncodingKey::from_secret(self.jwt_secret.as_bytes());
        encode(&Header::default(), &claims, &encoding_key).map_err(|_| AuthError::TokenGeneration)
    }

    fn hash_password(&self, password: &str) -> Result<String, AuthError> {
        let salt = SaltString::generate(&mut rand::thread_rng());
        let argon2 = Argon2::default();
        argon2
            .hash_password(password.as_bytes(), &salt)
            .map(|hash| hash.to_string())
            .map_err(|_| AuthError::PasswordHashing)
    }

    fn verify_password(&self, hash: &str, password: &str) -> Result<(), AuthError> {
        let parsed_hash = PasswordHash::new(hash).map_err(|_| AuthError::PasswordVerification)?;
        Argon2::default()
            .verify_password(password.as_bytes(), &parsed_hash)
            .map_err(|_| AuthError::PasswordVerification)
    }
}

#[derive(Debug, thiserror::Error)]
pub enum AuthError {
    #[error("Database error: {0}")]
    Database(#[from] sqlx::Error),

    #[error("User already exists")]
    UserAlreadyExists,

    #[error("Invalid credentials")]
    InvalidCredentials,

    #[error("Invalid token")]
    InvalidToken,

    #[error("Token generation failed")]
    TokenGeneration,

    #[error("Password hashing failed")]
    PasswordHashing,

    #[error("Password verification failed")]
    PasswordVerification,
}
