//! AI-built data pages and ComposerX-style public publish URLs.

use axum::{
    Json,
    extract::{Path, Query, State},
    http::{StatusCode, header},
    response::{IntoResponse, Response},
};
use morphai::{Client, Config, Message};
use serde::{Deserialize, Serialize};
use serde_json::{Value, json};
use uuid::Uuid;

use crate::api::error_response;
use crate::db::data_page_build_repository::DataPageBuildRepository;
use crate::db::data_table_repository::DataTableRepository;
use crate::db::published_data_page_repository::PublishedDataPageRepository;
use crate::services::AppState;

const SCHEMA_SAMPLE_ROWS: usize = 3;
const INJECT_BATCH_SIZE: i32 = 500;
const MAX_INJECT_ROWS: usize = 5000;
const MAX_SCHEMA_JSON_CHARS: usize = 4_000;
const MAX_HTML_FOR_AI_CHARS: usize = 10_000;

const MARKER_SUMMARY_START: &str = "<!--DATAX_SUMMARY_START-->";
const MARKER_SUMMARY_END: &str = "<!--DATAX_SUMMARY_END-->";
const MARKER_TABLE_START: &str = "<!--DATAX_TABLE_START-->";
const MARKER_TABLE_END: &str = "<!--DATAX_TABLE_END-->";

const PAGE_BASE_STYLE_ID: &str = "datax-page-base";

/// Dim slate theme — between light and dark, leaning dark. Injected on every published page.
const PAGE_BASE_CSS: &str = r#"
:root {
  --datax-bg: #13171f;
  --datax-bg-elevated: #1a2030;
  --datax-bg-card: #222a3a;
  --datax-border: rgba(255, 255, 255, 0.08);
  --datax-border-strong: rgba(255, 255, 255, 0.14);
  --datax-text: #e9ecf1;
  --datax-text-muted: #a3adbe;
  --datax-text-dim: #7c8799;
  --datax-accent: #6ea8fe;
  --datax-accent-soft: rgba(110, 168, 254, 0.14);
  --datax-radius: 14px;
  --datax-radius-sm: 10px;
  --datax-shadow: 0 8px 32px rgba(0, 0, 0, 0.35);
  --datax-font: "Inter", system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
  --datax-max-width: 1120px;
  --datax-gutter: 28px;
  --datax-section-gap: 32px;
}

html, body {
  margin: 0;
  padding: 0;
  min-height: 100%;
  background: var(--datax-bg) !important;
  color: var(--datax-text) !important;
  font-family: var(--datax-font);
  font-size: 15px;
  line-height: 1.55;
  -webkit-font-smoothing: antialiased;
}

body {
  padding: var(--datax-gutter);
  box-sizing: border-box;
}

body * {
  box-sizing: border-box;
}

main, .datax-page-inner {
  max-width: var(--datax-max-width);
  margin: 0 auto;
}

header, .datax-hero {
  margin-bottom: var(--datax-section-gap);
  padding-bottom: 20px;
  border-bottom: 1px solid var(--datax-border);
}

h1, h2, h3, h4 {
  color: var(--datax-text) !important;
  letter-spacing: -0.02em;
  line-height: 1.25;
}

h1 { font-size: 1.75rem; font-weight: 700; margin: 0 0 8px; }
h2 { font-size: 1.125rem; font-weight: 600; margin: 0 0 12px; }

p, .datax-lead {
  color: var(--datax-text-muted) !important;
  margin: 0 0 12px;
  max-width: 52rem;
}

a { color: var(--datax-accent); }

.datax-section {
  margin-bottom: var(--datax-section-gap);
}

.datax-section-title {
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--datax-text-dim);
  margin: 0 0 14px;
}

.datax-summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(148px, 1fr));
  gap: 14px;
}

.datax-kpi {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 18px 20px;
  background: var(--datax-bg-card);
  border: 1px solid var(--datax-border);
  border-radius: var(--datax-radius-sm);
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.04);
}

.datax-kpi-label {
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--datax-text-dim);
}

.datax-kpi strong {
  font-size: 1.375rem;
  font-weight: 600;
  color: var(--datax-text);
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
}

.datax-table-wrap {
  overflow-x: auto;
  border: 1px solid var(--datax-border);
  border-radius: var(--datax-radius);
  background: var(--datax-bg-elevated);
  box-shadow: var(--datax-shadow);
}

.datax-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.datax-table thead th {
  position: sticky;
  top: 0;
  z-index: 1;
  padding: 14px 18px;
  text-align: left;
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--datax-text-muted);
  background: var(--datax-bg-card);
  border-bottom: 1px solid var(--datax-border-strong);
  white-space: nowrap;
}

.datax-table tbody td {
  padding: 13px 18px;
  border-bottom: 1px solid var(--datax-border);
  color: var(--datax-text);
  vertical-align: top;
}

.datax-table tbody tr:last-child td {
  border-bottom: none;
}

.datax-table tbody tr:hover td {
  background: rgba(255, 255, 255, 0.03);
}

.datax-table tbody tr:nth-child(even) td {
  background: rgba(255, 255, 255, 0.02);
}

.datax-table tbody tr:nth-child(even):hover td {
  background: rgba(255, 255, 255, 0.05);
}

html {
  scrollbar-width: thin;
  scrollbar-color: #3d4658 #1a2030;
}

::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #1a2030;
}

::-webkit-scrollbar-thumb {
  background: #3d4658;
  border-radius: 999px;
  border: 2px solid #1a2030;
}

::-webkit-scrollbar-thumb:hover {
  background: #556070;
}
"#;

#[derive(Debug, Deserialize)]
pub struct ResolvePathQuery {
    pub name: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct ListPublishesQuery {
    pub data_table_id: Option<String>,
    pub limit: Option<i32>,
    pub offset: Option<i32>,
}

#[derive(Debug, Deserialize)]
pub struct CreatePublishRequest {
    pub data_table_id: String,
    pub name: String,
    pub theme: Option<String>,
    pub html_content: String,
}

#[derive(Debug, Deserialize)]
pub struct PageChatMessage {
    pub role: String,
    pub content: String,
}

#[derive(Debug, Deserialize)]
pub struct PageChatRequest {
    pub messages: Vec<PageChatMessage>,
    pub current_html: Option<String>,
    pub theme: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct PublishSummary {
    pub id: String,
    pub data_table_id: String,
    pub name: String,
    pub slug: String,
    pub theme: String,
    pub public_path: String,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize)]
pub struct PublishDetail {
    pub id: String,
    pub data_table_id: String,
    pub name: String,
    pub slug: String,
    pub theme: String,
    pub html_content: String,
    pub public_path: String,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize)]
pub struct PageAIResponse {
    pub assistant_message: String,
    pub proposed_page_html: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct PageBuildSummary {
    pub id: String,
    pub data_table_id: String,
    pub label: String,
    pub source: String,
    pub created_at: String,
}

#[derive(Debug, Serialize)]
pub struct PageBuildDetail {
    pub id: String,
    pub data_table_id: String,
    pub label: String,
    pub source: String,
    pub html_content: String,
    pub created_at: String,
}

fn build_summary_from_row(
    row: crate::db::data_page_build_repository::DataPageBuildRow,
) -> PageBuildSummary {
    PageBuildSummary {
        id: row.id.to_string(),
        data_table_id: row.data_table_id.to_string(),
        label: row.label,
        source: row.source,
        created_at: row.created_at.to_rfc3339(),
    }
}

fn build_detail_from_row(
    row: crate::db::data_page_build_repository::DataPageBuildRow,
) -> PageBuildDetail {
    PageBuildDetail {
        id: row.id.to_string(),
        data_table_id: row.data_table_id.to_string(),
        label: row.label,
        source: row.source,
        html_content: row.html_content,
        created_at: row.created_at.to_rfc3339(),
    }
}

fn truncate_label(text: &str, max: usize) -> String {
    let t = text.trim().replace('\n', " ");
    if t.chars().count() <= max {
        return t;
    }
    format!("{}…", morphai::truncate_chars(&t, max.saturating_sub(1)))
}

async fn persist_page_build(
    state: &AppState,
    table_id: Uuid,
    html: &str,
    source: &str,
    label: &str,
) {
    let repo = DataPageBuildRepository::new(state.db_pool.clone());
    if let Err(e) = repo.save_and_trim(table_id, html, source, label).await {
        tracing::warn!("failed to save page build snapshot: {}", e);
    }
}

fn last_user_message(messages: &[PageChatMessage]) -> Option<String> {
    messages
        .iter()
        .rev()
        .find(|m| m.role.trim().eq_ignore_ascii_case("user"))
        .map(|m| m.content.trim())
        .filter(|s| !s.is_empty())
        .map(|s| truncate_label(s, 72))
}

fn user_email(headers: &axum::http::HeaderMap) -> Option<String> {
    headers
        .get("x-user-email")
        .and_then(|v| v.to_str().ok())
        .map(|s| s.to_string())
}

fn summary_from_row(
    row: crate::db::published_data_page_repository::PublishedDataPageRow,
) -> PublishSummary {
    PublishSummary {
        id: row.id.to_string(),
        data_table_id: row.data_table_id.to_string(),
        name: row.name,
        slug: row.slug.clone(),
        theme: row.theme,
        public_path: format!("/public/p/{}", row.slug),
        created_at: row.created_at.to_rfc3339(),
        updated_at: row.updated_at.to_rfc3339(),
    }
}

fn detail_from_row(
    row: crate::db::published_data_page_repository::PublishedDataPageRow,
) -> PublishDetail {
    PublishDetail {
        id: row.id.to_string(),
        data_table_id: row.data_table_id.to_string(),
        name: row.name,
        slug: row.slug.clone(),
        theme: row.theme,
        html_content: row.html_content,
        public_path: format!("/public/p/{}", row.slug),
        created_at: row.created_at.to_rfc3339(),
        updated_at: row.updated_at.to_rfc3339(),
    }
}

pub async fn resolve_path(
    State(state): State<AppState>,
    Query(q): Query<ResolvePathQuery>,
) -> Result<Json<Value>, (StatusCode, String)> {
    let name = q.name.unwrap_or_default();
    if name.trim().is_empty() {
        return Err((StatusCode::BAD_REQUEST, "name is required".into()));
    }
    let repo = PublishedDataPageRepository::new(state.db_pool.clone());
    let slug = repo
        .resolve_unique_slug(name.trim())
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
    Ok(Json(json!({
        "slug": slug,
        "public_path": format!("/public/p/{}", slug)
    })))
}

pub async fn list_publishes(
    State(state): State<AppState>,
    Query(q): Query<ListPublishesQuery>,
) -> Result<Json<Value>, (StatusCode, String)> {
    let table_id = q
        .data_table_id
        .as_deref()
        .map(Uuid::parse_str)
        .transpose()
        .map_err(|_| (StatusCode::BAD_REQUEST, "invalid data_table_id".into()))?
        .ok_or((StatusCode::BAD_REQUEST, "data_table_id is required".into()))?;

    let limit = q.limit.unwrap_or(50).clamp(1, 200);
    let offset = q.offset.unwrap_or(0).max(0);

    let repo = PublishedDataPageRepository::new(state.db_pool.clone());
    let rows = repo
        .list_by_table(table_id, limit, offset)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
    let total = repo
        .count_by_table(table_id)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let items: Vec<PublishSummary> = rows.into_iter().map(summary_from_row).collect();
    Ok(Json(json!({
        "items": items,
        "total": total,
        "limit": limit,
        "offset": offset
    })))
}

pub async fn list_publish_catalog(
    State(state): State<AppState>,
    Query(q): Query<ListPublishesQuery>,
) -> Result<Json<Value>, (StatusCode, String)> {
    let limit = q.limit.unwrap_or(100).clamp(1, 200);
    let offset = q.offset.unwrap_or(0).max(0);

    let repo = PublishedDataPageRepository::new(state.db_pool.clone());
    let rows = repo
        .list_recent_all(limit, offset)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
    let total = repo
        .count_all()
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let items: Vec<PublishSummary> = rows.into_iter().map(summary_from_row).collect();
    Ok(Json(json!({
        "items": items,
        "total": total,
        "limit": limit,
        "offset": offset
    })))
}

pub async fn create_publish(
    State(state): State<AppState>,
    headers: axum::http::HeaderMap,
    Json(body): Json<CreatePublishRequest>,
) -> Result<(StatusCode, Json<Value>), (StatusCode, String)> {
    let table_id = Uuid::parse_str(&body.data_table_id)
        .map_err(|_| (StatusCode::BAD_REQUEST, "invalid data_table_id".into()))?;

    let tables = DataTableRepository::new(state.db_pool.clone());
    if tables
        .find_by_id(table_id)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .is_none()
    {
        return Err((StatusCode::NOT_FOUND, "data table not found".into()));
    }

    let theme = body.theme.unwrap_or_else(|| "light".to_string());
    let created_by = user_email(&headers);

    let html_content = if body.html_content.contains(MARKER_TABLE_START)
        || body.html_content.contains(MARKER_SUMMARY_START)
    {
        finalize_page_with_data(&state, table_id, body.html_content.clone())
            .await
            .unwrap_or(body.html_content)
    } else {
        body.html_content
    };

    let repo = PublishedDataPageRepository::new(state.db_pool.clone());
    let row = repo
        .create(
            table_id,
            &body.name,
            &theme,
            &html_content,
            created_by.as_deref(),
        )
        .await
        .map_err(|e| {
            if e.to_string().contains("name and html required") {
                (StatusCode::BAD_REQUEST, "name and html required".into())
            } else {
                (StatusCode::INTERNAL_SERVER_ERROR, e.to_string())
            }
        })?;

    let summary = summary_from_row(row);
    Ok((StatusCode::CREATED, Json(json!(summary))))
}

pub async fn serve_published_page(
    State(state): State<AppState>,
    Path(slug): Path<String>,
) -> Result<Response, (StatusCode, String)> {
    let slug = slug.trim();
    if slug.is_empty() {
        return Err((StatusCode::BAD_REQUEST, "invalid slug".into()));
    }
    let repo = PublishedDataPageRepository::new(state.db_pool.clone());
    let row = repo
        .find_by_slug(slug)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .ok_or((StatusCode::NOT_FOUND, "published page not found".into()))?;

    let html = ensure_full_html_document(&row.html_content, &row.name);
    let html = polish_published_html(&html);
    Ok((
        StatusCode::OK,
        [(header::CONTENT_TYPE, "text/html; charset=utf-8")],
        html,
    )
        .into_response())
}

pub async fn get_published_page_json(
    State(state): State<AppState>,
    Path(slug): Path<String>,
) -> Result<Json<PublishDetail>, (StatusCode, String)> {
    let slug = slug.trim();
    if slug.is_empty() {
        return Err((StatusCode::BAD_REQUEST, "invalid slug".into()));
    }
    let repo = PublishedDataPageRepository::new(state.db_pool.clone());
    let row = repo
        .find_by_slug(slug)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .ok_or((StatusCode::NOT_FOUND, "published page not found".into()))?;

    Ok(Json(detail_from_row(row)))
}

pub async fn list_page_builds(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<Value>, (StatusCode, String)> {
    let table_id =
        Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "invalid table id".into()))?;

    let tables = DataTableRepository::new(state.db_pool.clone());
    if tables
        .find_by_id(table_id)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .is_none()
    {
        return Err((StatusCode::NOT_FOUND, "data table not found".into()));
    }

    let repo = DataPageBuildRepository::new(state.db_pool.clone());
    let rows = repo
        .list_recent(
            table_id,
            crate::db::data_page_build_repository::MAX_BUILDS_PER_TABLE,
        )
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let items: Vec<PageBuildSummary> = rows.into_iter().map(build_summary_from_row).collect();
    Ok(Json(json!({ "items": items })))
}

pub async fn get_page_build(
    State(state): State<AppState>,
    Path((id, build_id)): Path<(String, String)>,
) -> Result<Json<PageBuildDetail>, (StatusCode, String)> {
    let table_id =
        Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "invalid table id".into()))?;
    let build_id = Uuid::parse_str(&build_id)
        .map_err(|_| (StatusCode::BAD_REQUEST, "invalid build id".into()))?;

    let repo = DataPageBuildRepository::new(state.db_pool.clone());
    let row = repo
        .find_by_id(table_id, build_id)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .ok_or((StatusCode::NOT_FOUND, "build not found".into()))?;

    Ok(Json(build_detail_from_row(row)))
}

pub async fn build_page(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<PageAIResponse>, Response> {
    let table_id = match Uuid::parse_str(&id) {
        Ok(id) => id,
        Err(_) => {
            return Err(error_response(StatusCode::BAD_REQUEST, "invalid table id").into_response());
        }
    };
    let context = match load_schema_context(&state, table_id).await {
        Ok(c) => c,
        Err((status, msg)) => return Err(error_response(status, &msg).into_response()),
    };
    let prompt = format!(
        r#"Design a polished public HTML dashboard SHELL for this data table (layout + CSS only).

Table name: {name}
Row count: {row_count}
Schema for layout decisions (sample rows + column summaries — full data is injected server-side later):
{schema_json}

Reply with ONLY valid JSON (no markdown):
{{
  "assistant_message": "brief summary of the layout you designed",
  "proposed_page_html": "complete self-contained HTML document with embedded <style>"
}}

TEMPLATE RULES (required):
- Visual theme: dim slate (between light and dark, leaning dark). Background ~#13171f, cards ~#222a3a, text ~#e9ecf1. Generous spacing (28px page padding, 18px card padding). Max-width ~1120px centered.
- Style injected regions via classes: datax-summary-grid, datax-kpi, datax-kpi-label, datax-table-wrap, datax-table. Do not fight the dim palette with bright white backgrounds.
- Do NOT paste the full dataset or large tables of row values.
- Include these exact markers where injected content will appear:
  {summary_start} ... {summary_end}  (overview / KPI cards area)
  {table_start} ... {table_end}  (main data table area)
- You may add titles, sections, and CSS around the markers.
- Use column_summaries only to choose labels and chart-friendly layout (not exact numbers in charts).
- Prefer clean responsive layout; small inline script is OK only if needed for layout."#,
        name = context.name,
        row_count = context.row_count,
        schema_json = context.schema_json,
        summary_start = MARKER_SUMMARY_START,
        summary_end = MARKER_SUMMARY_END,
        table_start = MARKER_TABLE_START,
        table_end = MARKER_TABLE_END,
    );

    match run_page_ai(&prompt, None, None, &[]).await {
        Ok(mut res) => {
            if let Some(html) = res.proposed_page_html {
                let finalized = finalize_page_with_data(&state, table_id, html)
                    .await
                    .map_err(|(s, m)| error_response(s, &m).into_response())?;
                persist_page_build(&state, table_id, &finalized, "build", "Quick rebuild").await;
                res.proposed_page_html = Some(finalized);
            }
            Ok(Json(res))
        }
        Err((status, msg)) => Err(error_response(status, &msg).into_response()),
    }
}

pub async fn page_chat(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(body): Json<PageChatRequest>,
) -> Result<Json<PageAIResponse>, Response> {
    let table_id = match Uuid::parse_str(&id) {
        Ok(id) => id,
        Err(_) => {
            return Err(error_response(StatusCode::BAD_REQUEST, "invalid table id").into_response());
        }
    };
    let context = match load_schema_context(&state, table_id).await {
        Ok(c) => c,
        Err((status, msg)) => return Err(error_response(status, &msg).into_response()),
    };

    let user_msgs: Vec<Message> = body
        .messages
        .iter()
        .rev()
        .take(6)
        .rev()
        .filter_map(|m| {
            let role = m.role.trim().to_lowercase();
            let content = m.content.trim();
            if content.is_empty() {
                return None;
            }
            let content = if content.contains("<html") || content.len() > 2000 {
                morphai::truncate_chars(content, 800)
            } else {
                content.to_string()
            };
            let role = match role.as_str() {
                "assistant" => "assistant",
                _ => "user",
            };
            Some(Message {
                role: role.to_string(),
                content,
            })
        })
        .collect();

    if user_msgs.is_empty() {
        return Err(error_response(StatusCode::BAD_REQUEST, "messages required").into_response());
    }

    let schema_hint = format!(
        "Table schema for \"{}\" (full rows are injected automatically — edit layout/markers only):\n{}",
        context.name, context.schema_json
    );

    let current_for_ai = body
        .current_html
        .as_deref()
        .map(strip_injected_regions_for_ai);

    match run_page_ai(
        &schema_hint,
        current_for_ai.as_deref(),
        body.theme.as_deref(),
        &user_msgs,
    )
    .await
    {
        Ok(mut res) => {
            if let Some(html) = res.proposed_page_html {
                let finalized = finalize_page_with_data(&state, table_id, html)
                    .await
                    .map_err(|(s, m)| error_response(s, &m).into_response())?;
                let label =
                    last_user_message(&body.messages).unwrap_or_else(|| "AI edit".to_string());
                persist_page_build(&state, table_id, &finalized, "chat", &label).await;
                res.proposed_page_html = Some(finalized);
            }
            Ok(Json(res))
        }
        Err((status, msg)) => Err(error_response(status, &msg).into_response()),
    }
}

struct TableSchemaContext {
    name: String,
    row_count: i32,
    schema_json: String,
}

struct TablePageData {
    name: String,
    columns: Vec<String>,
    rows: Vec<Value>,
    summaries: Value,
}

async fn load_schema_context(
    state: &AppState,
    table_id: Uuid,
) -> Result<TableSchemaContext, (StatusCode, String)> {
    let repo = DataTableRepository::new(state.db_pool.clone());
    let table = repo
        .find_by_id(table_id)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .ok_or((StatusCode::NOT_FOUND, "data table not found".into()))?;

    let columns: Vec<String> = serde_json::from_str(&table.column_schema)
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let sample_batch = repo
        .fetch_rows(table_id, SCHEMA_SAMPLE_ROWS as i32, 0)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let mut sample_rows: Vec<Value> = Vec::new();
    for r in sample_batch {
        if let Ok(v) = serde_json::from_str(&r.data) {
            sample_rows.push(v);
        }
    }

    // Summaries from first batch only for schema (fast); full data injected separately.
    let summaries = summarize_columns(&columns, &sample_rows);

    let payload = json!({
        "columns": columns,
        "row_count": table.row_count,
        "sample_rows": sample_rows,
        "column_summaries": summaries,
    });
    let mut schema_json = payload.to_string();
    if schema_json.chars().count() > MAX_SCHEMA_JSON_CHARS {
        schema_json = morphai::truncate_chars(&schema_json, MAX_SCHEMA_JSON_CHARS);
    }

    Ok(TableSchemaContext {
        name: table.name,
        row_count: table.row_count,
        schema_json,
    })
}

async fn load_page_data(
    state: &AppState,
    table_id: Uuid,
) -> Result<TablePageData, (StatusCode, String)> {
    let repo = DataTableRepository::new(state.db_pool.clone());
    let table = repo
        .find_by_id(table_id)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .ok_or((StatusCode::NOT_FOUND, "data table not found".into()))?;

    let columns: Vec<String> = serde_json::from_str(&table.column_schema)
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let mut rows: Vec<Value> = Vec::new();
    let mut offset = 0;
    while rows.len() < MAX_INJECT_ROWS {
        let batch = repo
            .fetch_rows(table_id, INJECT_BATCH_SIZE, offset)
            .await
            .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
        if batch.is_empty() {
            break;
        }
        let n = batch.len();
        for r in batch {
            if rows.len() >= MAX_INJECT_ROWS {
                break;
            }
            if let Ok(v) = serde_json::from_str(&r.data) {
                rows.push(v);
            }
        }
        offset += INJECT_BATCH_SIZE;
        if n < INJECT_BATCH_SIZE as usize {
            break;
        }
    }

    let summaries = summarize_columns(&columns, &rows);

    Ok(TablePageData {
        name: table.name,
        columns,
        rows,
        summaries,
    })
}

async fn finalize_page_with_data(
    state: &AppState,
    table_id: Uuid,
    template: String,
) -> Result<String, (StatusCode, String)> {
    let data = load_page_data(state, table_id).await?;
    Ok(inject_data_into_template(&template, &data))
}

fn summarize_columns(columns: &[String], rows: &[Value]) -> Value {
    let mut out = serde_json::Map::new();
    for col in columns {
        let mut nums: Vec<f64> = Vec::new();
        let mut freq: std::collections::HashMap<String, u32> = std::collections::HashMap::new();
        let mut nulls = 0u32;
        for row in rows {
            let v = row.get(col);
            if v.is_none() || v.map(|x| x.is_null()).unwrap_or(true) {
                nulls += 1;
                continue;
            }
            if let Some(n) = v.and_then(|x| x.as_f64()) {
                nums.push(n);
            } else if let Some(n) = v.and_then(|x| x.as_i64()) {
                nums.push(n as f64);
            } else {
                let s = match v {
                    Some(Value::String(s)) => s.clone(),
                    Some(other) => other.to_string().trim_matches('"').to_string(),
                    None => continue,
                };
                if s.is_empty() {
                    nulls += 1;
                } else {
                    let key = s.clone();
                    freq.insert(key, freq.get(&s).unwrap_or(&0) + 1);
                }
            }
        }
        let mut col_summary = serde_json::Map::new();
        col_summary.insert("null_count".into(), json!(nulls));
        col_summary.insert(
            "non_null_count".into(),
            json!(rows.len().saturating_sub(nulls as usize)),
        );
        if !nums.is_empty() {
            let sum = nums.iter().sum::<f64>();
            let min = nums.iter().copied().fold(f64::INFINITY, f64::min);
            let max = nums.iter().copied().fold(f64::NEG_INFINITY, f64::max);
            col_summary.insert("type".into(), json!("numeric"));
            col_summary.insert("min".into(), json!(min));
            col_summary.insert("max".into(), json!(max));
            col_summary.insert("avg".into(), json!(sum / nums.len() as f64));
        } else {
            col_summary.insert("type".into(), json!("text"));
            let mut top: Vec<(String, u32)> = freq.into_iter().collect();
            top.sort_by(|a, b| b.1.cmp(&a.1));
            top.truncate(5);
            let top_json: Vec<Value> = top
                .into_iter()
                .map(|(value, count)| json!({ "value": value, "count": count }))
                .collect();
            col_summary.insert("top_values".into(), Value::Array(top_json));
        }
        out.insert(col.clone(), Value::Object(col_summary));
    }
    Value::Object(out)
}

fn inject_data_into_template(template: &str, data: &TablePageData) -> String {
    let summary_html = render_summary_html(data);
    let table_html = render_table_html(&data.columns, &data.rows);

    let mut html = inject_between_markers(
        template,
        MARKER_SUMMARY_START,
        MARKER_SUMMARY_END,
        &summary_html,
    );
    html = inject_between_markers(&html, MARKER_TABLE_START, MARKER_TABLE_END, &table_html);

    if !html.contains(MARKER_TABLE_START) {
        html = inject_before_body_end(&html, &table_html);
    }
    polish_published_html(&html)
}

fn strip_injected_regions_for_ai(html: &str) -> String {
    let mut out = clear_between_markers(html, MARKER_SUMMARY_START, MARKER_SUMMARY_END);
    out = clear_between_markers(&out, MARKER_TABLE_START, MARKER_TABLE_END);
    if out.chars().count() > MAX_HTML_FOR_AI_CHARS {
        morphai::truncate_chars(&out, MAX_HTML_FOR_AI_CHARS)
    } else {
        out
    }
}

fn clear_between_markers(html: &str, start: &str, end: &str) -> String {
    if let Some((start_idx, end_idx)) = find_marker_span(html, start, end) {
        let mut out = String::new();
        out.push_str(&html[..start_idx + start.len()]);
        out.push_str(&html[end_idx..]);
        out
    } else {
        html.to_string()
    }
}

fn inject_between_markers(html: &str, start: &str, end: &str, content: &str) -> String {
    if let Some((start_idx, end_idx)) = find_marker_span(html, start, end) {
        let mut out = String::new();
        out.push_str(&html[..start_idx + start.len()]);
        out.push_str(content);
        out.push_str(&html[end_idx..]);
        out
    } else {
        html.to_string()
    }
}

fn find_marker_span(html: &str, start: &str, end: &str) -> Option<(usize, usize)> {
    let start_idx = html.find(start)?;
    let end_idx = html.find(end)?;
    if end_idx > start_idx {
        Some((start_idx, end_idx))
    } else {
        None
    }
}

fn inject_before_body_end(html: &str, chunk: &str) -> String {
    let lower = html.to_lowercase();
    if let Some(idx) = lower.rfind("</body>") {
        let mut out = String::new();
        out.push_str(&html[..idx]);
        out.push_str(chunk);
        out.push_str(&html[idx..]);
        out
    } else {
        format!("{}{}", html, chunk)
    }
}

fn render_summary_html(data: &TablePageData) -> String {
    let mut cards = vec![
        format!(
            "<div class=\"datax-kpi\"><span class=\"datax-kpi-label\">Rows</span><strong>{}</strong></div>",
            data.rows.len()
        ),
        format!(
            "<div class=\"datax-kpi\"><span class=\"datax-kpi-label\">Columns</span><strong>{}</strong></div>",
            data.columns.len()
        ),
    ];

    for col in data.columns.iter().take(4) {
        if let Some(summary) = data.summaries.get(col) {
            if summary.get("type").and_then(|v| v.as_str()) == Some("numeric") {
                let avg = summary
                    .get("avg")
                    .and_then(|v| v.as_f64())
                    .map(|n| format!("{:.2}", n))
                    .unwrap_or_else(|| "—".into());
                cards.push(format!(
                    "<div class=\"datax-kpi\"><span class=\"datax-kpi-label\">{label} (avg)</span><strong>{avg}</strong></div>",
                    label = escape_html(col),
                    avg = escape_html(&avg)
                ));
            } else if let Some(top) = summary.get("top_values").and_then(|v| v.as_array()) {
                if let Some(first) = top.first() {
                    let val = first.get("value").and_then(|v| v.as_str()).unwrap_or("—");
                    cards.push(format!(
                        "<div class=\"datax-kpi\"><span class=\"datax-kpi-label\">{label} (top)</span><strong>{val}</strong></div>",
                        label = escape_html(col),
                        val = escape_html(val)
                    ));
                }
            }
        }
    }

    format!(
        "<section class=\"datax-section datax-section--summary\"><p class=\"datax-section-title\">Overview</p><div class=\"datax-summary-grid\">{}</div></section>",
        cards.join("")
    )
}

fn render_table_html(columns: &[String], rows: &[Value]) -> String {
    let mut head = String::from("<tr>");
    for c in columns {
        head.push_str(&format!("<th>{}</th>", escape_html(c)));
    }
    head.push_str("</tr>");

    let mut body = String::new();
    for row in rows {
        body.push_str("<tr>");
        for c in columns {
            let cell = cell_value_as_string(row.get(c));
            body.push_str(&format!("<td>{}</td>", escape_html(&cell)));
        }
        body.push_str("</tr>");
    }

    format!(
        "<section class=\"datax-section datax-section--table\"><h2 class=\"datax-section-title\">Data</h2><div class=\"datax-table-wrap\"><table class=\"datax-table\"><thead>{head}</thead><tbody>{body}</tbody></table></div></section>",
        head = head,
        body = body
    )
}

fn cell_value_as_string(v: Option<&Value>) -> String {
    match v {
        None | Some(Value::Null) => String::new(),
        Some(Value::String(s)) => s.clone(),
        Some(Value::Number(n)) => n.to_string(),
        Some(Value::Bool(b)) => b.to_string(),
        Some(other) => other.to_string(),
    }
}

fn escape_html(s: &str) -> String {
    s.replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('"', "&quot;")
}

async fn run_page_ai(
    data_context: &str,
    current_html: Option<&str>,
    theme: Option<&str>,
    prior_messages: &[Message],
) -> Result<PageAIResponse, (StatusCode, String)> {
    crate::config::load_env_files();
    let cfg = Config::from_env();
    if !cfg.configured() {
        return Err((
            StatusCode::SERVICE_UNAVAILABLE,
            "AI assistant is not configured (set MORPH_AI_API_KEY in SharpReport/.env)".into(),
        ));
    }
    let client = Client::new(cfg);

    let mut instructions = String::from(
        r#"You are an expert HTML page composer for data analytics dashboards.
CRITICAL: Output one JSON object only with exactly these keys:
- "assistant_message": concise plain-text guidance for the user.
- "proposed_page_html": complete HTML for a public webpage (include <style> in the markup), OR null if no page update is needed.

Goal:
- Design layout and styles only. Row data is injected server-side between marker pairs:
  <!--DATAX_SUMMARY_START--> ... <!--DATAX_SUMMARY_END-->
  <!--DATAX_TABLE_START--> ... <!--DATAX_TABLE_END-->
- Do not embed large datasets or hundreds of table rows in your HTML.
- Keep semantic structure (header/main/section) with comfortable padding and clear vertical rhythm.
- Default visual theme: dim slate (between light and dark, leaning dark) — soft charcoal backgrounds, muted borders, light gray text. Avoid pure white (#fff) page backgrounds.
- Prefer CSS-based presentation; avoid heavy JavaScript."#,
    );

    if let Some(html) = current_html {
        let h = html.trim();
        if !h.is_empty() {
            instructions.push_str(
                "\n\nCurrent draft HTML (injected data regions may be empty — preserve markers):\n",
            );
            instructions.push_str(h);
        }
    }
    if let Some(t) = theme {
        let t = t.trim();
        if !t.is_empty() {
            instructions.push_str("\n\nRequested page theme: ");
            instructions.push_str(t);
            instructions.push_str(" (dim slate, leaning dark).");
        }
    } else {
        instructions.push_str("\n\nPage theme: dim slate (between light and dark, leaning dark).");
    }
    instructions.push_str("\n\nSchema / layout context:\n");
    instructions.push_str(data_context);

    for m in prior_messages {
        let role = m.role.trim();
        let content = m.content.trim();
        if content.is_empty() {
            continue;
        }
        instructions.push_str("\n\n");
        instructions.push_str(role);
        instructions.push(':');
        instructions.push(' ');
        instructions.push_str(content);
    }

    // Single user message keeps trim/retry logic predictable for providers.
    let messages: Vec<Message> = vec![Message {
        role: "user".to_string(),
        content: instructions,
    }];

    let reply = client
        .chat_completion(&messages)
        .await
        .map_err(|e| (StatusCode::BAD_GATEWAY, e))?;

    parse_page_ai_response(&reply).map_err(|e| (StatusCode::BAD_GATEWAY, e))
}

fn parse_page_ai_response(raw: &str) -> Result<PageAIResponse, String> {
    for json_str in candidate_json_strings(raw) {
        if let Ok(parsed) = serde_json::from_str::<Value>(&json_str) {
            return Ok(page_ai_response_from_value(parsed, raw));
        }
    }

    Ok(PageAIResponse {
        assistant_message: raw.trim().to_string(),
        proposed_page_html: None,
    })
}

fn page_ai_response_from_value(parsed: Value, raw: &str) -> PageAIResponse {
    let assistant_message = parsed
        .get("assistant_message")
        .or_else(|| parsed.get("assistantMessage"))
        .and_then(|v| v.as_str())
        .unwrap_or("")
        .trim()
        .to_string();
    let proposed = parsed
        .get("proposed_page_html")
        .or_else(|| parsed.get("proposedPageHtml"))
        .and_then(|v| v.as_str())
        .map(|s| s.to_string())
        .filter(|s| !s.trim().is_empty());

    let msg = if assistant_message.is_empty() {
        if proposed.is_some() {
            "Page draft is ready. Apply it when you are happy.".to_string()
        } else {
            raw.trim().to_string()
        }
    } else {
        assistant_message
    };

    PageAIResponse {
        assistant_message: msg,
        proposed_page_html: proposed,
    }
}

fn candidate_json_strings(raw: &str) -> Vec<String> {
    let mut out: Vec<String> = Vec::new();
    let mut push = |s: &str| {
        let t = s.trim();
        if t.is_empty() {
            return;
        }
        if !out.iter().any(|existing| existing == t) {
            out.push(t.to_string());
        }
    };

    push(raw);
    if let Some(fenced) = strip_code_fence(raw) {
        push(&fenced);
    }
    if let Some(obj) = extract_json_object_string_aware(raw) {
        push(&obj);
    }
    if let Some(fenced) = strip_code_fence(raw) {
        if let Some(obj) = extract_json_object_string_aware(&fenced) {
            push(&obj);
        }
    }
    out
}

fn strip_code_fence(s: &str) -> Option<String> {
    let t = s.trim();
    if !t.starts_with("```") {
        return None;
    }
    let after_open = t[3..].trim_start();
    let after_lang = after_open
        .strip_prefix("json")
        .unwrap_or(after_open)
        .trim_start();
    let end = after_lang.rfind("```")?;
    Some(after_lang[..end].trim().to_string())
}

fn extract_json_object_string_aware(s: &str) -> Option<String> {
    let start = s.find('{')?;
    let slice = &s[start..];
    let mut depth = 0i32;
    let mut in_string = false;
    let mut escape = false;

    for (i, ch) in slice.char_indices() {
        if in_string {
            if escape {
                escape = false;
                continue;
            }
            if ch == '\\' {
                escape = true;
                continue;
            }
            if ch == '"' {
                in_string = false;
            }
            continue;
        }

        match ch {
            '"' => in_string = true,
            '{' => depth += 1,
            '}' => {
                depth -= 1;
                if depth == 0 {
                    return Some(slice[..=i].to_string());
                }
            }
            _ => {}
        }
    }
    None
}

#[cfg(test)]
mod page_ai_parse_tests {
    use super::*;

    #[test]
    fn parses_json_with_css_braces_in_html_string() {
        let raw = r#"```json
{
  "assistant_message": "Dashboard ready.",
  "proposed_page_html": "<style>.kpi { color: red; }</style><body>{content}</body>"
}
```"#;
        let parsed = parse_page_ai_response(raw).expect("parse");
        assert_eq!(parsed.assistant_message, "Dashboard ready.");
        assert!(
            parsed
                .proposed_page_html
                .as_deref()
                .unwrap_or("")
                .contains(".kpi { color: red; }")
        );
    }

    #[test]
    fn parses_plain_json_object() {
        let raw = r#"{"assistant_message":"Hi","proposed_page_html":"<html></html>"}"#;
        let parsed = parse_page_ai_response(raw).expect("parse");
        assert_eq!(parsed.assistant_message, "Hi");
        assert_eq!(parsed.proposed_page_html.as_deref(), Some("<html></html>"));
    }
}

fn extract_json_object(s: &str) -> Option<String> {
    extract_json_object_string_aware(s)
}

fn ensure_full_html_document(html: &str, title: &str) -> String {
    let trimmed = html.trim();
    let doc = if trimmed.is_empty() {
        "<!doctype html><html><head><meta charset=\"utf-8\"><title>Published page</title></head><body></body></html>".to_string()
    } else {
        let lower = trimmed.to_lowercase();
        if lower.contains("<html") && lower.contains("<body") {
            trimmed.to_string()
        } else {
            let safe_title = title.trim();
            let safe_title = if safe_title.is_empty() {
                "Published page"
            } else {
                safe_title
            };
            format!(
                "<!doctype html><html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>{}</title></head><body>{}</body></html>",
                safe_title, trimmed
            )
        }
    };
    polish_published_html(&doc)
}

fn polish_published_html(html: &str) -> String {
    if html.contains(PAGE_BASE_STYLE_ID) {
        return html.to_string();
    }
    let fonts = r#"<link rel="preconnect" href="https://fonts.googleapis.com"><link rel="preconnect" href="https://fonts.gstatic.com" crossorigin><link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">"#;
    let style = format!(
        "<style id=\"{id}\">{css}</style>",
        id = PAGE_BASE_STYLE_ID,
        css = PAGE_BASE_CSS
    );
    inject_into_head(html, &format!("{}{}", fonts, style))
}

fn inject_into_head(html: &str, fragment: &str) -> String {
    let lower = html.to_lowercase();
    if let Some(pos) = lower.find("<head") {
        if let Some(rel) = html[pos..].find('>') {
            let insert = pos + rel + 1;
            let mut out = String::new();
            out.push_str(&html[..insert]);
            out.push_str(fragment);
            out.push_str(&html[insert..]);
            return out;
        }
    }
    if let Some(pos) = lower.find("<html") {
        if let Some(rel) = html[pos..].find('>') {
            let insert = pos + rel + 1;
            let mut out = String::new();
            out.push_str(&html[..insert]);
            out.push_str("<head>");
            out.push_str(fragment);
            out.push_str("</head>");
            out.push_str(&html[insert..]);
            return out;
        }
    }
    format!("<head>{}</head>{}", fragment, html)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::db::published_data_page_repository::slugify_publish_name;

    #[test]
    fn slugify_basic() {
        assert_eq!(slugify_publish_name("Summer Launch"), "summer-launch");
        assert_eq!(slugify_publish_name(""), "page");
    }

    #[test]
    fn polish_injects_base_styles() {
        let html =
            "<!doctype html><html><head><title>T</title></head><body><main>hi</main></body></html>";
        let out = polish_published_html(html);
        assert!(out.contains(PAGE_BASE_STYLE_ID));
        assert!(out.contains("--datax-bg"));
    }

    #[test]
    fn inject_markers() {
        let template = format!(
            "<main>{summary_start}{summary_end}{table_start}{table_end}</main>",
            summary_start = MARKER_SUMMARY_START,
            summary_end = MARKER_SUMMARY_END,
            table_start = MARKER_TABLE_START,
            table_end = MARKER_TABLE_END,
        );
        let data = TablePageData {
            name: "T".into(),
            columns: vec!["a".into()],
            rows: vec![json!({"a": "1"})],
            summaries: json!({}),
        };
        let out = inject_data_into_template(&template, &data);
        assert!(out.contains("datax-table"));
        assert!(out.contains("datax-summary-grid"));
        assert!(out.contains(MARKER_TABLE_START));
        assert!(out.contains(">1<"));
    }

    #[test]
    fn escape_html_basic() {
        assert_eq!(escape_html("<b>"), "&lt;b&gt;");
    }
}
