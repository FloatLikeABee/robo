//! Visual report builder: validated SELECT from a single table with filters, aggregates, order, limit.

use axum::{Json, extract::State, http::StatusCode};
use serde::{Deserialize, Serialize};
use serde_json::{Value, json};
use sqlx::{
    Column, Row,
    mysql::{MySqlConnectOptions, MySqlSslMode},
    postgres::PgConnectOptions,
    sqlite::SqliteConnectOptions,
};
use std::str::FromStr;
use uuid::Uuid;

use crate::api::schema::{TableInfo, column_exists, load_schema_snapshot};
use crate::db::models::DatabaseConnection as StoredConnection;
use crate::services::AppState;
use crate::utils::crypto::CryptoService;

#[derive(Debug, Deserialize)]
pub struct ReportColumnSpec {
    pub schema: String,
    pub table: String,
    pub column: String,
    pub alias: Option<String>,
    pub aggregation: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct ReportFilterSpec {
    pub schema: String,
    pub table: String,
    pub column: String,
    pub op: String,
    pub value: Value,
}

#[derive(Debug, Deserialize)]
pub struct ReportOrderSpec {
    pub column_index: usize,
    pub dir: String,
}

#[derive(Debug, Deserialize)]
pub struct ReportExecuteRequest {
    pub database_id: String,
    pub columns: Vec<ReportColumnSpec>,
    pub filters: Vec<ReportFilterSpec>,
    pub order_by: Option<Vec<ReportOrderSpec>>,
    #[serde(default = "default_limit")]
    pub limit: u32,
}

fn default_limit() -> u32 {
    500
}

#[derive(Debug, Serialize)]
pub struct ReportExecuteResponse {
    pub columns: Vec<String>,
    pub rows: Vec<Value>,
    pub row_count: usize,
}

fn port_u16(p: i32) -> u16 {
    p.clamp(0, 65535) as u16
}

fn validate_single_table(
    cols: &[ReportColumnSpec],
    filters: &[ReportFilterSpec],
) -> Result<(), String> {
    if cols.is_empty() {
        return Err("At least one column is required".into());
    }
    let (s, t) = (cols[0].schema.clone(), cols[0].table.clone());
    for c in cols {
        if c.schema != s || c.table != t {
            return Err("All columns must belong to the same table (single-table reports)".into());
        }
    }
    for f in filters {
        if f.schema != s || f.table != t {
            return Err("Filters must target the same table as columns".into());
        }
    }
    Ok(())
}

fn validate_against_schema(
    tables: &[TableInfo],
    cols: &[ReportColumnSpec],
    filters: &[ReportFilterSpec],
) -> Result<(), String> {
    for c in cols {
        if !column_exists(tables, &c.schema, &c.table, &c.column) {
            return Err(format!(
                "Unknown column {}.{}.{}",
                c.schema, c.table, c.column
            ));
        }
    }
    for f in filters {
        if !column_exists(tables, &f.schema, &f.table, &f.column) {
            return Err(format!(
                "Unknown filter column {}.{}.{}",
                f.schema, f.table, f.column
            ));
        }
    }
    Ok(())
}

pub async fn execute(
    State(state): State<AppState>,
    Json(req): Json<ReportExecuteRequest>,
) -> Result<Json<ReportExecuteResponse>, (StatusCode, String)> {
    let uuid = Uuid::parse_str(&req.database_id)
        .map_err(|_| (StatusCode::BAD_REQUEST, "Invalid database_id".to_string()))?;

    let (conn, tables) = load_schema_snapshot(&state, &uuid)
        .await
        .map_err(|e| (StatusCode::BAD_REQUEST, e))?;

    validate_against_schema(&tables, &req.columns, &req.filters)
        .map_err(|e| (StatusCode::BAD_REQUEST, e))?;

    validate_single_table(&req.columns, &req.filters).map_err(|e| (StatusCode::BAD_REQUEST, e))?;

    let crypto = CryptoService::new(&state.settings.jwt.secret);
    let password = crypto
        .decrypt(&conn.password_encrypted)
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let limit = req.limit.clamp(1, 2000);

    let engine = conn.engine.to_lowercase();
    match engine.as_str() {
        "postgres" | "postgresql" => run_postgres(&conn, &password, &req, limit).await,
        "mysql" | "mariadb" => run_mysql(&conn, &password, &req, limit).await,
        "sqlite" => run_sqlite(&conn, &req, limit).await,
        other => Err((
            StatusCode::BAD_REQUEST,
            format!("Report builder not implemented for {other}"),
        )),
    }
}

fn pg_ident(s: &str) -> String {
    format!("\"{}\"", s.replace('"', "\"\""))
}

fn mysql_ident(s: &str) -> String {
    format!("`{}`", s.replace('`', "``"))
}

fn build_select_pg(
    cols: &[ReportColumnSpec],
    alias: &str,
) -> Result<(Vec<String>, Vec<String>, Vec<String>), String> {
    let mut select = Vec::new();
    let mut group_by: Vec<String> = Vec::new();
    let mut out_names: Vec<String> = Vec::new();
    for c in cols {
        let col_ref = format!("{}.{}", alias, pg_ident(&c.column));
        let agg = c.aggregation.as_deref().unwrap_or("none");
        let alias_name = c.alias.clone().unwrap_or_else(|| c.column.clone());
        let (expr, gb) = match agg {
            "none" | "" => (col_ref.clone(), Some(col_ref)),
            "sum" => (
                format!("SUM({}) AS {}", col_ref, pg_ident(&alias_name)),
                None,
            ),
            "avg" => (
                format!("AVG({}) AS {}", col_ref, pg_ident(&alias_name)),
                None,
            ),
            "count" => (
                format!("COUNT({}) AS {}", col_ref, pg_ident(&alias_name)),
                None,
            ),
            "min" => (
                format!("MIN({}) AS {}", col_ref, pg_ident(&alias_name)),
                None,
            ),
            "max" => (
                format!("MAX({}) AS {}", col_ref, pg_ident(&alias_name)),
                None,
            ),
            a => return Err(format!("Unsupported aggregation: {a}")),
        };
        select.push(expr);
        out_names.push(alias_name);
        if let Some(g) = gb {
            group_by.push(g);
        }
    }
    Ok((select, group_by, out_names))
}

fn filter_clause_pg(
    alias: &str,
    f: &ReportFilterSpec,
    param: &mut usize,
) -> Result<String, String> {
    let col = format!("{}.{}", alias, pg_ident(&f.column));
    match f.op.as_str() {
        "isnotnull" => Ok(format!("{col} IS NOT NULL")),
        other => {
            let p = *param;
            *param += 1;
            let ph = format!("${p}");
            Ok(match other {
                "eq" => format!("{col} = {ph}"),
                "ne" => format!("{col} <> {ph}"),
                "gt" => format!("{col} > {ph}"),
                "gte" => format!("{col} >= {ph}"),
                "lt" => format!("{col} < {ph}"),
                "lte" => format!("{col} <= {ph}"),
                "contains" => format!("CAST({col} AS TEXT) ILIKE {ph}"),
                "starts_with" => format!("CAST({col} AS TEXT) ILIKE {ph}"),
                o => return Err(format!("Unsupported filter op: {o}")),
            })
        }
    }
}

fn bind_string(f: &ReportFilterSpec) -> Result<String, String> {
    match &f.value {
        Value::Null => Err("NULL filters not supported in builder".into()),
        Value::String(s) => {
            if f.op == "contains" {
                Ok(format!("%{s}%"))
            } else if f.op == "starts_with" {
                Ok(format!("{s}%"))
            } else {
                Ok(s.clone())
            }
        }
        Value::Number(n) => Ok(n.to_string()),
        Value::Bool(b) => Ok(b.to_string()),
        _ => Err("Unsupported filter value".into()),
    }
}

fn pg_row_map(row: &sqlx::postgres::PgRow) -> serde_json::Map<String, Value> {
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
        } else {
            json!(null)
        };
        map.insert(name, v);
    }
    map
}

async fn run_postgres(
    conn: &StoredConnection,
    password: &str,
    req: &ReportExecuteRequest,
    limit: u32,
) -> Result<Json<ReportExecuteResponse>, (StatusCode, String)> {
    let (select, group_by, out_names) =
        build_select_pg(&req.columns, "t").map_err(|e| (StatusCode::BAD_REQUEST, e))?;

    let from = format!(
        "{}.{} AS t",
        pg_ident(&req.columns[0].schema),
        pg_ident(&req.columns[0].table)
    );

    let mut sql = format!("SELECT {} FROM {}", select.join(", "), from);
    let mut binds: Vec<String> = Vec::new();

    if !req.filters.is_empty() {
        let mut param = 1usize;
        let parts: Result<Vec<_>, String> = req
            .filters
            .iter()
            .map(|f| {
                let clause = filter_clause_pg("t", f, &mut param)?;
                if f.op != "isnotnull" {
                    binds.push(bind_string(f)?);
                }
                Ok(clause)
            })
            .collect();
        sql.push_str(" WHERE ");
        sql.push_str(
            &parts
                .map_err(|e| (StatusCode::BAD_REQUEST, e))?
                .join(" AND "),
        );
    }

    if !group_by.is_empty() {
        sql.push_str(" GROUP BY ");
        sql.push_str(&group_by.join(", "));
    }

    if let Some(orders) = &req.order_by {
        let mut ob = Vec::new();
        for o in orders {
            if o.column_index >= req.columns.len() {
                return Err((
                    StatusCode::BAD_REQUEST,
                    "order_by column_index out of range".into(),
                ));
            }
            let c = &req.columns[o.column_index];
            let agg = c.aggregation.as_deref().unwrap_or("none");
            let col_expr = if agg == "none" || agg.is_empty() {
                format!("t.{}", pg_ident(&c.column))
            } else {
                pg_ident(&c.alias.clone().unwrap_or_else(|| c.column.clone()))
            };
            let dir = if o.dir.eq_ignore_ascii_case("desc") {
                "DESC"
            } else {
                "ASC"
            };
            ob.push(format!("{} {}", col_expr, dir));
        }
        if !ob.is_empty() {
            sql.push_str(" ORDER BY ");
            sql.push_str(&ob.join(", "));
        }
    }

    sql.push_str(&format!(" LIMIT {limit}"));

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
        .map_err(|e| (StatusCode::BAD_REQUEST, format!("Connect: {e}")))?;

    let mut q = sqlx::query(&sql);
    for b in binds {
        q = q.bind(b);
    }
    let rows = q
        .fetch_all(&pool)
        .await
        .map_err(|e| (StatusCode::BAD_REQUEST, format!("Query: {e}")))?;
    pool.close().await;

    let mut out_rows = Vec::new();
    for row in &rows {
        out_rows.push(Value::Object(pg_row_map(row)));
    }

    let columns = if rows.is_empty() {
        out_names
    } else {
        rows[0]
            .columns()
            .iter()
            .map(|c| c.name().to_string())
            .collect()
    };

    Ok(Json(ReportExecuteResponse {
        row_count: out_rows.len(),
        columns,
        rows: out_rows,
    }))
}

fn build_select_mysql(
    cols: &[ReportColumnSpec],
    alias: &str,
) -> Result<(Vec<String>, Vec<String>, Vec<String>), String> {
    let mut select = Vec::new();
    let mut group_by = Vec::new();
    let mut out_names = Vec::new();
    for c in cols {
        let col_ref = format!("{}.{}", alias, mysql_ident(&c.column));
        let agg = c.aggregation.as_deref().unwrap_or("none");
        let alias_name = c.alias.clone().unwrap_or_else(|| c.column.clone());
        let (expr, gb) = match agg {
            "none" | "" => (col_ref.clone(), Some(col_ref)),
            "sum" => (
                format!("SUM({}) AS {}", col_ref, mysql_ident(&alias_name)),
                None,
            ),
            "avg" => (
                format!("AVG({}) AS {}", col_ref, mysql_ident(&alias_name)),
                None,
            ),
            "count" => (
                format!("COUNT({}) AS {}", col_ref, mysql_ident(&alias_name)),
                None,
            ),
            "min" => (
                format!("MIN({}) AS {}", col_ref, mysql_ident(&alias_name)),
                None,
            ),
            "max" => (
                format!("MAX({}) AS {}", col_ref, mysql_ident(&alias_name)),
                None,
            ),
            a => return Err(format!("Unsupported aggregation: {a}")),
        };
        select.push(expr);
        out_names.push(alias_name);
        if let Some(g) = gb {
            group_by.push(g);
        }
    }
    Ok((select, group_by, out_names))
}

fn filter_clause_mysql(alias: &str, f: &ReportFilterSpec) -> Result<String, String> {
    let col = format!("{}.{}", alias, mysql_ident(&f.column));
    Ok(match f.op.as_str() {
        "isnotnull" => format!("{col} IS NOT NULL"),
        "eq" => format!("{col} = ?"),
        "ne" => format!("{col} <> ?"),
        "gt" => format!("{col} > ?"),
        "gte" => format!("{col} >= ?"),
        "lt" => format!("{col} < ?"),
        "lte" => format!("{col} <= ?"),
        "contains" => format!("CAST({col} AS CHAR) LIKE ?"),
        "starts_with" => format!("CAST({col} AS CHAR) LIKE ?"),
        o => return Err(format!("Unsupported filter op: {o}")),
    })
}

fn mysql_row_map(row: &sqlx::mysql::MySqlRow) -> serde_json::Map<String, Value> {
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
        } else {
            json!(null)
        };
        map.insert(name, v);
    }
    map
}

async fn run_mysql(
    conn: &StoredConnection,
    password: &str,
    req: &ReportExecuteRequest,
    limit: u32,
) -> Result<Json<ReportExecuteResponse>, (StatusCode, String)> {
    let (select, group_by, out_names) =
        build_select_mysql(&req.columns, "t").map_err(|e| (StatusCode::BAD_REQUEST, e))?;

    let from = format!(
        "{}.{} AS t",
        mysql_ident(&req.columns[0].schema),
        mysql_ident(&req.columns[0].table)
    );

    let mut sql = format!("SELECT {} FROM {}", select.join(", "), from);
    let mut binds: Vec<String> = Vec::new();

    if !req.filters.is_empty() {
        let parts: Result<Vec<_>, String> = req
            .filters
            .iter()
            .map(|f| {
                let clause = filter_clause_mysql("t", f)?;
                if f.op != "isnotnull" {
                    binds.push(bind_string(f)?);
                }
                Ok(clause)
            })
            .collect();
        sql.push_str(" WHERE ");
        sql.push_str(
            &parts
                .map_err(|e| (StatusCode::BAD_REQUEST, e))?
                .join(" AND "),
        );
    }

    if !group_by.is_empty() {
        sql.push_str(" GROUP BY ");
        sql.push_str(&group_by.join(", "));
    }

    if let Some(orders) = &req.order_by {
        let mut ob = Vec::new();
        for o in orders {
            if o.column_index >= req.columns.len() {
                return Err((
                    StatusCode::BAD_REQUEST,
                    "order_by column_index out of range".into(),
                ));
            }
            let c = &req.columns[o.column_index];
            let agg = c.aggregation.as_deref().unwrap_or("none");
            let col_expr = if agg == "none" || agg.is_empty() {
                format!("t.{}", mysql_ident(&c.column))
            } else {
                mysql_ident(&c.alias.clone().unwrap_or_else(|| c.column.clone()))
            };
            let dir = if o.dir.eq_ignore_ascii_case("desc") {
                "DESC"
            } else {
                "ASC"
            };
            ob.push(format!("{} {}", col_expr, dir));
        }
        if !ob.is_empty() {
            sql.push_str(" ORDER BY ");
            sql.push_str(&ob.join(", "));
        }
    }

    sql.push_str(&format!(" LIMIT {limit}"));

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
        .map_err(|e| (StatusCode::BAD_REQUEST, format!("Connect: {e}")))?;

    let mut q = sqlx::query(&sql);
    for b in binds {
        q = q.bind(b);
    }
    let rows = q
        .fetch_all(&pool)
        .await
        .map_err(|e| (StatusCode::BAD_REQUEST, format!("Query: {e}")))?;
    pool.close().await;

    let mut out_rows = Vec::new();
    for row in &rows {
        out_rows.push(Value::Object(mysql_row_map(row)));
    }

    let columns = if rows.is_empty() {
        out_names
    } else {
        rows[0]
            .columns()
            .iter()
            .map(|c| c.name().to_string())
            .collect()
    };

    Ok(Json(ReportExecuteResponse {
        row_count: out_rows.len(),
        columns,
        rows: out_rows,
    }))
}

fn build_select_sqlite(
    cols: &[ReportColumnSpec],
    alias: &str,
) -> Result<(Vec<String>, Vec<String>, Vec<String>), String> {
    let q = |s: &str| format!("\"{}\"", s.replace('"', "\"\""));
    let mut select = Vec::new();
    let mut group_by = Vec::new();
    let mut out_names = Vec::new();
    for c in cols {
        let col_ref = format!("{}.{}", alias, q(&c.column));
        let agg = c.aggregation.as_deref().unwrap_or("none");
        let alias_name = c.alias.clone().unwrap_or_else(|| c.column.clone());
        let (expr, gb) = match agg {
            "none" | "" => (col_ref.clone(), Some(col_ref)),
            "sum" => (format!("SUM({}) AS {}", col_ref, q(&alias_name)), None),
            "avg" => (format!("AVG({}) AS {}", col_ref, q(&alias_name)), None),
            "count" => (format!("COUNT({}) AS {}", col_ref, q(&alias_name)), None),
            "min" => (format!("MIN({}) AS {}", col_ref, q(&alias_name)), None),
            "max" => (format!("MAX({}) AS {}", col_ref, q(&alias_name)), None),
            a => return Err(format!("Unsupported aggregation: {a}")),
        };
        select.push(expr);
        out_names.push(alias_name);
        if let Some(g) = gb {
            group_by.push(g);
        }
    }
    Ok((select, group_by, out_names))
}

fn filter_clause_sqlite(alias: &str, f: &ReportFilterSpec) -> String {
    let col = format!("{}.\"{}\"", alias, f.column.replace('"', "\"\""));
    format!("{col} = ?")
}

fn sqlite_row_map(row: &sqlx::sqlite::SqliteRow) -> serde_json::Map<String, Value> {
    let mut map = serde_json::Map::new();
    for (i, col) in row.columns().iter().enumerate() {
        let name = col.name().to_string();
        let v: Value = if let Ok(v) = row.try_get::<Option<String>, _>(i) {
            json!(v)
        } else if let Ok(v) = row.try_get::<Option<i64>, _>(i) {
            json!(v)
        } else if let Ok(v) = row.try_get::<Option<f64>, _>(i) {
            json!(v)
        } else {
            json!(null)
        };
        map.insert(name, v);
    }
    map
}

async fn run_sqlite(
    conn: &StoredConnection,
    req: &ReportExecuteRequest,
    limit: u32,
) -> Result<Json<ReportExecuteResponse>, (StatusCode, String)> {
    let (select, group_by, out_names) =
        build_select_sqlite(&req.columns, "t").map_err(|e| (StatusCode::BAD_REQUEST, e))?;

    let from = format!(
        "{}.{} AS t",
        format!("\"{}\"", req.columns[0].schema.replace('"', "\"\"")),
        format!("\"{}\"", req.columns[0].table.replace('"', "\"\""))
    );

    let mut sql = format!("SELECT {} FROM {}", select.join(", "), from);
    let mut binds: Vec<String> = Vec::new();

    if !req.filters.is_empty() {
        let mut parts = Vec::new();
        for f in &req.filters {
            if f.op != "eq" {
                return Err((
                    StatusCode::BAD_REQUEST,
                    "SQLite builder supports only eq filters for now".into(),
                ));
            }
            binds.push(bind_string(f).map_err(|e| (StatusCode::BAD_REQUEST, e))?);
            parts.push(filter_clause_sqlite("t", f));
        }
        sql.push_str(" WHERE ");
        sql.push_str(&parts.join(" AND "));
    }

    if !group_by.is_empty() {
        sql.push_str(" GROUP BY ");
        sql.push_str(&group_by.join(", "));
    }

    if let Some(orders) = &req.order_by {
        let mut ob = Vec::new();
        for o in orders {
            if o.column_index >= req.columns.len() {
                return Err((
                    StatusCode::BAD_REQUEST,
                    "order_by column_index out of range".into(),
                ));
            }
            let c = &req.columns[o.column_index];
            let agg = c.aggregation.as_deref().unwrap_or("none");
            let col_expr = if agg == "none" || agg.is_empty() {
                format!("t.\"{}\"", c.column.replace('"', "\"\""))
            } else {
                format!(
                    "\"{}\"",
                    c.alias
                        .clone()
                        .unwrap_or_else(|| c.column.clone())
                        .replace('"', "\"\"")
                )
            };
            let dir = if o.dir.eq_ignore_ascii_case("desc") {
                "DESC"
            } else {
                "ASC"
            };
            ob.push(format!("{} {}", col_expr, dir));
        }
        if !ob.is_empty() {
            sql.push_str(" ORDER BY ");
            sql.push_str(&ob.join(", "));
        }
    }

    sql.push_str(&format!(" LIMIT {limit}"));

    let db_path = format!(
        "sqlite:///{}",
        conn.database_name
            .trim()
            .trim_start_matches(|c| c == '/' || c == '\\')
    );
    let opts = SqliteConnectOptions::from_str(&db_path)
        .map_err(|e| (StatusCode::BAD_REQUEST, e.to_string()))?
        .create_if_missing(false);
    let pool = sqlx::sqlite::SqlitePoolOptions::new()
        .max_connections(1)
        .connect_with(opts)
        .await
        .map_err(|e| (StatusCode::BAD_REQUEST, format!("SQLite: {e}")))?;

    let mut q = sqlx::query(&sql);
    for b in binds {
        q = q.bind(b);
    }
    let rows = q
        .fetch_all(&pool)
        .await
        .map_err(|e| (StatusCode::BAD_REQUEST, format!("Query: {e}")))?;
    pool.close().await;

    let mut out_rows = Vec::new();
    for row in &rows {
        out_rows.push(Value::Object(sqlite_row_map(row)));
    }

    let columns = if rows.is_empty() {
        out_names
    } else {
        rows[0]
            .columns()
            .iter()
            .map(|c| c.name().to_string())
            .collect()
    };

    Ok(Json(ReportExecuteResponse {
        row_count: out_rows.len(),
        columns,
        rows: out_rows,
    }))
}
