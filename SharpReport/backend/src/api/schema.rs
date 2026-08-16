//! Table/column introspection for connected databases (PostgreSQL, MySQL/MariaDB, SQLite).

use axum::{
    Json,
    extract::{Path, State},
    http::StatusCode,
};
use serde::Serialize;
use sqlx::{
    mysql::{MySqlConnectOptions, MySqlSslMode},
    postgres::PgConnectOptions,
    sqlite::SqliteConnectOptions,
};
use std::str::FromStr;
use uuid::Uuid;

use crate::db::models::DatabaseConnection as StoredConnection;
use crate::db::repositories::DatabaseConnectionRepository;
use crate::services::AppState;
use crate::utils::crypto::CryptoService;

/// Returns connection row + table list (for report builder validation).
pub(crate) async fn load_schema_snapshot(
    state: &AppState,
    id: &Uuid,
) -> Result<(StoredConnection, Vec<TableInfo>), String> {
    let repo = DatabaseConnectionRepository::new(state.db_pool.clone());
    let conn = repo
        .find_by_id(*id)
        .await
        .map_err(|e| e.to_string())?
        .ok_or_else(|| "Database connection not found".to_string())?;
    let crypto = CryptoService::new(&state.settings.jwt.secret);
    let password = crypto
        .decrypt(&conn.password_encrypted)
        .map_err(|e| e.to_string())?;
    let engine = conn.engine.to_lowercase();
    let tables = match engine.as_str() {
        "postgres" | "postgresql" => introspect_postgres(&conn, &password).await,
        "mysql" | "mariadb" => introspect_mysql(&conn, &password).await,
        "sqlite" => introspect_sqlite(&conn).await,
        other => Err(format!("Schema not supported for engine: {other}")),
    }?;
    Ok((conn, tables))
}

pub(crate) fn column_exists(tables: &[TableInfo], schema: &str, table: &str, column: &str) -> bool {
    tables.iter().any(|t| {
        t.schema == schema && t.name == table && t.columns.iter().any(|c| c.name == column)
    })
}

#[derive(Debug, Clone, Serialize)]
pub struct ColumnInfo {
    pub name: String,
    pub data_type: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct TableInfo {
    pub schema: String,
    pub name: String,
    pub columns: Vec<ColumnInfo>,
}

#[derive(Debug, Serialize)]
pub struct SchemaResponse {
    pub database_id: String,
    pub engine: String,
    pub tables: Vec<TableInfo>,
}

fn port_u16(p: i32) -> u16 {
    p.clamp(0, 65535) as u16
}

pub async fn get_schema(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<SchemaResponse>, (StatusCode, String)> {
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

    let engine = conn.engine.to_lowercase();
    let tables = match engine.as_str() {
        "postgres" | "postgresql" => introspect_postgres(&conn, &password).await,
        "mysql" | "mariadb" => introspect_mysql(&conn, &password).await,
        "sqlite" => introspect_sqlite(&conn).await,
        other => Err(format!(
            "Schema introspection not supported for engine: {other}"
        )),
    }
    .map_err(|e| (StatusCode::BAD_REQUEST, e))?;

    Ok(Json(SchemaResponse {
        database_id: id,
        engine: conn.engine,
        tables,
    }))
}

async fn introspect_postgres(
    conn: &StoredConnection,
    password: &str,
) -> Result<Vec<TableInfo>, String> {
    let port = port_u16(conn.port);
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

    let rows = sqlx::query_as::<_, (String, String, String, String)>(
        r#"
        SELECT c.table_schema::text, c.table_name::text, c.column_name::text, c.data_type::text
        FROM information_schema.columns c
        WHERE c.table_schema NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
          AND c.table_schema NOT LIKE 'pg\_%' ESCAPE '\'
        ORDER BY c.table_schema, c.table_name, c.ordinal_position
        "#,
    )
    .fetch_all(&pool)
    .await
    .map_err(|e| format!("PostgreSQL schema: {e}"))?;
    pool.close().await;

    fold_columns(rows)
}

async fn introspect_mysql(
    conn: &StoredConnection,
    password: &str,
) -> Result<Vec<TableInfo>, String> {
    let port = port_u16(conn.port);
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

    let rows = sqlx::query_as::<_, (String, String, String, String)>(
        r#"
        SELECT TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME, DATA_TYPE
        FROM information_schema.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
        ORDER BY TABLE_NAME, ORDINAL_POSITION
        "#,
    )
    .fetch_all(&pool)
    .await
    .map_err(|e| format!("MySQL schema: {e}"))?;
    pool.close().await;

    fold_columns(rows)
}

async fn introspect_sqlite(conn: &StoredConnection) -> Result<Vec<TableInfo>, String> {
    let db_path = format!(
        "sqlite:///{}",
        conn.database_name
            .trim()
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

    let table_names: Vec<(String,)> =
        sqlx::query_as(r#"SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name"#)
            .fetch_all(&pool)
            .await
            .map_err(|e| format!("SQLite tables: {e}"))?;

    let mut tables = Vec::new();
    for (tname,) in table_names {
        let cols: Vec<(String, String)> =
            sqlx::query_as("SELECT name, type FROM pragma_table_info(?)")
                .bind(&tname)
                .fetch_all(&pool)
                .await
                .map_err(|e| format!("SQLite columns {tname}: {e}"))?;

        let columns = cols
            .into_iter()
            .map(|(name, typ)| ColumnInfo {
                name,
                data_type: typ,
            })
            .collect();

        tables.push(TableInfo {
            schema: "main".to_string(),
            name: tname,
            columns,
        });
    }

    pool.close().await;
    Ok(tables)
}

fn fold_columns(rows: Vec<(String, String, String, String)>) -> Result<Vec<TableInfo>, String> {
    use std::collections::BTreeMap;
    let mut map: BTreeMap<(String, String), Vec<ColumnInfo>> = BTreeMap::new();
    for (schema, table, col, dtype) in rows {
        map.entry((schema, table)).or_default().push(ColumnInfo {
            name: col,
            data_type: dtype,
        });
    }
    Ok(map
        .into_iter()
        .map(|((schema, name), columns)| TableInfo {
            schema,
            name,
            columns,
        })
        .collect())
}
