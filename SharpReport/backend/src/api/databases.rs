use axum::{
    Json,
    extract::{Path, State},
    http::StatusCode,
};
use serde::{Deserialize, Serialize};
use sqlx::{
    mysql::{MySqlConnectOptions, MySqlSslMode},
    postgres::PgConnectOptions,
    sqlite::SqliteConnectOptions,
};
use std::str::FromStr;
use std::time::Duration;
use tokio::time::timeout;
use uuid::Uuid;

use crate::db::models::DatabaseConnection as StoredConnection;
use crate::db::repositories::DatabaseConnectionRepository;
use crate::services::AppState;
use crate::utils::crypto::CryptoService;

#[derive(Debug, Serialize, Deserialize)]
pub struct DatabaseConnection {
    pub id: String,
    pub name: String,
    pub engine: String,
    pub host: String,
    pub port: u16,
    pub database_name: String,
    pub username: String,
    pub ssl_enabled: bool,
}

fn port_u16(p: i32) -> u16 {
    p.clamp(0, 65535) as u16
}

fn to_api(row: StoredConnection) -> DatabaseConnection {
    DatabaseConnection {
        id: row.id.to_string(),
        name: row.name,
        engine: row.engine,
        host: row.host,
        port: port_u16(row.port),
        database_name: row.database_name,
        username: row.username,
        ssl_enabled: row.ssl_enabled,
    }
}

pub async fn list(
    State(state): State<AppState>,
) -> Result<Json<Vec<DatabaseConnection>>, (StatusCode, String)> {
    let repo = DatabaseConnectionRepository::new(state.db_pool.clone());
    let rows = repo
        .find_all()
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
    Ok(Json(rows.into_iter().map(to_api).collect()))
}

pub async fn create(
    State(_state): State<AppState>,
    Json(_request): Json<DatabaseConnection>,
) -> Result<Json<DatabaseConnection>, (StatusCode, String)> {
    Err((
        StatusCode::NOT_IMPLEMENTED,
        "Include a password field in the API to create connections here; setup wizard uses POST /api/v1/setup/database."
            .to_string(),
    ))
}

pub async fn get(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<DatabaseConnection>, (StatusCode, String)> {
    let uuid = Uuid::parse_str(&id)
        .map_err(|_| (StatusCode::BAD_REQUEST, "Invalid connection id".to_string()))?;
    let repo = DatabaseConnectionRepository::new(state.db_pool.clone());
    let row = repo
        .find_by_id(uuid)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .ok_or((
            StatusCode::NOT_FOUND,
            "Database connection not found".to_string(),
        ))?;
    Ok(Json(to_api(row)))
}

pub async fn update(
    State(_state): State<AppState>,
    Path(id): Path<String>,
    Json(request): Json<DatabaseConnection>,
) -> Result<Json<DatabaseConnection>, (StatusCode, String)> {
    Ok(Json(DatabaseConnection {
        id,
        name: request.name,
        engine: request.engine,
        host: request.host,
        port: request.port,
        database_name: request.database_name,
        username: request.username,
        ssl_enabled: request.ssl_enabled,
    }))
}

pub async fn delete(
    State(_state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    Ok(Json(serde_json::json!({"status": "deleted", "id": id})))
}

pub async fn test_connection(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    let uuid = Uuid::parse_str(&id)
        .map_err(|_| (StatusCode::BAD_REQUEST, "Invalid connection id".to_string()))?;
    let repo = DatabaseConnectionRepository::new(state.db_pool.clone());
    let conn = repo
        .find_by_id(uuid)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .ok_or((
            StatusCode::NOT_FOUND,
            "Database connection not found".to_string(),
        ))?;

    let crypto = CryptoService::new(&state.settings.jwt.secret);
    let password = crypto
        .decrypt(&conn.password_encrypted)
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let port = port_u16(conn.port);
    let engine = conn.engine.to_lowercase();
    let test_result = timeout(
        Duration::from_secs(10),
        test_engine_connection(&engine, &conn, &password, port),
    )
    .await;

    match test_result {
        Ok(Ok(())) => Ok(Json(serde_json::json!({
            "status": "ok",
            "id": id,
            "message": "Connection succeeded"
        }))),
        Ok(Err(msg)) => Ok(Json(serde_json::json!({
            "status": "failed",
            "id": id,
            "message": msg
        }))),
        Err(_) => Ok(Json(serde_json::json!({
            "status": "failed",
            "id": id,
            "message": "Connection timed out after 10s"
        }))),
    }
}

async fn test_engine_connection(
    engine: &str,
    conn: &StoredConnection,
    password: &str,
    port: u16,
) -> Result<(), String> {
    match engine {
        "postgres" | "postgresql" => {
            let mut opts = PgConnectOptions::new()
                .host(&conn.host)
                .port(port)
                .username(&conn.username)
                .password(password)
                .database(&conn.database_name);
            if conn.ssl_enabled {
                opts = opts.ssl_mode(sqlx::postgres::PgSslMode::Require);
            }
            let pool = sqlx::postgres::PgPoolOptions::new()
                .max_connections(1)
                .connect_with(opts)
                .await
                .map_err(|e| format!("PostgreSQL: {e}"))?;
            sqlx::query_scalar::<_, i32>("SELECT 1")
                .fetch_one(&pool)
                .await
                .map_err(|e| format!("PostgreSQL query: {e}"))?;
            pool.close().await;
            Ok(())
        }
        "mysql" | "mariadb" => {
            let mut opts = MySqlConnectOptions::new()
                .host(&conn.host)
                .port(port)
                .username(&conn.username)
                .password(password)
                .database(&conn.database_name);
            if conn.ssl_enabled {
                opts = opts.ssl_mode(MySqlSslMode::Required);
            }
            let pool = sqlx::mysql::MySqlPoolOptions::new()
                .max_connections(1)
                .connect_with(opts)
                .await
                .map_err(|e| format!("MySQL: {e}"))?;
            sqlx::query_scalar::<_, i64>("SELECT 1")
                .fetch_one(&pool)
                .await
                .map_err(|e| format!("MySQL query: {e}"))?;
            pool.close().await;
            Ok(())
        }
        "sqlite" => {
            let db_path = format!(
                "sqlite:///{}",
                conn.database_name
                    .trim_start_matches(|c| c == '/' || c == '\\')
            );
            let opts = SqliteConnectOptions::from_str(&db_path)
                .map_err(|e| e.to_string())?
                .create_if_missing(false);
            let pool = sqlx::sqlite::SqlitePoolOptions::new()
                .max_connections(1)
                .connect_with(opts)
                .await
                .map_err(|e| format!("SQLite: {e}"))?;
            sqlx::query_scalar::<_, i64>("SELECT 1")
                .fetch_one(&pool)
                .await
                .map_err(|e| format!("SQLite query: {e}"))?;
            pool.close().await;
            Ok(())
        }
        other => Err(format!(
            "Connection test is not implemented for engine {other}; supported: postgres, mysql/mariadb, sqlite"
        )),
    }
}
