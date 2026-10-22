use axum::{extract::State, http::StatusCode, Json};
use sqlx::Row;
use std::sync::Arc;

use crate::api::extract::AuthUser;
use crate::services::AppState;

pub async fn dashboard(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
) -> Result<Json<serde_json::Value>, (StatusCode, Json<serde_json::Value>)> {
    let org = auth.org_id;
    let projects: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM projects WHERE organization_id = ?",
    )
    .bind(org)
    .fetch_one(&state.pool)
    .await
    .unwrap_or(0);

    let active_projects: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM projects WHERE organization_id = ? AND status = 'active'",
    )
    .bind(org)
    .fetch_one(&state.pool)
    .await
    .unwrap_or(0);

    let budget_planned: f64 = sqlx::query_scalar(
        "SELECT COALESCE(SUM(budget_total), 0) FROM projects WHERE organization_id = ?",
    )
    .bind(org)
    .fetch_one(&state.pool)
    .await
    .unwrap_or(0.0);

    let budget_actual: f64 = sqlx::query_scalar(
        "SELECT COALESCE(SUM(actual_amount), 0) FROM budget_lines WHERE organization_id = ?",
    )
    .bind(org)
    .fetch_one(&state.pool)
    .await
    .unwrap_or(0.0);

    let materials_low: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM materials WHERE organization_id = ? AND stock_qty <= reorder_level AND reorder_level > 0",
    )
    .bind(org)
    .fetch_one(&state.pool)
    .await
    .unwrap_or(0);

    let open_tasks: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM project_tasks WHERE organization_id = ? AND status != 'done'",
    )
    .bind(org)
    .fetch_one(&state.pool)
    .await
    .unwrap_or(0);

    let recent = sqlx::query(
        "SELECT id, code, name, status, progress_pct, budget_total FROM projects WHERE organization_id = ? ORDER BY updated_at DESC LIMIT 5",
    )
    .bind(org)
    .fetch_all(&state.pool)
    .await
    .unwrap_or_default();

    let recent_projects: Vec<serde_json::Value> = recent
        .iter()
        .map(|r| {
            serde_json::json!({
                "id": r.get::<i64, _>("id"),
                "code": r.get::<String, _>("code"),
                "name": r.get::<String, _>("name"),
                "status": r.get::<String, _>("status"),
                "progress_pct": r.get::<f64, _>("progress_pct"),
                "budget_total": r.get::<f64, _>("budget_total"),
            })
        })
        .collect();

    Ok(Json(serde_json::json!({
        "projects_total": projects,
        "projects_active": active_projects,
        "budget_planned": budget_planned,
        "budget_actual": budget_actual,
        "materials_low_stock": materials_low,
        "open_tasks": open_tasks,
        "recent_projects": recent_projects,
    })))
}
