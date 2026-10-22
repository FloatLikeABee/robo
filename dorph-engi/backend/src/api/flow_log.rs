//! Flow log — free-form income/expense notes (ported from Booki into Projects).

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    Json,
};
use serde::Deserialize;
use serde_json::{json, Map, Value};
use sqlx::Row;
use std::collections::HashMap;
use std::sync::Arc;

use crate::api::extract::{json_err, AuthUser};
use crate::services::AppState;

type ApiResult = Result<Json<Value>, (StatusCode, Json<Value>)>;

fn db_err(e: sqlx::Error) -> (StatusCode, Json<Value>) {
    (
        StatusCode::INTERNAL_SERVER_ERROR,
        Json(json!({"error": e.to_string()})),
    )
}

use crate::api::record_validation::validate_flow_entry;

fn entry_from_row(r: &sqlx::sqlite::SqliteRow) -> Value {
    let tags_raw: String = r.try_get("tags").unwrap_or_default();
    let tags: Value = if tags_raw.trim().is_empty() {
        json!([])
    } else {
        serde_json::from_str(&tags_raw).unwrap_or_else(|_| json!([]))
    };
    json!({
        "id": r.get::<i64, _>("id"),
        "entry_date": r.get::<String, _>("entry_date"),
        "direction": r.get::<String, _>("direction"),
        "amount": r.get::<f64, _>("amount"),
        "currency": r.get::<String, _>("currency"),
        "category": r.get::<String, _>("category"),
        "status": r.get::<String, _>("status"),
        "title": r.get::<String, _>("title"),
        "notes": r.get::<String, _>("notes"),
        "tags": tags,
        "created_at": r.get::<String, _>("created_at"),
        "updated_at": r.get::<String, _>("updated_at"),
    })
}

#[derive(Deserialize)]
pub struct ListQuery {
    pub from: Option<String>,
    pub to: Option<String>,
    pub direction: Option<String>,
    pub status: Option<String>,
    pub category: Option<String>,
    pub limit: Option<i64>,
}

pub async fn list_entries(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ListQuery>,
) -> ApiResult {
    let limit = q.limit.unwrap_or(200).clamp(1, 500);
    let rows = sqlx::query(
        r#"SELECT id, entry_date, direction, amount, currency, category, status, title, notes, tags, created_at, updated_at
           FROM flow_log_entries WHERE organization_id = ? ORDER BY entry_date DESC, id DESC LIMIT ?"#,
    )
    .bind(auth.org_id)
    .bind(limit)
    .fetch_all(&state.pool)
    .await
    .map_err(db_err)?;

    let from = q.from.as_deref().unwrap_or("").trim();
    let to = q.to.as_deref().unwrap_or("").trim();
    let direction = q.direction.as_deref().unwrap_or("").trim().to_ascii_lowercase();
    let status = q.status.as_deref().unwrap_or("").trim().to_ascii_lowercase();
    let category = q.category.as_deref().unwrap_or("").trim().to_ascii_lowercase();

    let entries: Vec<Value> = rows
        .iter()
        .map(entry_from_row)
        .filter(|e| {
            let date = e["entry_date"].as_str().unwrap_or("");
            if !from.is_empty() && date < from {
                return false;
            }
            if !to.is_empty() && date > to {
                return false;
            }
            if !direction.is_empty() && e["direction"].as_str().unwrap_or("") != direction {
                return false;
            }
            if !status.is_empty()
                && e["status"].as_str().unwrap_or("").to_ascii_lowercase() != status
            {
                return false;
            }
            if !category.is_empty()
                && e["category"].as_str().unwrap_or("").to_ascii_lowercase() != category
            {
                return false;
            }
            true
        })
        .collect();

    Ok(Json(json!({ "entries": entries })))
}

#[derive(Deserialize)]
pub struct EntryBody {
    pub entry_date: String,
    pub direction: String,
    pub amount: f64,
    #[serde(default)]
    pub currency: String,
    #[serde(default)]
    pub category: String,
    #[serde(default)]
    pub status: String,
    #[serde(default)]
    pub title: String,
    #[serde(default)]
    pub notes: String,
    #[serde(default)]
    pub tags: Vec<String>,
}

pub async fn create_entry(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<EntryBody>,
) -> ApiResult {
    let fields = validate_flow_entry(
        &body.entry_date,
        &body.direction,
        body.amount,
        &body.currency,
        &body.status,
        &body.title,
    )
    .map_err(|e| json_err(&e))?;
    let (entry_date, direction, currency, status, title) = (
        fields.entry_date,
        fields.direction,
        fields.currency,
        fields.status,
        fields.title,
    );
    let tags = serde_json::to_string(&body.tags).unwrap_or_else(|_| "[]".into());

    let res = sqlx::query(
        r#"INSERT INTO flow_log_entries
           (organization_id, entry_date, direction, amount, currency, category, status, title, notes, tags, created_by)
           VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"#,
    )
    .bind(auth.org_id)
    .bind(&entry_date)
    .bind(direction)
    .bind(fields.amount)
    .bind(&currency)
    .bind(body.category.trim())
    .bind(&status)
    .bind(&title)
    .bind(body.notes.trim())
    .bind(&tags)
    .bind(auth.user_id)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;

    let row = sqlx::query(
        r#"SELECT id, entry_date, direction, amount, currency, category, status, title, notes, tags, created_at, updated_at
           FROM flow_log_entries WHERE id = ? AND organization_id = ?"#,
    )
    .bind(res.last_insert_rowid())
    .bind(auth.org_id)
    .fetch_one(&state.pool)
    .await
    .map_err(db_err)?;

    Ok(Json(entry_from_row(&row)))
}

pub async fn delete_entry(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Path(id): Path<i64>,
) -> ApiResult {
    let res = sqlx::query("DELETE FROM flow_log_entries WHERE organization_id = ? AND id = ?")
        .bind(auth.org_id)
        .bind(id)
        .execute(&state.pool)
        .await
        .map_err(db_err)?;
    if res.rows_affected() == 0 {
        return Err((
            StatusCode::NOT_FOUND,
            Json(json!({"error": "entry not found"})),
        ));
    }
    Ok(Json(json!({"ok": true})))
}

#[derive(Deserialize)]
pub struct SummaryQuery {
    pub from: Option<String>,
    pub to: Option<String>,
}

pub async fn summary(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<SummaryQuery>,
) -> ApiResult {
    let from = q.from.as_deref().unwrap_or("1970-01-01").trim();
    let to = q.to.as_deref().unwrap_or("9999-12-31").trim();
    let rows = sqlx::query(
        r#"SELECT direction, amount, category, status FROM flow_log_entries
           WHERE organization_id = ? AND entry_date BETWEEN ? AND ?"#,
    )
    .bind(auth.org_id)
    .bind(from)
    .bind(to)
    .fetch_all(&state.pool)
    .await
    .map_err(db_err)?;

    let mut income = 0.0;
    let mut expense = 0.0;
    let mut by_category: HashMap<String, f64> = HashMap::new();
    let mut by_status: HashMap<String, f64> = HashMap::new();

    for r in &rows {
        let direction: String = r.get("direction");
        let amount: f64 = r.get("amount");
        let category: String = r.try_get("category").unwrap_or_default();
        let status: String = r.try_get("status").unwrap_or_default();
        let cat_key = if category.trim().is_empty() {
            "Uncategorized".to_string()
        } else {
            category
        };
        let signed = if direction == "expense" { -amount } else { amount };
        if direction == "income" {
            income += amount;
        } else {
            expense += amount;
        }
        *by_category.entry(cat_key).or_default() += signed;
        *by_status.entry(if status.is_empty() {
            "logged".into()
        } else {
            status
        })
        .or_default() += amount;
    }

    let by_category: Map<String, Value> = by_category
        .into_iter()
        .map(|(k, v)| (k, json!(v)))
        .collect();
    let by_status: Map<String, Value> = by_status
        .into_iter()
        .map(|(k, v)| (k, json!(v)))
        .collect();

    Ok(Json(json!({
        "from": from,
        "to": to,
        "income": income,
        "expense": expense,
        "net": income - expense,
        "entry_count": rows.len(),
        "by_category": by_category,
        "by_status": by_status,
    })))
}
