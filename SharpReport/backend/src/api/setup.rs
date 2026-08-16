use axum::{Json, extract::State, http::StatusCode, response::IntoResponse};
use chrono::Utc;
use serde::{Deserialize, Deserializer, Serialize};
use tracing::warn;
use uuid::Uuid;

use crate::db::models::DatabaseConnection as StoredDatabaseConnection;
use crate::db::repositories::DatabaseConnectionRepository;
use crate::services::AppState;
use crate::services::auth_service::{AuthError, AuthService};
use crate::utils::crypto::CryptoService;

fn deserialize_port<'de, D>(deserializer: D) -> Result<u16, D::Error>
where
    D: Deserializer<'de>,
{
    struct PortVisitor;

    impl serde::de::Visitor<'_> for PortVisitor {
        type Value = u16;

        fn expecting(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
            f.write_str("a database port number or numeric string")
        }

        fn visit_u64<E: serde::de::Error>(self, v: u64) -> Result<u16, E> {
            u16::try_from(v).map_err(|_| E::custom("port out of range for u16"))
        }

        fn visit_i64<E: serde::de::Error>(self, v: i64) -> Result<u16, E> {
            u16::try_from(v).map_err(|_| E::custom("port out of range for u16"))
        }

        fn visit_str<E: serde::de::Error>(self, v: &str) -> Result<u16, E> {
            v.trim()
                .parse()
                .map_err(|_| E::custom("invalid port string"))
        }
    }

    deserializer.deserialize_any(PortVisitor)
}

#[derive(Debug, Serialize)]
pub struct SetupStatusResponse {
    pub is_completed: bool,
    pub metabase_initialized: bool,
    pub admin_user_created: bool,
    pub metabase_port: u16,
}

pub async fn status(
    State(state): State<AppState>,
) -> Result<Json<SetupStatusResponse>, (StatusCode, String)> {
    Ok(Json(SetupStatusResponse {
        is_completed: false,
        metabase_initialized: false,
        admin_user_created: false,
        metabase_port: state.metabase_handle.get_port(),
    }))
}

#[derive(Debug, Deserialize)]
pub struct InitializeRequest {
    pub jvm_path: Option<String>,
}

pub async fn initialize(
    State(state): State<AppState>,
    Json(_request): Json<InitializeRequest>,
) -> impl IntoResponse {
    let port = state.metabase_handle.get_port();
    let host = state.metabase_handle.get_host().to_string();

    match state.metabase_handle.health_check().await {
        Ok(true) => (
            StatusCode::OK,
            Json(serde_json::json!({
                "status": "initialized",
                "metabase_port": port,
            })),
        )
            .into_response(),
        Ok(false) => {
            warn!(%host, %port, "Metabase /api/health returned non-success");
            (
                StatusCode::SERVICE_UNAVAILABLE,
                Json(serde_json::json!({
                    "status": "error",
                    "error": format!(
                        "Metabase at http://{host}:{port} is not healthy yet. Start Metabase or wait until it finishes booting."
                    ),
                    "metabase_port": port,
                })),
            )
                .into_response()
        }
        Err(e) => {
            warn!(error = %e, %host, %port, "Metabase health check failed");
            (
                StatusCode::SERVICE_UNAVAILABLE,
                Json(serde_json::json!({
                    "status": "error",
                    "error": format!(
                        "Cannot reach Metabase at http://{host}:{port}: {e}. Start Metabase (see config metabase.host/port) or enable metabase.autostart."
                    ),
                    "metabase_port": port,
                })),
            )
                .into_response()
        }
    }
}

#[derive(Debug, Deserialize)]
pub struct CreateAdminRequest {
    pub email: String,
    pub password: String,
    pub name: String,
}

pub async fn create_admin(
    State(state): State<AppState>,
    Json(request): Json<CreateAdminRequest>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    let auth = AuthService::new(
        state.db_pool.clone(),
        &state.settings.jwt.secret,
        state.settings.jwt.expiry_hours as i64,
    );

    match auth
        .create_admin_user(&request.email, &request.password, &request.name)
        .await
    {
        Ok(user) => Ok(Json(serde_json::json!({
            "status": "admin_created",
            "email": user.email,
            "id": user.id.to_string(),
        }))),
        Err(AuthError::UserAlreadyExists) => Err((
            StatusCode::CONFLICT,
            "A user with this email already exists".to_string(),
        )),
        Err(e) => Err((StatusCode::INTERNAL_SERVER_ERROR, e.to_string())),
    }
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AddDatabaseRequest {
    pub name: String,
    pub engine: String,
    pub host: String,
    #[serde(deserialize_with = "deserialize_port")]
    pub port: u16,
    pub database_name: String,
    pub username: String,
    pub password: String,
    pub ssl_enabled: bool,
}

pub async fn add_database(
    State(state): State<AppState>,
    Json(request): Json<AddDatabaseRequest>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    let crypto = CryptoService::new(&state.settings.jwt.secret);
    let password_encrypted = crypto
        .encrypt(&request.password)
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let now = Utc::now();
    let conn = StoredDatabaseConnection {
        id: Uuid::new_v4(),
        name: request.name,
        engine: request.engine,
        host: request.host,
        port: i32::from(request.port),
        database_name: request.database_name,
        username: request.username,
        password_encrypted,
        ssl_enabled: request.ssl_enabled,
        ssl_config: None,
        additional_config: None,
        metabase_database_id: None,
        is_active: true,
        created_at: now,
        updated_at: now,
    };

    let repo = DatabaseConnectionRepository::new(state.db_pool.clone());
    let saved = repo
        .insert(&conn)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    Ok(Json(serde_json::json!({
        "status": "database_added",
        "id": saved.id.to_string(),
        "name": saved.name,
    })))
}
