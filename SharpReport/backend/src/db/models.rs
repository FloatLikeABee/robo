use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::{FromRow, Postgres, Sqlite};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct User {
    pub id: Uuid,
    pub email: String,
    pub password_hash: String,
    pub name: String,
    pub role: String,
    pub avatar_url: Option<String>,
    pub metabase_user_id: Option<i32>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct DatabaseConnection {
    pub id: Uuid,
    pub name: String,
    pub engine: String,
    pub host: String,
    pub port: i32,
    pub database_name: String,
    pub username: String,
    pub password_encrypted: String,
    pub ssl_enabled: bool,
    pub ssl_config: Option<serde_json::Value>,
    pub additional_config: Option<serde_json::Value>,
    pub metabase_database_id: Option<i32>,
    pub is_active: bool,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Dashboard {
    pub id: Uuid,
    pub name: String,
    pub description: Option<String>,
    pub metabase_dashboard_id: Option<i32>,
    pub database_id: Uuid,
    pub layout_config: Option<serde_json::Value>,
    pub is_public: bool,
    pub created_by: Uuid,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SavedQuery {
    pub id: Uuid,
    pub name: String,
    pub description: Option<String>,
    pub query_text: String,
    pub database_id: Uuid,
    pub metabase_card_id: Option<i32>,
    pub created_by: Uuid,
    pub is_favorite: bool,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApiKey {
    pub id: Uuid,
    pub name: String,
    pub key_hash: String,
    pub user_id: Uuid,
    pub permissions: serde_json::Value,
    pub expires_at: Option<DateTime<Utc>>,
    pub last_used_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuditLog {
    pub id: Uuid,
    pub user_id: Uuid,
    pub action: String,
    pub resource_type: Option<String>,
    pub resource_id: Option<Uuid>,
    pub details: Option<serde_json::Value>,
    pub ip_address: Option<String>,
    pub user_agent: Option<String>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SetupState {
    pub id: i32,
    pub is_completed: bool,
    pub metabase_initialized: bool,
    pub metabase_port: i32,
    pub metabase_pid: Option<i32>,
    pub admin_user_created: bool,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}
