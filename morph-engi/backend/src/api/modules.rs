//! REST handlers for all Morph Engi domain modules.

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    Json,
};
use serde::Deserialize;
use serde_json::{json, Value};
use sqlx::{Row, SqlitePool};
use std::sync::Arc;

use crate::api::extract::{json_err, AuthUser};
use crate::api::record_validation::{
    project_code_or_generated, require_person_name, require_project_name, validate_site_log,
};
use crate::services::AppState;

type ApiResult = Result<Json<Value>, (StatusCode, Json<Value>)>;

fn db_err(e: sqlx::Error) -> (StatusCode, Json<Value>) {
    (
        StatusCode::INTERNAL_SERVER_ERROR,
        Json(json!({"error": e.to_string()})),
    )
}

async fn set_project_links(
    pool: &SqlitePool,
    table: &str,
    entity_col: &str,
    entity_id: i64,
    project_ids: &[i64],
) -> Result<(), sqlx::Error> {
    let del = format!("DELETE FROM {table} WHERE {entity_col} = ?");
    sqlx::query(&del).bind(entity_id).execute(pool).await?;
    for pid in project_ids {
        if *pid > 0 {
            let ins = format!("INSERT OR IGNORE INTO {table} (project_id, {entity_col}) VALUES (?, ?)");
            sqlx::query(&ins).bind(pid).bind(entity_id).execute(pool).await?;
        }
    }
    Ok(())
}

async fn project_ids_for(
    pool: &SqlitePool,
    table: &str,
    entity_col: &str,
    entity_id: i64,
) -> Result<Vec<i64>, sqlx::Error> {
    let q = format!("SELECT project_id FROM {table} WHERE {entity_col} = ? ORDER BY project_id");
    let rows = sqlx::query(&q).bind(entity_id).fetch_all(pool).await?;
    Ok(rows.iter().map(|r| r.get::<i64, _>("project_id")).collect())
}

#[derive(Deserialize)]
pub struct ProjectQuery {
    pub status: Option<String>,
}

#[derive(Deserialize)]
pub struct ProjectBody {
    pub code: String,
    pub name: String,
    #[serde(default)]
    pub client: String,
    #[serde(default)]
    pub location: String,
    #[serde(default = "default_status")]
    pub status: String,
    pub start_date: Option<String>,
    pub end_date: Option<String>,
    #[serde(default)]
    pub budget_total: f64,
    #[serde(default)]
    pub progress_pct: f64,
    #[serde(default)]
    pub description: String,
}

fn default_status() -> String {
    "planning".into()
}

pub async fn list_projects(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectQuery>,
) -> ApiResult {
    let rows = if let Some(status) = q.status.filter(|s| !s.is_empty()) {
        sqlx::query(
            "SELECT * FROM projects WHERE organization_id = ? AND status = ? ORDER BY updated_at DESC",
        )
        .bind(auth.org_id)
        .bind(status)
        .fetch_all(&state.pool)
        .await
    } else {
        sqlx::query(
            "SELECT * FROM projects WHERE organization_id = ? ORDER BY updated_at DESC",
        )
        .bind(auth.org_id)
        .fetch_all(&state.pool)
        .await
    }
    .map_err(db_err)?;

    Ok(Json(json!({"projects": rows_to_projects(&rows)})))
}

pub async fn create_project(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<ProjectBody>,
) -> ApiResult {
    let name = require_project_name(&body.name).map_err(|e| json_err(&e))?;
    let code = project_code_or_generated(&body.code, &name);
    let res = sqlx::query(
        "INSERT INTO projects (organization_id, code, name, client, location, status, start_date, end_date, budget_total, progress_pct, description) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(&code)
    .bind(&name)
    .bind(body.client)
    .bind(body.location)
    .bind(body.status)
    .bind(body.start_date)
    .bind(body.end_date)
    .bind(body.budget_total)
    .bind(body.progress_pct)
    .bind(body.description)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;

    fetch_project(&state.pool, auth.org_id, res.last_insert_rowid()).await
}

pub async fn update_project(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Path(id): Path<i64>,
    Json(body): Json<ProjectBody>,
) -> ApiResult {
    let n = sqlx::query(
        "UPDATE projects SET code=?, name=?, client=?, location=?, status=?, start_date=?, end_date=?, budget_total=?, progress_pct=?, description=?, updated_at=datetime('now') WHERE id=? AND organization_id=?",
    )
    .bind(body.code.trim())
    .bind(body.name.trim())
    .bind(body.client)
    .bind(body.location)
    .bind(body.status)
    .bind(body.start_date)
    .bind(body.end_date)
    .bind(body.budget_total)
    .bind(body.progress_pct)
    .bind(body.description)
    .bind(id)
    .bind(auth.org_id)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    if n.rows_affected() == 0 {
        return Err(json_err("project not found"));
    }
    fetch_project(&state.pool, auth.org_id, id).await
}

pub async fn delete_project(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Path(id): Path<i64>,
) -> ApiResult {
    sqlx::query("DELETE FROM projects WHERE id = ? AND organization_id = ?")
        .bind(id)
        .bind(auth.org_id)
        .execute(&state.pool)
        .await
        .map_err(db_err)?;
    Ok(Json(json!({"deleted": true})))
}

async fn fetch_project(pool: &SqlitePool, org_id: i64, id: i64) -> ApiResult {
    let row = sqlx::query("SELECT * FROM projects WHERE id = ? AND organization_id = ?")
        .bind(id)
        .bind(org_id)
        .fetch_optional(pool)
        .await
        .map_err(db_err)?;
    match row {
        Some(r) => Ok(Json(json!({"project": row_to_project(&r)}))),
        None => Err(json_err("project not found")),
    }
}

fn rows_to_projects(rows: &[sqlx::sqlite::SqliteRow]) -> Vec<Value> {
    rows.iter().map(row_to_project).collect()
}

fn row_to_project(r: &sqlx::sqlite::SqliteRow) -> Value {
    json!({
        "id": r.get::<i64, _>("id"),
        "code": r.get::<String, _>("code"),
        "name": r.get::<String, _>("name"),
        "client": r.get::<String, _>("client"),
        "location": r.get::<String, _>("location"),
        "status": r.get::<String, _>("status"),
        "start_date": r.try_get::<String, _>("start_date").ok(),
        "end_date": r.try_get::<String, _>("end_date").ok(),
        "budget_total": r.get::<f64, _>("budget_total"),
        "progress_pct": r.get::<f64, _>("progress_pct"),
        "description": r.get::<String, _>("description"),
        "markdown_content": r.try_get::<String, _>("markdown_content").unwrap_or_default(),
        "html_content": r.try_get::<String, _>("html_content").unwrap_or_default(),
        "source_summary": r.try_get::<String, _>("source_summary").unwrap_or_default(),
        "published_slug": r.try_get::<Option<String>, _>("published_slug").ok().flatten(),
        "published_path": r.try_get::<Option<String>, _>("published_path").ok().flatten(),
        "created_at": r.try_get::<String, _>("created_at").ok(),
        "updated_at": r.try_get::<String, _>("updated_at").ok(),
    })
}

// --- Tasks ---

#[derive(Deserialize)]
pub struct TaskBody {
    pub project_id: i64,
    pub title: String,
    #[serde(default)]
    pub assignee: String,
    #[serde(default = "default_open")]
    pub status: String,
    #[serde(default = "default_normal")]
    pub priority: String,
    pub due_date: Option<String>,
}

fn default_open() -> String {
    "open".into()
}
fn default_normal() -> String {
    "normal".into()
}

pub async fn list_tasks(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id {
        sqlx::query("SELECT * FROM project_tasks WHERE organization_id = ? AND project_id = ? ORDER BY id DESC")
            .bind(auth.org_id).bind(pid).fetch_all(&state.pool).await
    } else {
        sqlx::query("SELECT * FROM project_tasks WHERE organization_id = ? ORDER BY id DESC")
            .bind(auth.org_id).fetch_all(&state.pool).await
    }.map_err(db_err)?;
    Ok(Json(json!({"tasks": rows.iter().map(row_to_task).collect::<Vec<_>>() })))
}

pub async fn create_task(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<TaskBody>,
) -> ApiResult {
    let res = sqlx::query(
        "INSERT INTO project_tasks (organization_id, project_id, title, assignee, status, priority, due_date) VALUES (?,?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(body.project_id)
    .bind(body.title.trim())
    .bind(body.assignee)
    .bind(body.status)
    .bind(body.priority)
    .bind(body.due_date)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    Ok(Json(json!({"task": {"id": res.last_insert_rowid(), "title": body.title}})))
}

fn row_to_task(r: &sqlx::sqlite::SqliteRow) -> Value {
    json!({
        "id": r.get::<i64, _>("id"),
        "project_id": r.get::<i64, _>("project_id"),
        "title": r.get::<String, _>("title"),
        "assignee": r.get::<String, _>("assignee"),
        "status": r.get::<String, _>("status"),
        "priority": r.get::<String, _>("priority"),
        "due_date": r.try_get::<String, _>("due_date").ok(),
    })
}

// --- Site logs ---

#[derive(Deserialize)]
pub struct ProjectFilter {
    pub project_id: Option<i64>,
}

#[derive(Deserialize)]
pub struct SiteLogBody {
    pub project_id: i64,
    pub log_date: String,
    #[serde(default)]
    pub weather: String,
    #[serde(default)]
    pub crew_count: i64,
    pub summary: String,
    #[serde(default)]
    pub issues: String,
}

pub async fn list_site_logs(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id {
        sqlx::query("SELECT * FROM site_logs WHERE organization_id = ? AND project_id = ? ORDER BY log_date DESC")
            .bind(auth.org_id).bind(pid).fetch_all(&state.pool).await
    } else {
        sqlx::query("SELECT * FROM site_logs WHERE organization_id = ? ORDER BY log_date DESC")
            .bind(auth.org_id).fetch_all(&state.pool).await
    }.map_err(db_err)?;
    Ok(Json(json!({"site_logs": rows.iter().map(|r| json!({
        "id": r.get::<i64,_>("id"),
        "project_id": r.get::<i64,_>("project_id"),
        "log_date": r.get::<String,_>("log_date"),
        "weather": r.get::<String,_>("weather"),
        "crew_count": r.get::<i64,_>("crew_count"),
        "summary": r.get::<String,_>("summary"),
        "issues": r.get::<String,_>("issues"),
    })).collect::<Vec<_>>() })))
}

pub async fn create_site_log(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<SiteLogBody>,
) -> ApiResult {
    let fields = validate_site_log(&body.log_date, &body.summary).map_err(|e| json_err(&e))?;
    let res = sqlx::query(
        "INSERT INTO site_logs (organization_id, project_id, log_date, weather, crew_count, summary, issues) VALUES (?,?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(body.project_id)
    .bind(&fields.log_date)
    .bind(body.weather)
    .bind(body.crew_count)
    .bind(&fields.summary)
    .bind(body.issues)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    Ok(Json(json!({"id": res.last_insert_rowid()})))
}

// --- Materials ---

#[derive(Deserialize)]
pub struct MaterialBody {
    pub name: String,
    #[serde(default)]
    pub category: String,
    #[serde(default = "default_unit")]
    pub unit: String,
    #[serde(default)]
    pub unit_cost: f64,
    #[serde(default)]
    pub supplier: String,
    #[serde(default)]
    pub stock_qty: f64,
    #[serde(default)]
    pub reorder_level: f64,
    #[serde(default)]
    pub description: String,
    #[serde(default)]
    pub project_ids: Vec<i64>,
}

fn default_unit() -> String {
    "ea".into()
}

pub async fn list_materials(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id {
        sqlx::query(
            "SELECT m.* FROM materials m INNER JOIN project_materials pm ON pm.material_id = m.id WHERE m.organization_id = ? AND pm.project_id = ? ORDER BY m.name",
        )
        .bind(auth.org_id)
        .bind(pid)
        .fetch_all(&state.pool)
        .await
    } else {
        sqlx::query("SELECT * FROM materials WHERE organization_id = ? ORDER BY name")
            .bind(auth.org_id)
            .fetch_all(&state.pool)
            .await
    }
    .map_err(db_err)?;

    let mut out = Vec::new();
    for r in &rows {
        let id = r.get::<i64, _>("id");
        let pids = project_ids_for(&state.pool, "project_materials", "material_id", id)
            .await
            .unwrap_or_default();
        let mut item = row_to_material(r);
        if let Some(obj) = item.as_object_mut() {
            obj.insert("project_ids".into(), json!(pids));
        }
        out.push(item);
    }
    Ok(Json(json!({"materials": out})))
}

pub async fn create_material(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<MaterialBody>,
) -> ApiResult {
    let res = sqlx::query(
        "INSERT INTO materials (organization_id, name, category, unit, unit_cost, supplier, stock_qty, reorder_level, description) VALUES (?,?,?,?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(body.name.trim())
    .bind(body.category)
    .bind(body.unit)
    .bind(body.unit_cost)
    .bind(body.supplier)
    .bind(body.stock_qty)
    .bind(body.reorder_level)
    .bind(body.description)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    let id = res.last_insert_rowid();
    if !body.project_ids.is_empty() {
        set_project_links(
            &state.pool,
            "project_materials",
            "material_id",
            id,
            &body.project_ids,
        )
        .await
        .map_err(db_err)?;
    }
    Ok(Json(json!({"material": {"id": id, "name": body.name, "project_ids": body.project_ids}})))
}

pub async fn update_material(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Path(id): Path<i64>,
    Json(body): Json<MaterialBody>,
) -> ApiResult {
    sqlx::query(
        "UPDATE materials SET name=?, category=?, unit=?, unit_cost=?, supplier=?, stock_qty=?, reorder_level=?, description=? WHERE id=? AND organization_id=?",
    )
    .bind(body.name.trim())
    .bind(body.category)
    .bind(body.unit)
    .bind(body.unit_cost)
    .bind(body.supplier)
    .bind(body.stock_qty)
    .bind(body.reorder_level)
    .bind(body.description)
    .bind(id)
    .bind(auth.org_id)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    if !body.project_ids.is_empty() {
        set_project_links(
            &state.pool,
            "project_materials",
            "material_id",
            id,
            &body.project_ids,
        )
        .await
        .map_err(db_err)?;
    }
    Ok(Json(json!({"updated": true})))
}

fn row_to_material(r: &sqlx::sqlite::SqliteRow) -> Value {
    json!({
        "id": r.get::<i64, _>("id"),
        "name": r.get::<String, _>("name"),
        "category": r.get::<String, _>("category"),
        "unit": r.get::<String, _>("unit"),
        "unit_cost": r.get::<f64, _>("unit_cost"),
        "supplier": r.get::<String, _>("supplier"),
        "stock_qty": r.get::<f64, _>("stock_qty"),
        "reorder_level": r.get::<f64, _>("reorder_level"),
        "description": r.try_get::<String, _>("description").unwrap_or_default(),
    })
}

#[derive(Deserialize)]
pub struct MaterialUsageBody {
    pub project_id: i64,
    pub material_id: i64,
    pub qty: f64,
    pub used_at: String,
    #[serde(default)]
    pub notes: String,
}

pub async fn list_material_usages(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id {
        sqlx::query("SELECT * FROM material_usages WHERE organization_id = ? AND project_id = ? ORDER BY used_at DESC")
            .bind(auth.org_id).bind(pid).fetch_all(&state.pool).await
    } else {
        sqlx::query("SELECT * FROM material_usages WHERE organization_id = ? ORDER BY used_at DESC")
            .bind(auth.org_id).fetch_all(&state.pool).await
    }.map_err(db_err)?;
    Ok(Json(json!({"usages": rows.iter().map(|r| json!({
        "id": r.get::<i64,_>("id"),
        "project_id": r.get::<i64,_>("project_id"),
        "material_id": r.get::<i64,_>("material_id"),
        "qty": r.get::<f64,_>("qty"),
        "used_at": r.get::<String,_>("used_at"),
        "notes": r.get::<String,_>("notes"),
    })).collect::<Vec<_>>() })))
}

pub async fn create_material_usage(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<MaterialUsageBody>,
) -> ApiResult {
    let mut tx = state.pool.begin().await.map_err(db_err)?;
    sqlx::query(
        "INSERT INTO material_usages (organization_id, project_id, material_id, qty, used_at, notes) VALUES (?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(body.project_id)
    .bind(body.material_id)
    .bind(body.qty)
    .bind(body.used_at)
    .bind(body.notes)
    .execute(&mut *tx)
    .await
    .map_err(db_err)?;
    sqlx::query("UPDATE materials SET stock_qty = stock_qty - ? WHERE id = ? AND organization_id = ?")
        .bind(body.qty)
        .bind(body.material_id)
        .bind(auth.org_id)
        .execute(&mut *tx)
        .await
        .map_err(db_err)?;
    tx.commit().await.map_err(db_err)?;
    Ok(Json(json!({"recorded": true})))
}

// --- Resource files (URL or upload), many-to-many with projects ---

#[derive(Deserialize)]
pub struct ResourceFileBody {
    pub name: String,
    #[serde(default = "default_url_source")]
    pub source_type: String,
    #[serde(default)]
    pub file_url: String,
    #[serde(default)]
    pub file_name: String,
    #[serde(default)]
    pub description: String,
    #[serde(default)]
    pub project_ids: Vec<i64>,
}

fn default_url_source() -> String {
    "url".into()
}

pub async fn list_resource_files(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id {
        sqlx::query(
            "SELECT rf.* FROM resource_files rf INNER JOIN project_resource_files prf ON prf.resource_file_id = rf.id WHERE rf.organization_id = ? AND prf.project_id = ? ORDER BY rf.id DESC",
        )
        .bind(auth.org_id)
        .bind(pid)
        .fetch_all(&state.pool)
        .await
    } else {
        sqlx::query("SELECT * FROM resource_files WHERE organization_id = ? ORDER BY id DESC")
            .bind(auth.org_id)
            .fetch_all(&state.pool)
            .await
    }
    .map_err(db_err)?;

    let mut out = Vec::new();
    for r in &rows {
        let id = r.get::<i64, _>("id");
        let pids = project_ids_for(&state.pool, "project_resource_files", "resource_file_id", id)
            .await
            .unwrap_or_default();
        let mut item = row_to_resource_file(r);
        if let Some(obj) = item.as_object_mut() {
            obj.insert("project_ids".into(), json!(pids));
        }
        out.push(item);
    }
    Ok(Json(json!({"resource_files": out})))
}

pub async fn create_resource_file(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<ResourceFileBody>,
) -> ApiResult {
    if body.name.trim().is_empty() {
        return Err(json_err("name required"));
    }
    let res = sqlx::query(
        "INSERT INTO resource_files (organization_id, name, source_type, file_url, file_name, description) VALUES (?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(body.name.trim())
    .bind(body.source_type)
    .bind(body.file_url)
    .bind(body.file_name)
    .bind(body.description)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    let id = res.last_insert_rowid();
    if !body.project_ids.is_empty() {
        set_project_links(
            &state.pool,
            "project_resource_files",
            "resource_file_id",
            id,
            &body.project_ids,
        )
        .await
        .map_err(db_err)?;
    }
    Ok(Json(json!({"resource_file": {"id": id}})))
}

pub async fn update_resource_file(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Path(id): Path<i64>,
    Json(body): Json<ResourceFileBody>,
) -> ApiResult {
    sqlx::query(
        "UPDATE resource_files SET name=?, source_type=?, file_url=?, file_name=?, description=? WHERE id=? AND organization_id=?",
    )
    .bind(body.name.trim())
    .bind(body.source_type)
    .bind(body.file_url)
    .bind(body.file_name)
    .bind(body.description)
    .bind(id)
    .bind(auth.org_id)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    if !body.project_ids.is_empty() {
        set_project_links(
            &state.pool,
            "project_resource_files",
            "resource_file_id",
            id,
            &body.project_ids,
        )
        .await
        .map_err(db_err)?;
    }
    Ok(Json(json!({"updated": true})))
}

fn row_to_resource_file(r: &sqlx::sqlite::SqliteRow) -> Value {
    json!({
        "id": r.get::<i64, _>("id"),
        "name": r.get::<String, _>("name"),
        "source_type": r.get::<String, _>("source_type"),
        "file_url": r.get::<String, _>("file_url"),
        "file_name": r.get::<String, _>("file_name"),
        "description": r.get::<String, _>("description"),
    })
}

// --- Project finance (1:1 with project) ---

#[derive(Deserialize)]
pub struct FinanceBody {
    pub project_id: i64,
    #[serde(default)]
    pub summary: String,
    #[serde(default)]
    pub total_planned: f64,
    #[serde(default)]
    pub total_actual: f64,
    #[serde(default)]
    pub notes: String,
}

pub async fn get_project_finance(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let pid = q.project_id.ok_or(json_err("project_id required"))?;
    let row = sqlx::query(
        "SELECT * FROM project_finances WHERE organization_id = ? AND project_id = ?",
    )
    .bind(auth.org_id)
    .bind(pid)
    .fetch_optional(&state.pool)
    .await
    .map_err(db_err)?;
    match row {
        Some(r) => Ok(Json(json!({"finance": row_to_finance(&r)}))),
        None => Ok(Json(json!({"finance": null}))),
    }
}

pub async fn upsert_project_finance(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<FinanceBody>,
) -> ApiResult {
    if body.project_id <= 0 {
        return Err(json_err("project_id required"));
    }
    let existing = sqlx::query(
        "SELECT id FROM project_finances WHERE organization_id = ? AND project_id = ?",
    )
    .bind(auth.org_id)
    .bind(body.project_id)
    .fetch_optional(&state.pool)
    .await
    .map_err(db_err)?;

    if let Some(row) = existing {
        let id = row.get::<i64, _>("id");
        sqlx::query(
            "UPDATE project_finances SET summary=?, total_planned=?, total_actual=?, notes=?, updated_at=datetime('now') WHERE id=? AND organization_id=?",
        )
        .bind(body.summary)
        .bind(body.total_planned)
        .bind(body.total_actual)
        .bind(body.notes)
        .bind(id)
        .bind(auth.org_id)
        .execute(&state.pool)
        .await
        .map_err(db_err)?;
    } else {
        sqlx::query(
            "INSERT INTO project_finances (organization_id, project_id, summary, total_planned, total_actual, notes) VALUES (?,?,?,?,?,?)",
        )
        .bind(auth.org_id)
        .bind(body.project_id)
        .bind(body.summary)
        .bind(body.total_planned)
        .bind(body.total_actual)
        .bind(body.notes)
        .execute(&state.pool)
        .await
        .map_err(db_err)?;
    }
    get_project_finance(
        State(state),
        AuthUser(auth),
        Query(ProjectFilter {
            project_id: Some(body.project_id),
        }),
    )
    .await
}

fn row_to_finance(r: &sqlx::sqlite::SqliteRow) -> Value {
    json!({
        "id": r.get::<i64, _>("id"),
        "project_id": r.get::<i64, _>("project_id"),
        "summary": r.get::<String, _>("summary"),
        "total_planned": r.get::<f64, _>("total_planned"),
        "total_actual": r.get::<f64, _>("total_actual"),
        "notes": r.get::<String, _>("notes"),
    })
}

// --- Budget ---

#[derive(Deserialize)]
pub struct BudgetBody {
    pub project_id: i64,
    pub cost_code: String,
    pub description: String,
    #[serde(default = "default_material_cat")]
    pub category: String,
    #[serde(default)]
    pub planned_amount: f64,
    #[serde(default)]
    pub actual_amount: f64,
}

fn default_material_cat() -> String {
    "material".into()
}

pub async fn list_budget_lines(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id {
        sqlx::query("SELECT * FROM budget_lines WHERE organization_id = ? AND project_id = ? ORDER BY cost_code")
            .bind(auth.org_id).bind(pid).fetch_all(&state.pool).await
    } else {
        sqlx::query("SELECT * FROM budget_lines WHERE organization_id = ? ORDER BY cost_code")
            .bind(auth.org_id).fetch_all(&state.pool).await
    }.map_err(db_err)?;
    Ok(Json(json!({"budget_lines": rows.iter().map(row_to_budget).collect::<Vec<_>>() })))
}

pub async fn create_budget_line(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<BudgetBody>,
) -> ApiResult {
    let res = sqlx::query(
        "INSERT INTO budget_lines (organization_id, project_id, cost_code, description, category, planned_amount, actual_amount) VALUES (?,?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(body.project_id)
    .bind(body.cost_code.trim())
    .bind(body.description.trim())
    .bind(body.category)
    .bind(body.planned_amount)
    .bind(body.actual_amount)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    Ok(Json(json!({"id": res.last_insert_rowid()})))
}

pub async fn update_budget_line(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Path(id): Path<i64>,
    Json(body): Json<BudgetBody>,
) -> ApiResult {
    sqlx::query(
        "UPDATE budget_lines SET project_id=?, cost_code=?, description=?, category=?, planned_amount=?, actual_amount=? WHERE id=? AND organization_id=?",
    )
    .bind(body.project_id)
    .bind(body.cost_code.trim())
    .bind(body.description.trim())
    .bind(body.category)
    .bind(body.planned_amount)
    .bind(body.actual_amount)
    .bind(id)
    .bind(auth.org_id)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    Ok(Json(json!({"updated": true})))
}

fn row_to_budget(r: &sqlx::sqlite::SqliteRow) -> Value {
    json!({
        "id": r.get::<i64, _>("id"),
        "project_id": r.get::<i64, _>("project_id"),
        "cost_code": r.get::<String, _>("cost_code"),
        "description": r.get::<String, _>("description"),
        "category": r.get::<String, _>("category"),
        "planned_amount": r.get::<f64, _>("planned_amount"),
        "actual_amount": r.get::<f64, _>("actual_amount"),
    })
}

// --- Resources ---

#[derive(Deserialize)]
pub struct ResourceBody {
    pub name: String,
    #[serde(default = "default_equipment")]
    pub resource_type: String,
    #[serde(default)]
    pub cost_per_day: f64,
    #[serde(default = "default_available")]
    pub availability: String,
}

fn default_equipment() -> String {
    "equipment".into()
}
fn default_available() -> String {
    "available".into()
}

pub async fn list_resources(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
) -> ApiResult {
    let rows = sqlx::query("SELECT * FROM resources WHERE organization_id = ? ORDER BY name")
        .bind(auth.org_id)
        .fetch_all(&state.pool)
        .await
        .map_err(db_err)?;
    Ok(Json(json!({"resources": rows.iter().map(|r| json!({
        "id": r.get::<i64,_>("id"),
        "name": r.get::<String,_>("name"),
        "resource_type": r.get::<String,_>("resource_type"),
        "cost_per_day": r.get::<f64,_>("cost_per_day"),
        "availability": r.get::<String,_>("availability"),
    })).collect::<Vec<_>>() })))
}

pub async fn create_resource(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<ResourceBody>,
) -> ApiResult {
    let res = sqlx::query(
        "INSERT INTO resources (organization_id, name, resource_type, cost_per_day, availability) VALUES (?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(body.name.trim())
    .bind(body.resource_type)
    .bind(body.cost_per_day)
    .bind(body.availability)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    Ok(Json(json!({"id": res.last_insert_rowid()})))
}

#[derive(Deserialize)]
pub struct AllocationBody {
    pub project_id: i64,
    pub resource_id: i64,
    pub start_date: String,
    pub end_date: Option<String>,
    #[serde(default = "default_one")]
    pub qty: f64,
    #[serde(default)]
    pub notes: String,
}

fn default_one() -> f64 {
    1.0
}

pub async fn list_resource_allocations(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id {
        sqlx::query("SELECT * FROM resource_allocations WHERE organization_id = ? AND project_id = ? ORDER BY start_date DESC")
            .bind(auth.org_id).bind(pid).fetch_all(&state.pool).await
    } else {
        sqlx::query("SELECT * FROM resource_allocations WHERE organization_id = ? ORDER BY start_date DESC")
            .bind(auth.org_id).fetch_all(&state.pool).await
    }.map_err(db_err)?;
    Ok(Json(json!({"allocations": rows.iter().map(|r| json!({
        "id": r.get::<i64,_>("id"),
        "project_id": r.get::<i64,_>("project_id"),
        "resource_id": r.get::<i64,_>("resource_id"),
        "start_date": r.get::<String,_>("start_date"),
        "end_date": r.try_get::<String,_>("end_date").ok(),
        "qty": r.get::<f64,_>("qty"),
        "notes": r.get::<String,_>("notes"),
    })).collect::<Vec<_>>() })))
}

pub async fn create_resource_allocation(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<AllocationBody>,
) -> ApiResult {
    let res = sqlx::query(
        "INSERT INTO resource_allocations (organization_id, project_id, resource_id, start_date, end_date, qty, notes) VALUES (?,?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(body.project_id)
    .bind(body.resource_id)
    .bind(body.start_date)
    .bind(body.end_date)
    .bind(body.qty)
    .bind(body.notes)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    Ok(Json(json!({"id": res.last_insert_rowid()})))
}

// --- Contractors ---

#[derive(Deserialize)]
pub struct ContractorBody {
    pub name: String,
    #[serde(default)]
    pub trade: String,
    #[serde(default)]
    pub contact_name: String,
    #[serde(default)]
    pub phone: String,
    #[serde(default)]
    pub email: String,
    #[serde(default)]
    pub rating: f64,
    #[serde(default = "default_active")]
    pub status: String,
    #[serde(default)]
    pub description: String,
    #[serde(default)]
    pub project_ids: Vec<i64>,
}

fn default_active() -> String {
    "active".into()
}

pub async fn list_contractors(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id {
        sqlx::query(
            "SELECT c.* FROM contractors c INNER JOIN project_contractors pc ON pc.contractor_id = c.id WHERE c.organization_id = ? AND pc.project_id = ? ORDER BY c.name",
        )
        .bind(auth.org_id)
        .bind(pid)
        .fetch_all(&state.pool)
        .await
    } else {
        sqlx::query("SELECT * FROM contractors WHERE organization_id = ? ORDER BY name")
            .bind(auth.org_id)
            .fetch_all(&state.pool)
            .await
    }
    .map_err(db_err)?;

    let mut out = Vec::new();
    for r in &rows {
        let id = r.get::<i64, _>("id");
        let pids = project_ids_for(&state.pool, "project_contractors", "contractor_id", id)
            .await
            .unwrap_or_default();
        let mut item = row_to_contractor(r);
        if let Some(obj) = item.as_object_mut() {
            obj.insert("project_ids".into(), json!(pids));
        }
        out.push(item);
    }
    Ok(Json(json!({"contractors": out})))
}

pub async fn create_contractor(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<ContractorBody>,
) -> ApiResult {
    let name = require_person_name(&body.name).map_err(|e| json_err(&e))?;
    let res = sqlx::query(
        "INSERT INTO contractors (organization_id, name, trade, contact_name, phone, email, rating, status, description) VALUES (?,?,?,?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(&name)
    .bind(body.trade)
    .bind(body.contact_name)
    .bind(body.phone)
    .bind(body.email)
    .bind(body.rating)
    .bind(body.status)
    .bind(body.description)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    let id = res.last_insert_rowid();
    if !body.project_ids.is_empty() {
        set_project_links(
            &state.pool,
            "project_contractors",
            "contractor_id",
            id,
            &body.project_ids,
        )
        .await
        .map_err(db_err)?;
    }
    Ok(Json(json!({"id": id})))
}

fn row_to_contractor(r: &sqlx::sqlite::SqliteRow) -> Value {
    json!({
        "id": r.get::<i64, _>("id"),
        "name": r.get::<String, _>("name"),
        "trade": r.get::<String, _>("trade"),
        "contact_name": r.get::<String, _>("contact_name"),
        "phone": r.get::<String, _>("phone"),
        "email": r.get::<String, _>("email"),
        "rating": r.get::<f64, _>("rating"),
        "status": r.get::<String, _>("status"),
        "description": r.try_get::<String, _>("description").unwrap_or_default(),
    })
}

fn default_draft() -> String {
    "draft".into()
}

#[derive(Deserialize)]
pub struct ContractBody {
    pub project_id: i64,
    pub contractor_id: i64,
    pub scope: String,
    #[serde(default)]
    pub contract_value: f64,
    #[serde(default = "default_draft")]
    pub status: String,
    pub start_date: Option<String>,
    pub end_date: Option<String>,
}

pub async fn list_contracts(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id {
        sqlx::query("SELECT * FROM contractor_contracts WHERE organization_id = ? AND project_id = ? ORDER BY id DESC")
            .bind(auth.org_id).bind(pid).fetch_all(&state.pool).await
    } else {
        sqlx::query("SELECT * FROM contractor_contracts WHERE organization_id = ? ORDER BY id DESC")
            .bind(auth.org_id).fetch_all(&state.pool).await
    }.map_err(db_err)?;
    Ok(Json(json!({"contracts": rows.iter().map(|r| json!({
        "id": r.get::<i64,_>("id"),
        "project_id": r.get::<i64,_>("project_id"),
        "contractor_id": r.get::<i64,_>("contractor_id"),
        "scope": r.get::<String,_>("scope"),
        "contract_value": r.get::<f64,_>("contract_value"),
        "status": r.get::<String,_>("status"),
        "start_date": r.try_get::<String,_>("start_date").ok(),
        "end_date": r.try_get::<String,_>("end_date").ok(),
    })).collect::<Vec<_>>() })))
}

pub async fn create_contract(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<ContractBody>,
) -> ApiResult {
    let res = sqlx::query(
        "INSERT INTO contractor_contracts (organization_id, project_id, contractor_id, scope, contract_value, status, start_date, end_date) VALUES (?,?,?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(body.project_id)
    .bind(body.contractor_id)
    .bind(body.scope.trim())
    .bind(body.contract_value)
    .bind(body.status)
    .bind(body.start_date)
    .bind(body.end_date)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    Ok(Json(json!({"id": res.last_insert_rowid()})))
}

// --- Public relations (per project) ---

#[derive(Deserialize)]
pub struct RelationBody {
    pub project_id: i64,
    pub name: String,
    #[serde(default)]
    pub role: String,
    #[serde(default)]
    pub contact: String,
    #[serde(default = "default_medium")]
    pub influence: String,
    #[serde(default = "default_neutral")]
    pub sentiment: String,
    #[serde(default)]
    pub description: String,
}

pub async fn list_relations(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id {
        sqlx::query(
            "SELECT * FROM public_relations WHERE organization_id = ? AND project_id = ? ORDER BY name",
        )
        .bind(auth.org_id)
        .bind(pid)
        .fetch_all(&state.pool)
        .await
    } else {
        sqlx::query("SELECT * FROM public_relations WHERE organization_id = ? ORDER BY name")
            .bind(auth.org_id)
            .fetch_all(&state.pool)
            .await
    }
    .map_err(db_err)?;
    Ok(Json(
        json!({"relations": rows.iter().map(row_to_relation).collect::<Vec<_>>() }),
    ))
}

pub async fn create_relation(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<RelationBody>,
) -> ApiResult {
    if body.project_id <= 0 {
        return Err(json_err("project_id required"));
    }
    let res = sqlx::query(
        "INSERT INTO public_relations (organization_id, project_id, name, role, contact, influence, sentiment, description) VALUES (?,?,?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(body.project_id)
    .bind(body.name.trim())
    .bind(body.role)
    .bind(body.contact)
    .bind(body.influence)
    .bind(body.sentiment)
    .bind(body.description)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    Ok(Json(json!({"id": res.last_insert_rowid()})))
}

fn row_to_relation(r: &sqlx::sqlite::SqliteRow) -> Value {
    json!({
        "id": r.get::<i64, _>("id"),
        "project_id": r.get::<i64, _>("project_id"),
        "name": r.get::<String, _>("name"),
        "role": r.get::<String, _>("role"),
        "contact": r.get::<String, _>("contact"),
        "influence": r.get::<String, _>("influence"),
        "sentiment": r.get::<String, _>("sentiment"),
        "description": r.get::<String, _>("description"),
    })
}

fn default_medium() -> String {
    "medium".into()
}
fn default_neutral() -> String {
    "neutral".into()
}

pub async fn list_communications(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<ProjectFilter>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id {
        sqlx::query("SELECT * FROM communications WHERE organization_id = ? AND project_id = ? ORDER BY occurred_at DESC")
            .bind(auth.org_id).bind(pid).fetch_all(&state.pool).await
    } else {
        sqlx::query("SELECT * FROM communications WHERE organization_id = ? ORDER BY occurred_at DESC")
            .bind(auth.org_id).fetch_all(&state.pool).await
    }.map_err(db_err)?;
    Ok(Json(json!({"communications": rows.iter().map(|r| json!({
        "id": r.get::<i64,_>("id"),
        "project_id": r.try_get::<i64,_>("project_id").ok(),
        "relation_id": r.try_get::<i64,_>("relation_id").ok(),
        "channel": r.get::<String,_>("channel"),
        "subject": r.get::<String,_>("subject"),
        "body": r.get::<String,_>("body"),
        "occurred_at": r.get::<String,_>("occurred_at"),
    })).collect::<Vec<_>>() })))
}

#[derive(Deserialize)]
pub struct CommunicationBody {
    pub project_id: Option<i64>,
    pub relation_id: Option<i64>,
    #[serde(default = "default_meeting")]
    pub channel: String,
    pub subject: String,
    #[serde(default)]
    pub body: String,
    pub occurred_at: String,
}

fn default_meeting() -> String {
    "meeting".into()
}

pub async fn create_communication(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(body): Json<CommunicationBody>,
) -> ApiResult {
    let res = sqlx::query(
        "INSERT INTO communications (organization_id, project_id, relation_id, channel, subject, body, occurred_at) VALUES (?,?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(body.project_id)
    .bind(body.relation_id)
    .bind(body.channel)
    .bind(body.subject.trim())
    .bind(body.body)
    .bind(body.occurred_at)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    Ok(Json(json!({"id": res.last_insert_rowid()})))
}

// --- File upload for resource files ---

use axum::body::Bytes;
use std::path::PathBuf;

pub async fn upload_resource_file(
    State(_state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    mut multipart: axum::extract::Multipart,
) -> ApiResult {
    let mut file_name = String::new();
    let mut file_bytes: Option<Bytes> = None;

    while let Some(field) = multipart.next_field().await.map_err(|e| {
        (
            StatusCode::BAD_REQUEST,
            Json(json!({"error": e.to_string()})),
        )
    })? {
        if field.name() == Some("file") {
            file_name = field.file_name().unwrap_or("upload").to_string();
            file_bytes = Some(field.bytes().await.map_err(|e| {
                (
                    StatusCode::BAD_REQUEST,
                    Json(json!({"error": e.to_string()})),
                )
            })?);
        }
    }

    let bytes = file_bytes.ok_or(json_err("file required"))?;
    let upload_dir = PathBuf::from("uploads").join(auth.org_id.to_string());
    tokio::fs::create_dir_all(&upload_dir).await.map_err(|e| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": e.to_string()})),
        )
    })?;

    let safe_name = file_name
        .chars()
        .map(|c| if c.is_ascii_alphanumeric() || c == '.' || c == '-' || c == '_' { c } else { '_' })
        .collect::<String>();
    let stored = format!("{}_{}", chrono::Utc::now().timestamp_millis(), safe_name);
    let path = upload_dir.join(&stored);
    tokio::fs::write(&path, &bytes).await.map_err(|e| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": e.to_string()})),
        )
    })?;

    let url = format!("/api/v1/uploads/{}/{}", auth.org_id, stored);
    Ok(Json(json!({
        "file_url": url,
        "file_name": file_name,
        "source_type": "upload"
    })))
}

pub async fn serve_upload(
    Path((org_id, filename)): Path<(i64, String)>,
) -> Result<axum::response::Response, (StatusCode, Json<Value>)> {
    let path = PathBuf::from("uploads").join(org_id.to_string()).join(&filename);
    if !path.exists() {
        return Err((
            StatusCode::NOT_FOUND,
            Json(json!({"error": "not found"})),
        ));
    }
    let data = tokio::fs::read(&path).await.map_err(|e| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": e.to_string()})),
        )
    })?;
    Ok(axum::response::Response::builder()
        .header(axum::http::header::CONTENT_TYPE, "application/octet-stream")
        .body(axum::body::Body::from(data))
        .unwrap())
}
