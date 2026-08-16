use axum::{Json, extract::State, http::StatusCode};
use serde::{Deserialize, Serialize};
use serde_json::{Value, json};
use sqlx::{
    Column, Row,
    mysql::{MySqlConnectOptions, MySqlSslMode},
    postgres::PgConnectOptions,
};
use std::time::Duration;
use tokio::time::timeout;
use uuid::Uuid;

use crate::db::repositories::DatabaseConnectionRepository;
use crate::services::AppState;
use crate::utils::crypto::CryptoService;

#[derive(Debug, Serialize, Deserialize)]
pub struct SavedQuery {
    pub id: String,
    pub name: String,
    pub description: Option<String>,
    pub query_text: String,
    pub database_id: String,
    pub metabase_id: Option<u64>,
    pub is_favorite: bool,
}

pub async fn list(
    State(_state): State<AppState>,
) -> Result<Json<Vec<SavedQuery>>, (StatusCode, String)> {
    Ok(Json(vec![SavedQuery {
        id: "1".to_string(),
        name: "Customer Analysis".to_string(),
        description: Some("Query for customer segmentation".to_string()),
        query_text: "SELECT * FROM customers WHERE active = true".to_string(),
        database_id: "1".to_string(),
        metabase_id: Some(1),
        is_favorite: true,
    }]))
}

pub async fn create(
    State(_state): State<AppState>,
    Json(_request): Json<SavedQuery>,
) -> Result<Json<SavedQuery>, (StatusCode, String)> {
    Ok(Json(SavedQuery {
        id: "new-id".to_string(),
        name: "New Query".to_string(),
        description: None,
        query_text: "SELECT * FROM table".to_string(),
        database_id: "1".to_string(),
        metabase_id: None,
        is_favorite: false,
    }))
}

#[derive(Debug, Deserialize)]
pub struct ExecuteQueryRequest {
    pub database_id: String,
    pub sql: String,
}

fn pg_row_to_json_map(row: &sqlx::postgres::PgRow) -> serde_json::Map<String, Value> {
    let mut map = serde_json::Map::new();
    for (i, col) in row.columns().iter().enumerate() {
        let name = col.name().to_string();
        let v: Value = if let Ok(v) = row.try_get::<Option<String>, _>(i) {
            json!(v)
        } else if let Ok(v) = row.try_get::<Option<i64>, _>(i) {
            json!(v)
        } else if let Ok(v) = row.try_get::<Option<i32>, _>(i) {
            json!(v)
        } else if let Ok(v) = row.try_get::<Option<f64>, _>(i) {
            json!(v)
        } else if let Ok(v) = row.try_get::<Option<bool>, _>(i) {
            json!(v)
        } else if let Ok(uid) = row.try_get::<Option<uuid::Uuid>, _>(i) {
            json!(uid.map(|u| u.to_string()))
        } else {
            json!(null)
        };
        map.insert(name, v);
    }
    map
}

fn mysql_row_to_json_map(row: &sqlx::mysql::MySqlRow) -> serde_json::Map<String, Value> {
    let mut map = serde_json::Map::new();
    for (i, col) in row.columns().iter().enumerate() {
        let name = col.name().to_string();
        let v: Value = if let Ok(v) = row.try_get::<Option<String>, _>(i) {
            json!(v)
        } else if let Ok(v) = row.try_get::<Option<i64>, _>(i) {
            json!(v)
        } else if let Ok(v) = row.try_get::<Option<i32>, _>(i) {
            json!(v)
        } else if let Ok(v) = row.try_get::<Option<f64>, _>(i) {
            json!(v)
        } else if let Ok(v) = row.try_get::<Option<bool>, _>(i) {
            json!(v)
        } else if let Ok(uid) = row.try_get::<Option<uuid::Uuid>, _>(i) {
            json!(uid.map(|u| u.to_string()))
        } else {
            json!(null)
        };
        map.insert(name, v);
    }
    map
}

/// Run a read-only SQL query against a connected database. Used by the HTTP API and Data AI tools.
pub async fn execute_readonly_sql(
    state: &AppState,
    database_id: &str,
    sql: &str,
) -> Result<(usize, Vec<Value>), String> {
    let sql = sql.trim();
    if sql.is_empty() {
        return Err("sql must not be empty".into());
    }
    let lower = sql.to_lowercase();
    if !(lower.starts_with("select") || lower.starts_with("with")) {
        return Err("Only read-only SELECT / WITH queries are allowed".into());
    }
    if sql.contains(';') {
        return Err("Multiple statements are not allowed".into());
    }

    let uuid = Uuid::parse_str(database_id).map_err(|_| "Invalid database_id".to_string())?;

    let repo = DatabaseConnectionRepository::new(state.db_pool.clone());
    let conn = repo
        .find_by_id(uuid)
        .await
        .map_err(|e| e.to_string())?
        .ok_or_else(|| "Database connection not found".to_string())?;

    let engine = conn.engine.to_lowercase();
    if !matches!(
        engine.as_str(),
        "postgres" | "postgresql" | "mysql" | "mariadb"
    ) {
        return Err(
            "Native SQL execution supports PostgreSQL and MySQL/MariaDB; use Metabase for other engines."
                .into(),
        );
    }

    let crypto = CryptoService::new(&state.settings.jwt.secret);
    let password = crypto
        .decrypt(&conn.password_encrypted)
        .map_err(|e| e.to_string())?;

    let port = conn.port.clamp(0, 65535) as u16;

    let wrapped = format!(
        "SELECT * FROM ({}) AS datapulse_exec_subquery LIMIT 500",
        sql
    );

    let run = async {
        match engine.as_str() {
            "postgres" | "postgresql" => {
                let mut opts = PgConnectOptions::new()
                    .host(&conn.host)
                    .port(port)
                    .username(&conn.username)
                    .password(&password)
                    .database(&conn.database_name);
                if conn.ssl_enabled {
                    opts = opts.ssl_mode(sqlx::postgres::PgSslMode::Require);
                }
                let pool = sqlx::postgres::PgPoolOptions::new()
                    .max_connections(1)
                    .connect_with(opts)
                    .await
                    .map_err(|e| format!("Connect: {e}"))?;
                let rows = sqlx::query(&wrapped)
                    .fetch_all(&pool)
                    .await
                    .map_err(|e| format!("Query: {e}"))?;
                pool.close().await;
                let out: Vec<Value> = rows
                    .iter()
                    .map(|r| Value::Object(pg_row_to_json_map(r)))
                    .collect();
                Ok(out)
            }
            "mysql" | "mariadb" => {
                let mut opts = MySqlConnectOptions::new()
                    .host(&conn.host)
                    .port(port)
                    .username(&conn.username)
                    .password(&password)
                    .database(&conn.database_name);
                if conn.ssl_enabled {
                    opts = opts.ssl_mode(MySqlSslMode::Required);
                }
                let pool = sqlx::mysql::MySqlPoolOptions::new()
                    .max_connections(1)
                    .connect_with(opts)
                    .await
                    .map_err(|e| format!("Connect: {e}"))?;
                let rows = sqlx::query(&wrapped)
                    .fetch_all(&pool)
                    .await
                    .map_err(|e| format!("Query: {e}"))?;
                pool.close().await;
                let out: Vec<Value> = rows
                    .iter()
                    .map(|r| Value::Object(mysql_row_to_json_map(r)))
                    .collect();
                Ok(out)
            }
            _ => unreachable!(),
        }
    };

    match timeout(Duration::from_secs(30), run).await {
        Ok(Ok(rows)) => Ok((rows.len(), rows)),
        Ok(Err(msg)) => Err(msg),
        Err(_) => Err("Query timed out after 30s".into()),
    }
}

pub async fn execute(
    State(state): State<AppState>,
    Json(request): Json<ExecuteQueryRequest>,
) -> Result<Json<Value>, (StatusCode, String)> {
    match execute_readonly_sql(&state, &request.database_id, &request.sql).await {
        Ok((row_count, results)) => Ok(Json(json!({
            "status": "success",
            "row_count": row_count,
            "results": results
        }))),
        Err(msg) => {
            let status = if msg.contains("not found") || msg.contains("Invalid") {
                StatusCode::BAD_REQUEST
            } else {
                StatusCode::BAD_REQUEST
            };
            Err((status, msg))
        }
    }
}
