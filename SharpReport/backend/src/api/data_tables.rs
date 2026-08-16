use axum::{
    Json,
    extract::{Path, Query, State},
    http::{HeaderMap, StatusCode},
};
use morphai::{Client, Config, Message};
use serde::{Deserialize, Serialize};
use serde_json::{Value, json};
use uuid::Uuid;

use crate::api::data_table_query::{ParsedRow, RowQueryOptions, aggregate_rows, apply_row_query};
use crate::db::data_table_repository::{DataTableRepository, DataTableRow};
use crate::services::AppState;

const MAX_IMPORT_ROWS: usize = 5000;
const MAX_ANALYZE_CHARS: usize = 48_000;

#[derive(Debug, Serialize)]
pub struct DataTableSummary {
    pub id: String,
    pub name: String,
    pub source_filename: Option<String>,
    pub source_format: String,
    pub columns: Vec<String>,
    pub row_count: i32,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize)]
pub struct DataTableRowsResponse {
    pub columns: Vec<String>,
    pub rows: Vec<Value>,
    pub total: i32,
    pub limit: i32,
    pub offset: i32,
}

#[derive(Debug, Deserialize)]
pub struct AnalyzeRequest {
    pub source_filename: Option<String>,
    pub source_format: String,
    pub content_text: String,
}

#[derive(Debug, Serialize)]
pub struct AnalyzeResponse {
    pub can_be_table: bool,
    pub reason: String,
    pub columns: Vec<String>,
    pub rows: Vec<Value>,
}

#[derive(Debug, Deserialize)]
pub struct ImportRequest {
    pub name: Option<String>,
    pub source_filename: Option<String>,
    pub source_format: String,
    pub columns: Vec<String>,
    pub rows: Vec<Value>,
    /// When true, also push a CSV excerpt to Morph Knowledge Library + Neo4j GraphRAG.
    #[serde(default)]
    pub index_to_graph: bool,
}

#[derive(Debug, Serialize)]
pub struct ImportResponse {
    #[serde(flatten)]
    pub table: DataTableSummary,
    pub index_to_graph: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub graph_status: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct UpdateRowRequest {
    pub data: Value,
}

#[derive(Debug, Deserialize)]
pub struct RowsQuery {
    pub limit: Option<i32>,
    pub offset: Option<i32>,
    pub search: Option<String>,
    pub sort_by: Option<String>,
    pub sort_dir: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct QueryTableRequest {
    pub search: Option<String>,
    pub sort_by: Option<String>,
    pub sort_dir: Option<String>,
    pub limit: Option<i32>,
    pub offset: Option<i32>,
    pub group_by: Option<String>,
    pub aggregate_op: Option<String>,
    pub aggregate_column: Option<String>,
}

fn summary_from(row: DataTableRow) -> Result<DataTableSummary, String> {
    let columns: Vec<String> = serde_json::from_str(&row.column_schema)
        .map_err(|e| format!("invalid column_schema: {e}"))?;
    Ok(DataTableSummary {
        id: row.id.to_string(),
        name: row.name,
        source_filename: row.source_filename,
        source_format: row.source_format,
        columns,
        row_count: row.row_count,
        created_at: row.created_at.to_rfc3339(),
        updated_at: row.updated_at.to_rfc3339(),
    })
}

fn normalize_rows(columns: &[String], rows: Vec<Value>) -> Vec<Value> {
    rows.into_iter()
        .take(MAX_IMPORT_ROWS)
        .map(|row| {
            if let Some(obj) = row.as_object() {
                let mut out = serde_json::Map::new();
                for col in columns {
                    out.insert(col.clone(), obj.get(col).cloned().unwrap_or(Value::Null));
                }
                Value::Object(out)
            } else {
                let mut out = serde_json::Map::new();
                if let Some(first) = columns.first() {
                    out.insert(first.clone(), row);
                }
                Value::Object(out)
            }
        })
        .collect()
}

fn infer_columns(rows: &[Value]) -> Vec<String> {
    let mut cols: Vec<String> = Vec::new();
    for row in rows.iter().take(50) {
        if let Some(obj) = row.as_object() {
            for key in obj.keys() {
                if !cols.contains(key) {
                    cols.push(key.clone());
                }
            }
        }
    }
    if cols.is_empty() {
        cols.push("value".to_string());
    }
    cols
}

fn heuristic_text_table(text: &str) -> Option<AnalyzeResponse> {
    let trimmed = text.trim();
    if trimmed.is_empty() {
        return None;
    }

    let lines: Vec<&str> = trimmed
        .lines()
        .map(|l| l.trim())
        .filter(|l| !l.is_empty())
        .collect();

    if lines.len() < 2 {
        return None;
    }

    // Markdown pipe table
    if lines[0].contains('|') && lines.len() >= 2 {
        let header_line = lines[0].trim_matches('|');
        let headers: Vec<String> = header_line
            .split('|')
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
            .collect();
        if headers.len() >= 2 {
            let mut data_start = 1usize;
            if lines.get(1).map(|l| l.contains("---")).unwrap_or(false) {
                data_start = 2;
            }
            let mut rows: Vec<Value> = Vec::new();
            for line in lines.iter().skip(data_start) {
                if !line.contains('|') {
                    continue;
                }
                let cells: Vec<String> = line
                    .trim_matches('|')
                    .split('|')
                    .map(|s| s.trim().to_string())
                    .collect();
                if cells.is_empty() {
                    continue;
                }
                let mut obj = serde_json::Map::new();
                for (i, h) in headers.iter().enumerate() {
                    obj.insert(
                        h.clone(),
                        Value::String(cells.get(i).cloned().unwrap_or_default()),
                    );
                }
                rows.push(Value::Object(obj));
                if rows.len() >= 500 {
                    break;
                }
            }
            if !rows.is_empty() {
                return Some(AnalyzeResponse {
                    can_be_table: true,
                    reason: "Detected markdown-style table.".to_string(),
                    columns: headers,
                    rows,
                });
            }
        }
    }

    // Delimited text (csv/tsv)
    let delim = if lines[0].contains('\t') { '\t' } else { ',' };
    let headers: Vec<String> = lines[0]
        .split(delim)
        .map(|s| s.trim().trim_matches('"').to_string())
        .filter(|s| !s.is_empty())
        .collect();
    if headers.len() >= 2 && lines.len() >= 2 {
        let mut rows: Vec<Value> = Vec::new();
        for line in lines.iter().skip(1) {
            let cells: Vec<String> = line
                .split(delim)
                .map(|s| s.trim().trim_matches('"').to_string())
                .collect();
            if cells.iter().all(|c| c.is_empty()) {
                continue;
            }
            let mut obj = serde_json::Map::new();
            for (i, h) in headers.iter().enumerate() {
                obj.insert(
                    h.clone(),
                    Value::String(cells.get(i).cloned().unwrap_or_default()),
                );
            }
            rows.push(Value::Object(obj));
            if rows.len() >= 500 {
                break;
            }
        }
        if !rows.is_empty() {
            return Some(AnalyzeResponse {
                can_be_table: true,
                reason: "Detected delimited text table.".to_string(),
                columns: headers,
                rows,
            });
        }
    }

    None
}

async fn analyze_with_ai(
    source_filename: &str,
    source_format: &str,
    content: &str,
) -> Result<AnalyzeResponse, String> {
    let cfg = Config::from_env();
    if !cfg.configured() {
        return Err("MorphAI not configured".into());
    }
    let client = Client::new(cfg);
    let clipped = if content.len() > MAX_ANALYZE_CHARS {
        format!("{}…", &content[..MAX_ANALYZE_CHARS])
    } else {
        content.to_string()
    };

    let prompt = format!(
        r#"You extract structured tables from documents.

File: {source_filename}
Format: {source_format}

Decide if the content can be represented as a relational table (consistent columns across rows).

Reply with ONLY valid JSON (no markdown):
{{
  "can_be_table": boolean,
  "reason": "short explanation",
  "columns": ["col1", "col2"],
  "rows": [{{"col1": "a", "col2": 1}}]
}}

Rules:
- Max 500 rows in rows array.
- If not tabular, set can_be_table=false, columns=[], rows=[].
- Prefer concise column names.

Content:
{clipped}"#
    );

    let reply = client
        .chat_completion(&[Message {
            role: "user".to_string(),
            content: prompt,
        }])
        .await?;

    let json_str =
        extract_json_object(&reply).ok_or_else(|| "AI did not return JSON".to_string())?;
    let parsed: Value =
        serde_json::from_str(&json_str).map_err(|e| format!("invalid AI JSON: {e}"))?;

    let can_be_table = parsed
        .get("can_be_table")
        .and_then(|v| v.as_bool())
        .unwrap_or(false);
    let reason = parsed
        .get("reason")
        .and_then(|v| v.as_str())
        .unwrap_or("AI analysis complete")
        .to_string();
    let columns: Vec<String> = parsed
        .get("columns")
        .and_then(|v| v.as_array())
        .map(|arr| {
            arr.iter()
                .filter_map(|x| x.as_str().map(|s| s.to_string()))
                .collect()
        })
        .unwrap_or_default();
    let rows: Vec<Value> = parsed
        .get("rows")
        .and_then(|v| v.as_array())
        .map(|arr| arr.iter().take(500).cloned().collect())
        .unwrap_or_default();

    Ok(AnalyzeResponse {
        can_be_table,
        reason,
        columns,
        rows,
    })
}

fn extract_json_object(s: &str) -> Option<String> {
    let start = s.find('{')?;
    let mut depth = 0i32;
    for (i, ch) in s[start..].char_indices() {
        match ch {
            '{' => depth += 1,
            '}' => {
                depth -= 1;
                if depth == 0 {
                    return Some(s[start..start + i + 1].to_string());
                }
            }
            _ => {}
        }
    }
    None
}

pub async fn list(
    State(state): State<AppState>,
) -> Result<Json<Vec<DataTableSummary>>, (StatusCode, String)> {
    let repo = DataTableRepository::new(state.db_pool.clone());
    let rows = repo
        .list_all()
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
    let out: Result<Vec<_>, _> = rows.into_iter().map(summary_from).collect();
    Ok(Json(
        out.map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e))?,
    ))
}

pub async fn get(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<DataTableSummary>, (StatusCode, String)> {
    let uuid = Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "Invalid id".into()))?;
    let repo = DataTableRepository::new(state.db_pool.clone());
    let row = repo
        .find_by_id(uuid)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .ok_or((StatusCode::NOT_FOUND, "Data table not found".into()))?;
    Ok(Json(
        summary_from(row).map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e))?,
    ))
}

pub async fn analyze(
    Json(body): Json<AnalyzeRequest>,
) -> Result<Json<AnalyzeResponse>, (StatusCode, String)> {
    let filename = body.source_filename.as_deref().unwrap_or("upload.txt");
    let text = body.content_text.trim();
    if text.is_empty() {
        return Ok(Json(AnalyzeResponse {
            can_be_table: false,
            reason: "File is empty.".to_string(),
            columns: vec![],
            rows: vec![],
        }));
    }

    if let Some(heuristic) = heuristic_text_table(text) {
        return Ok(Json(heuristic));
    }

    match analyze_with_ai(filename, &body.source_format, text).await {
        Ok(resp) => Ok(Json(resp)),
        Err(e) => {
            if e.contains("not configured") {
                Ok(Json(AnalyzeResponse {
                    can_be_table: false,
                    reason: "Could not detect a table structure. Configure MorphAI for intelligent extraction from text, PDF, and markdown.".to_string(),
                    columns: vec![],
                    rows: vec![],
                }))
            } else {
                Err((StatusCode::BAD_GATEWAY, e))
            }
        }
    }
}

fn morph_api_base_url() -> String {
    std::env::var("MORPH_API_BASE_URL")
        .or_else(|_| std::env::var("MORPH_BASE_URL"))
        .unwrap_or_else(|_| "http://127.0.0.1:9090".into())
        .trim()
        .trim_end_matches('/')
        .to_string()
}

fn sanitize_multipart_token(s: &str) -> String {
    s.chars()
        .map(|c| match c {
            '"' | '\r' | '\n' | '\\' => '_',
            _ => c,
        })
        .collect()
}

fn rows_to_csv_excerpt(columns: &[String], rows: &[Value], max_rows: usize) -> String {
    let mut out = String::new();
    out.push_str(
        &columns
            .iter()
            .map(|c| csv_escape(c))
            .collect::<Vec<_>>()
            .join(","),
    );
    out.push('\n');
    for row in rows.iter().take(max_rows) {
        let cells: Vec<String> = columns
            .iter()
            .map(|col| {
                let v = row.get(col).cloned().unwrap_or(Value::Null);
                csv_escape(&value_as_plain(&v))
            })
            .collect();
        out.push_str(&cells.join(","));
        out.push('\n');
    }
    out
}

fn csv_escape(s: &str) -> String {
    if s.contains(',') || s.contains('"') || s.contains('\n') || s.contains('\r') {
        format!("\"{}\"", s.replace('"', "\"\""))
    } else {
        s.to_string()
    }
}

fn value_as_plain(v: &Value) -> String {
    match v {
        Value::Null => String::new(),
        Value::String(s) => s.clone(),
        Value::Bool(b) => b.to_string(),
        Value::Number(n) => n.to_string(),
        other => other.to_string(),
    }
}

async fn push_csv_to_morph_knowledge(
    auth_header: Option<&str>,
    filename: &str,
    title: &str,
    csv: &str,
) -> Result<(), String> {
    let base = morph_api_base_url();
    if base.is_empty() {
        return Err("MORPH_API_BASE_URL is empty".into());
    }
    let boundary = format!("----datax{}", Uuid::new_v4().as_simple());
    let safe_name = sanitize_multipart_token(filename);
    let safe_title = sanitize_multipart_token(title);
    let mut body = String::new();
    body.push_str(&format!(
        "--{boundary}\r\nContent-Disposition: form-data; name=\"file\"; filename=\"{safe_name}\"\r\nContent-Type: text/csv\r\n\r\n{csv}\r\n"
    ));
    body.push_str(&format!(
        "--{boundary}\r\nContent-Disposition: form-data; name=\"title\"\r\n\r\n{safe_title}\r\n"
    ));
    body.push_str(&format!(
        "--{boundary}\r\nContent-Disposition: form-data; name=\"index_to_graph\"\r\n\r\ntrue\r\n"
    ));
    body.push_str(&format!("--{boundary}--\r\n"));

    let client = reqwest::Client::new();
    let mut req = client
        .post(format!("{base}/api/knowledge/files"))
        .header(
            "Content-Type",
            format!("multipart/form-data; boundary={boundary}"),
        )
        .body(body);
    if let Some(auth) = auth_header.filter(|s| !s.trim().is_empty()) {
        req = req.header("Authorization", auth);
    }
    let resp = req
        .send()
        .await
        .map_err(|e| format!("Morph knowledge request failed: {e}"))?;
    if !resp.status().is_success() {
        let status = resp.status();
        let text = resp.text().await.unwrap_or_default();
        return Err(format!("Morph knowledge upload failed ({status}): {text}"));
    }
    Ok(())
}

pub async fn import_table(
    State(state): State<AppState>,
    headers: HeaderMap,
    Json(body): Json<ImportRequest>,
) -> Result<Json<ImportResponse>, (StatusCode, String)> {
    if body.rows.is_empty() {
        return Err((StatusCode::BAD_REQUEST, "No rows to import".into()));
    }

    let columns = if body.columns.is_empty() {
        infer_columns(&body.rows)
    } else {
        body.columns.clone()
    };

    let rows = normalize_rows(&columns, body.rows);
    let name = body
        .name
        .clone()
        .filter(|s| !s.trim().is_empty())
        .unwrap_or_else(|| {
            body.source_filename
                .clone()
                .filter(|s| !s.is_empty())
                .unwrap_or_else(|| "Imported table".to_string())
        });

    let id = Uuid::new_v4();
    let repo = DataTableRepository::new(state.db_pool.clone());
    let table = repo
        .insert_table(
            id,
            &name,
            body.source_filename.as_deref(),
            &body.source_format,
            &columns,
            None,
        )
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    repo.insert_rows(id, &rows)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let saved = repo
        .find_by_id(id)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .unwrap_or(table);

    let summary = summary_from(saved).map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e))?;

    let mut graph_status = None;
    if body.index_to_graph {
        let csv = rows_to_csv_excerpt(&columns, &rows, 500);
        let filename = body
            .source_filename
            .clone()
            .filter(|s| !s.trim().is_empty())
            .unwrap_or_else(|| format!("{}.csv", summary.name));
        let filename = if filename.to_ascii_lowercase().ends_with(".csv") {
            filename
        } else {
            format!(
                "{}.csv",
                filename
                    .rsplit_once('.')
                    .map(|(stem, _)| stem)
                    .unwrap_or(&filename)
            )
        };
        let auth = headers
            .get(axum::http::header::AUTHORIZATION)
            .and_then(|v| v.to_str().ok());
        match push_csv_to_morph_knowledge(auth, &filename, &summary.name, &csv).await {
            Ok(()) => graph_status = Some("queued for Neo4j GraphRAG via Morph Knowledge".into()),
            Err(e) => graph_status = Some(format!("table imported; Neo4j index failed: {e}")),
        }
    }

    Ok(Json(ImportResponse {
        table: summary,
        index_to_graph: body.index_to_graph,
        graph_status,
    }))
}

fn parse_data_rows(
    data_rows: Vec<crate::db::data_table_repository::DataTableDataRow>,
) -> Vec<ParsedRow> {
    data_rows
        .into_iter()
        .filter_map(|r| {
            let data: Value = serde_json::from_str(&r.data).ok()?;
            Some(ParsedRow {
                row_index: r.row_index,
                data,
            })
        })
        .collect()
}

fn parsed_to_response_rows(rows: Vec<ParsedRow>) -> Vec<Value> {
    rows.into_iter()
        .map(|r| {
            let mut v = r.data;
            if let Some(obj) = v.as_object_mut() {
                obj.insert("_row_index".to_string(), json!(r.row_index));
            }
            v
        })
        .collect()
}

pub async fn get_rows(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Query(q): Query<RowsQuery>,
) -> Result<Json<DataTableRowsResponse>, (StatusCode, String)> {
    let uuid = Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "Invalid id".into()))?;
    let limit = q.limit.unwrap_or(50).clamp(1, 500);
    let offset = q.offset.unwrap_or(0).max(0);
    let has_filter = q
        .search
        .as_ref()
        .map(|s| !s.trim().is_empty())
        .unwrap_or(false)
        || q.sort_by
            .as_ref()
            .map(|s| !s.trim().is_empty())
            .unwrap_or(false);

    let repo = DataTableRepository::new(state.db_pool.clone());
    let table = repo
        .find_by_id(uuid)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .ok_or((StatusCode::NOT_FOUND, "Data table not found".into()))?;

    let columns: Vec<String> = serde_json::from_str(&table.column_schema)
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let (rows, total) = if has_filter {
        let all = repo
            .fetch_all_rows(uuid)
            .await
            .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
        let parsed = parse_data_rows(all);
        let opts = RowQueryOptions {
            search: q.search.clone(),
            sort_by: q.sort_by.clone(),
            sort_dir: q.sort_dir.unwrap_or_else(|| "asc".to_string()),
            limit,
            offset,
        };
        apply_row_query(parsed, &columns, &opts)
    } else {
        let data_rows = repo
            .fetch_rows(uuid, limit, offset)
            .await
            .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
        let parsed = parse_data_rows(data_rows);
        (parsed, table.row_count)
    };

    Ok(Json(DataTableRowsResponse {
        columns,
        rows: parsed_to_response_rows(rows),
        total,
        limit,
        offset,
    }))
}

pub async fn query_table(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(body): Json<QueryTableRequest>,
) -> Result<Json<Value>, (StatusCode, String)> {
    let uuid = Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "Invalid id".into()))?;
    let repo = DataTableRepository::new(state.db_pool.clone());
    let table = repo
        .find_by_id(uuid)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?
        .ok_or((StatusCode::NOT_FOUND, "Data table not found".into()))?;

    let columns: Vec<String> = serde_json::from_str(&table.column_schema)
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;

    let all = repo
        .fetch_all_rows(uuid)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
    let parsed = parse_data_rows(all);

    let sort_dir = body.sort_dir.unwrap_or_else(|| "asc".to_string());
    let filter_opts = RowQueryOptions {
        search: body.search.clone(),
        sort_by: body.sort_by.clone(),
        sort_dir: sort_dir.clone(),
        limit: 5000,
        offset: 0,
    };
    let (filtered, total) = apply_row_query(parsed, &columns, &filter_opts);

    let page_opts = RowQueryOptions {
        search: None,
        sort_by: None,
        sort_dir,
        limit: body.limit.unwrap_or(100).clamp(1, 500),
        offset: body.offset.unwrap_or(0).max(0),
    };
    let (page, _) = apply_row_query(filtered.clone(), &columns, &page_opts);

    let aggregate = if let Some(op) = body
        .aggregate_op
        .as_ref()
        .map(|s| s.trim())
        .filter(|s| !s.is_empty())
    {
        Some(aggregate_rows(
            &filtered,
            body.group_by.as_deref(),
            op,
            body.aggregate_column.as_deref(),
        ))
    } else {
        None
    };

    Ok(Json(json!({
        "table_id": id,
        "name": table.name,
        "columns": columns,
        "total": total,
        "rows": parsed_to_response_rows(page),
        "aggregate": aggregate
    })))
}

pub async fn update_row(
    State(state): State<AppState>,
    Path((id, row_index)): Path<(String, i32)>,
    Json(body): Json<UpdateRowRequest>,
) -> Result<Json<Value>, (StatusCode, String)> {
    let uuid = Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "Invalid id".into()))?;
    let repo = DataTableRepository::new(state.db_pool.clone());
    let updated = repo
        .update_row_data(uuid, row_index, &body.data)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
    if !updated {
        return Err((StatusCode::NOT_FOUND, "Row not found".into()));
    }
    Ok(Json(json!({ "ok": true })))
}

pub async fn delete_row(
    State(state): State<AppState>,
    Path((id, row_index)): Path<(String, i32)>,
) -> Result<Json<Value>, (StatusCode, String)> {
    let uuid = Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "Invalid id".into()))?;
    let repo = DataTableRepository::new(state.db_pool.clone());
    let deleted = repo
        .delete_row(uuid, row_index)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
    if !deleted {
        return Err((StatusCode::NOT_FOUND, "Row not found".into()));
    }
    Ok(Json(json!({ "ok": true })))
}

pub async fn delete_table(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<Json<Value>, (StatusCode, String)> {
    let uuid = Uuid::parse_str(&id).map_err(|_| (StatusCode::BAD_REQUEST, "Invalid id".into()))?;
    let repo = DataTableRepository::new(state.db_pool.clone());
    let deleted = repo
        .delete_table(uuid)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
    if !deleted {
        return Err((StatusCode::NOT_FOUND, "Data table not found".into()));
    }
    Ok(Json(json!({ "ok": true })))
}
