//! Morph Engi AI assistant — MorphAI tool loop covering all app modules.

use axum::{extract::State, http::StatusCode, Json};
use morphai::{extract_json_object, tool_follow_up_prompt, truncate_chars, Message, DEFAULT_TOOL_MAX_ROUNDS, DEFAULT_TOOL_RESULT_MAX_CHARS};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use sqlx::{Column, Row};
use std::sync::Arc;

use crate::api::extract::{arg_f64, arg_i64, arg_str, AuthUser};
use crate::middleware::auth::AuthContext;
use crate::services::AppState;

const ENGI_INSTRUCTIONS: &str = r#"You are Morph Engi AI, a concise assistant for project management (civil / construction / field teams).

You help with:
- Projects (jobs, status, descriptions)
- Site logs
- Project files (links and uploads)
- People (contractors and contacts)

The UI sends JSON `app_context` with live records — prefer that before calling tools.

When you need live data or must create/update records, respond with ONLY one JSON object (no markdown):
{"tool":"<name>","args":{...}}

Primary read tools:
- list_projects — optional status
- get_project — args: id
- list_tasks — optional project_id
- list_site_logs — optional project_id
- list_resource_files — optional project_id
- list_contractors — optional project_id

Primary write tools:
- create_project — args: code, name; optional client, location, status, description
- update_project — args: id, plus fields to change
- create_task — args: project_id, title; optional assignee, status, priority, due_date
- create_site_log — args: project_id, log_date, summary; optional weather, crew_count, issues
- create_resource_file — args: name; optional source_type (url|upload), file_url, file_name, description, project_ids[]
- create_contractor — args: name; optional trade, contact_name, phone, email, description, project_ids[]

Use active_project_id from assistant state when the user says "this project".
After TOOL_RESULT, summarize in markdown.
For destructive actions, ask for confirmation in markdown only.
If no tool needed, reply in markdown only."#;

#[derive(Deserialize)]
struct ChatMessage {
    role: String,
    content: String,
}

#[derive(Deserialize, Serialize, Default, Clone)]
struct AssistantState {
    intent: Option<String>,
    #[serde(default)]
    fields: Value,
}

#[derive(Deserialize)]
pub struct AssistantChatRequest {
    messages: Vec<ChatMessage>,
    #[serde(default)]
    state: AssistantState,
}

#[derive(Deserialize)]
struct ToolCall {
    tool: String,
    args: Value,
}

pub async fn assistant_chat(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Json(req): Json<AssistantChatRequest>,
) -> Result<Json<Value>, (StatusCode, Json<Value>)> {
    let last_user = req
        .messages
        .iter()
        .rev()
        .find(|m| m.role == "user")
        .map(|m| m.content.trim())
        .unwrap_or("")
        .to_string();

    if last_user.is_empty() {
        return Ok(Json(json!({
            "assistant_message": "How can I help with your civil engineering projects today?",
            "intent": "general",
            "state": req.state,
            "completed": true
        })));
    }

    // Deterministic shortcuts without LLM
    if last_user.eq_ignore_ascii_case("help") {
        return Ok(Json(json!({
            "assistant_message": help_text(),
            "intent": "general",
            "state": req.state,
            "completed": true
        })));
    }

    let Some(ai) = state.ai.as_ref().as_ref() else {
        return Ok(Json(json!({
            "assistant_message": format!("{}\n\n_Set `MORPH_AI_API_KEY` to enable the full LLM assistant._", help_text()),
            "intent": "general",
            "state": req.state,
            "completed": true
        })));
    };

    let mut prompt = ENGI_INSTRUCTIONS.to_string();
    if let Some(hist) = format_history(&req.messages) {
        prompt.push_str("\n\nRecent conversation:\n");
        prompt.push_str(&hist);
    }
    if !req.state.fields.is_null() && req.state.fields != json!({}) {
        if let Some(ctx) = req.state.fields.get("app_context").and_then(|v| v.as_str()) {
            prompt.push_str("\n\nLive app context (JSON — includes descriptions and all visible records; use for answers and tools):\n");
            prompt.push_str(&truncate_chars(ctx, 14000));
        }
        prompt.push_str("\n\nAssistant state fields:\n");
        prompt.push_str(&truncate_chars(&req.state.fields.to_string(), 4000));
    }
    prompt.push_str("\n\nLatest user message:\n");
    prompt.push_str(&last_user);

    let mut messages = vec![Message {
        role: "user".to_string(),
        content: prompt,
    }];
    let mut last_record: Option<Value> = None;

    for _ in 0..DEFAULT_TOOL_MAX_ROUNDS {
        let reply = ai
            .chat_completion(&messages)
            .await
            .map_err(|e| {
                (
                    StatusCode::BAD_GATEWAY,
                    Json(json!({"error": e.to_string()})),
                )
            })?;
        let reply = reply.trim();
        if reply.is_empty() {
            break;
        }

        let Some(obj) = extract_json_object(reply) else {
            return Ok(Json(json!({
                "assistant_message": reply,
                "intent": req.state.intent.clone().unwrap_or_else(|| "general".into()),
                "state": req.state,
                "completed": true,
                "record": last_record
            })));
        };

        let call: ToolCall = match serde_json::from_str(&obj) {
            Ok(c) => c,
            Err(_) => {
                return Ok(Json(json!({
                    "assistant_message": reply,
                    "intent": "general",
                    "state": req.state,
                    "completed": true
                })));
            }
        };

        let (result, record) = exec_tool(&state, &auth, &call).await;
        if record.is_some() {
            last_record = record;
        }
        let tool_msg = format!(
            "TOOL_RESULT\n{}",
            truncate_chars(&result, DEFAULT_TOOL_RESULT_MAX_CHARS)
        );
        messages.push(Message {
            role: "user".to_string(),
            content: tool_follow_up_prompt(&tool_msg),
        });
    }

    Ok(Json(json!({
        "assistant_message": "I hit the tool limit while gathering data. Please narrow your question.",
        "intent": "general",
        "state": req.state,
        "completed": true,
        "record": last_record
    })))
}

fn help_text() -> &'static str {
    "I can help with **projects**, **site logs**, **files**, and **people**.\n\nTry:\n- `List active projects`\n- `Create project code: BRG-01 name: Riverside Bridge`\n- `Add a site log for this project`\n- `Who are the contractors on this project?`"
}

fn format_history(messages: &[ChatMessage]) -> Option<String> {
    let start = messages.len().saturating_sub(12);
    let slice = &messages[start..];
    if slice.is_empty() {
        return None;
    }
    Some(
        slice
            .iter()
            .filter(|m| m.role == "user" || m.role == "assistant")
            .map(|m| format!("{}: {}", m.role, truncate_chars(&m.content, 800)))
            .collect::<Vec<_>>()
            .join("\n"),
    )
}

async fn exec_tool(state: &AppState, auth: &AuthContext, call: &ToolCall) -> (String, Option<Value>) {
    let org = auth.org_id;
    let args = &call.args;
    let tool = call.tool.trim();

    let result = match tool {
        "dashboard_summary" => {
            match sqlx::query_scalar::<_, i64>("SELECT COUNT(*) FROM projects WHERE organization_id = ?")
                .bind(org)
                .fetch_one(&state.pool)
                .await
            {
                Ok(n) => json!({"projects_total": n}).to_string(),
                Err(e) => format!("error: {e}"),
            }
        }
        "list_projects" => list_projects_tool(&state.pool, org, arg_str(args, "status")).await,
        "get_project" => {
            let id = arg_i64(args, "id", 0);
            get_project_tool(&state.pool, org, id).await
        }
        "list_tasks" => list_simple(&state.pool, "project_tasks", org, arg_i64(args, "project_id", 0)).await,
        "list_site_logs" => list_simple(&state.pool, "site_logs", org, arg_i64(args, "project_id", 0)).await,
        "list_materials" => {
            match sqlx::query("SELECT id, name, category, unit, stock_qty FROM materials WHERE organization_id = ? ORDER BY name")
                .bind(org).fetch_all(&state.pool).await {
                Ok(rows) => json!({"materials": rows.iter().map(|r| json!({
                    "id": r.get::<i64,_>("id"),
                    "name": r.get::<String,_>("name"),
                    "category": r.get::<String,_>("category"),
                    "unit": r.get::<String,_>("unit"),
                    "stock_qty": r.get::<f64,_>("stock_qty"),
                })).collect::<Vec<_>>()}).to_string(),
                Err(e) => format!("error: {e}"),
            }
        }
        "list_material_usages" => list_simple(&state.pool, "material_usages", org, arg_i64(args, "project_id", 0)).await,
        "list_resource_files" => list_resource_files_tool(&state.pool, org, arg_i64(args, "project_id", 0)).await,
        "get_project_finance" => {
            let pid = arg_i64(args, "project_id", arg_i64(args, "active_project_id", 0));
            if pid == 0 { "error: project_id required".into() } else {
                match sqlx::query("SELECT * FROM project_finances WHERE organization_id = ? AND project_id = ?")
                    .bind(org).bind(pid).fetch_optional(&state.pool).await {
                    Ok(Some(r)) => row_to_finance_tool(&r).to_string(),
                    Ok(None) => json!({"finance": null}).to_string(),
                    Err(e) => format!("error: {e}"),
                }
            }
        }
        "list_budget_lines" => list_simple(&state.pool, "budget_lines", org, arg_i64(args, "project_id", 0)).await,
        "list_resources" => list_simple(&state.pool, "resources", org, 0).await,
        "list_resource_allocations" => list_simple(&state.pool, "resource_allocations", org, arg_i64(args, "project_id", 0)).await,
        "list_contractors" => list_contractors_tool(&state.pool, org, arg_i64(args, "project_id", 0)).await,
        "list_contracts" => list_simple(&state.pool, "contractor_contracts", org, arg_i64(args, "project_id", 0)).await,
        "list_relations" => list_simple(&state.pool, "public_relations", org, arg_i64(args, "project_id", 0)).await,
        "list_communications" => list_simple(&state.pool, "communications", org, arg_i64(args, "project_id", 0)).await,
        "create_project" => {
            let code = arg_str(args, "code");
            let name = arg_str(args, "name");
            if code.is_empty() || name.is_empty() {
                "error: code and name required".into()
            } else {
                match sqlx::query(
                    "INSERT INTO projects (organization_id, code, name, client, location, status, budget_total, progress_pct, description) VALUES (?,?,?,?,?,?,?,?,?)",
                )
                .bind(org)
                .bind(&code)
                .bind(&name)
                .bind(arg_str(args, "client"))
                .bind(arg_str(args, "location"))
                .bind(if arg_str(args, "status").is_empty() { "planning".to_string() } else { arg_str(args, "status") })
                .bind(arg_f64(args, "budget_total", 0.0))
                .bind(arg_f64(args, "progress_pct", 0.0))
                .bind(arg_str(args, "description"))
                .execute(&state.pool)
                .await
                {
                    Ok(r) => {
                        let rec = json!({"id": r.last_insert_rowid(), "code": code, "name": name});
                        return (rec.to_string(), Some(rec));
                    }
                    Err(e) => format!("error: {e}"),
                }
            }
        }
        "update_project" => {
            let id = arg_i64(args, "id", 0);
            if id == 0 {
                "error: id required".into()
            } else {
                match sqlx::query(
                    "UPDATE projects SET code=COALESCE(NULLIF(?,''), code), name=COALESCE(NULLIF(?,''), name), status=COALESCE(NULLIF(?,''), status), budget_total=CASE WHEN ?>=0 THEN ? ELSE budget_total END, progress_pct=CASE WHEN ?>=0 THEN ? ELSE progress_pct END, updated_at=datetime('now') WHERE id=? AND organization_id=?",
                )
                .bind(arg_str(args, "code"))
                .bind(arg_str(args, "name"))
                .bind(arg_str(args, "status"))
                .bind(arg_f64(args, "budget_total", -1.0))
                .bind(arg_f64(args, "budget_total", -1.0))
                .bind(arg_f64(args, "progress_pct", -1.0))
                .bind(arg_f64(args, "progress_pct", -1.0))
                .bind(id)
                .bind(org)
                .execute(&state.pool)
                .await
                {
                    Ok(_) => format!("updated project {id}"),
                    Err(e) => format!("error: {e}"),
                }
            }
        }
        "create_task" => {
            let pid = arg_i64(args, "project_id", arg_i64(args, "active_project_id", 0));
            let title = arg_str(args, "title");
            if pid == 0 || title.is_empty() {
                "error: project_id and title required".into()
            } else {
                match sqlx::query(
                    "INSERT INTO project_tasks (organization_id, project_id, title, assignee, status, priority, due_date) VALUES (?,?,?,?,?,?,?)",
                )
                .bind(org)
                .bind(pid)
                .bind(&title)
                .bind(arg_str(args, "assignee"))
                .bind(if arg_str(args, "status").is_empty() { "open".to_string() } else { arg_str(args, "status") })
                .bind(if arg_str(args, "priority").is_empty() { "normal".to_string() } else { arg_str(args, "priority") })
                .bind(if arg_str(args, "due_date").is_empty() { None::<String> } else { Some(arg_str(args, "due_date")) })
                .execute(&state.pool)
                .await
                {
                    Ok(r) => json!({"id": r.last_insert_rowid(), "title": title}).to_string(),
                    Err(e) => format!("error: {e}"),
                }
            }
        }
        "create_site_log" => {
            let pid = arg_i64(args, "project_id", 0);
            let summary = arg_str(args, "summary");
            let log_date = if arg_str(args, "log_date").is_empty() {
                chrono::Utc::now().format("%Y-%m-%d").to_string()
            } else {
                arg_str(args, "log_date")
            };
            if pid == 0 || summary.is_empty() {
                "error: project_id and summary required".into()
            } else {
                match sqlx::query(
                    "INSERT INTO site_logs (organization_id, project_id, log_date, weather, crew_count, summary, issues) VALUES (?,?,?,?,?,?,?)",
                )
                .bind(org)
                .bind(pid)
                .bind(&log_date)
                .bind(arg_str(args, "weather"))
                .bind(arg_i64(args, "crew_count", 0))
                .bind(&summary)
                .bind(arg_str(args, "issues"))
                .execute(&state.pool)
                .await
                {
                    Ok(r) => json!({"id": r.last_insert_rowid()}).to_string(),
                    Err(e) => format!("error: {e}"),
                }
            }
        }
        "create_material" => {
            match sqlx::query(
                "INSERT INTO materials (organization_id, name, category, unit, unit_cost, supplier, stock_qty, reorder_level, description) VALUES (?,?,?,?,?,?,?,?,?)",
            )
            .bind(org)
            .bind(arg_str(args, "name"))
            .bind(arg_str(args, "category"))
            .bind(if arg_str(args, "unit").is_empty() { "ea".to_string() } else { arg_str(args, "unit") })
            .bind(arg_f64(args, "unit_cost", 0.0))
            .bind(arg_str(args, "supplier"))
            .bind(arg_f64(args, "stock_qty", 0.0))
            .bind(arg_f64(args, "reorder_level", 0.0))
            .bind(arg_str(args, "description"))
            .execute(&state.pool)
            .await
            {
                Ok(r) => {
                    let id = r.last_insert_rowid();
                    let pids = project_ids_from_body(args);
                    if !pids.is_empty() {
                        let _ = set_project_links_tool(&state.pool, "project_materials", "material_id", id, &pids).await;
                    }
                    json!({"id": id}).to_string()
                }
                Err(e) => format!("error: {e}"),
            }
        }
        "record_material_usage" => {
            let pid = arg_i64(args, "project_id", 0);
            let mid = arg_i64(args, "material_id", 0);
            let qty = arg_f64(args, "qty", 0.0);
            if pid == 0 || mid == 0 || qty <= 0.0 {
                "error: project_id, material_id, qty required".into()
            } else {
                let used_at = if arg_str(args, "used_at").is_empty() {
                    chrono::Utc::now().format("%Y-%m-%d").to_string()
                } else {
                    arg_str(args, "used_at")
                };
                let _ = sqlx::query(
                    "INSERT INTO material_usages (organization_id, project_id, material_id, qty, used_at, notes) VALUES (?,?,?,?,?,?)",
                )
                .bind(org).bind(pid).bind(mid).bind(qty).bind(&used_at).bind(arg_str(args, "notes"))
                .execute(&state.pool).await;
                let _ = sqlx::query("UPDATE materials SET stock_qty = stock_qty - ? WHERE id = ? AND organization_id = ?")
                    .bind(qty).bind(mid).bind(org).execute(&state.pool).await;
                format!("recorded {qty} units on project {pid}")
            }
        }
        "create_resource_file" => {
            let name = arg_str(args, "name");
            if name.is_empty() {
                "error: name required".into()
            } else {
                match sqlx::query(
                    "INSERT INTO resource_files (organization_id, name, source_type, file_url, file_name, description) VALUES (?,?,?,?,?,?)",
                )
                .bind(org)
                .bind(&name)
                .bind(if arg_str(args, "source_type").is_empty() { "url".into() } else { arg_str(args, "source_type") })
                .bind(arg_str(args, "file_url"))
                .bind(arg_str(args, "file_name"))
                .bind(arg_str(args, "description"))
                .execute(&state.pool)
                .await
                {
                    Ok(r) => {
                        let id = r.last_insert_rowid();
                        let pids = project_ids_from_body(args);
                        if pids.is_empty() {
                            let pid = arg_i64(args, "project_id", arg_i64(args, "active_project_id", 0));
                            if pid > 0 {
                                let _ = set_project_links_tool(&state.pool, "project_resource_files", "resource_file_id", id, &[pid]).await;
                            }
                        } else {
                            let _ = set_project_links_tool(&state.pool, "project_resource_files", "resource_file_id", id, &pids).await;
                        }
                        json!({"id": id}).to_string()
                    }
                    Err(e) => format!("error: {e}"),
                }
            }
        }
        "upsert_project_finance" => {
            let pid = arg_i64(args, "project_id", arg_i64(args, "active_project_id", 0));
            if pid == 0 {
                "error: project_id required".into()
            } else {
                let existing = sqlx::query("SELECT id FROM project_finances WHERE organization_id = ? AND project_id = ?")
                    .bind(org).bind(pid).fetch_optional(&state.pool).await;
                match existing {
                    Ok(Some(row)) => {
                        let id = row.get::<i64, _>("id");
                        match sqlx::query("UPDATE project_finances SET summary=?, total_planned=?, total_actual=?, notes=?, updated_at=datetime('now') WHERE id=? AND organization_id=?")
                            .bind(arg_str(args, "summary")).bind(arg_f64(args, "total_planned", 0.0)).bind(arg_f64(args, "total_actual", 0.0)).bind(arg_str(args, "notes")).bind(id).bind(org)
                            .execute(&state.pool).await {
                            Ok(_) => format!("updated finance for project {pid}"),
                            Err(e) => format!("error: {e}"),
                        }
                    }
                    Ok(None) => {
                        match sqlx::query("INSERT INTO project_finances (organization_id, project_id, summary, total_planned, total_actual, notes) VALUES (?,?,?,?,?,?)")
                            .bind(org).bind(pid).bind(arg_str(args, "summary")).bind(arg_f64(args, "total_planned", 0.0)).bind(arg_f64(args, "total_actual", 0.0)).bind(arg_str(args, "notes"))
                            .execute(&state.pool).await {
                            Ok(r) => json!({"id": r.last_insert_rowid()}).to_string(),
                            Err(e) => format!("error: {e}"),
                        }
                    }
                    Err(e) => format!("error: {e}"),
                }
            }
        }
        "create_budget_line" => {
            match sqlx::query(
                "INSERT INTO budget_lines (organization_id, project_id, cost_code, description, category, planned_amount, actual_amount) VALUES (?,?,?,?,?,?,?)",
            )
            .bind(org)
            .bind(arg_i64(args, "project_id", 0))
            .bind(arg_str(args, "cost_code"))
            .bind(arg_str(args, "description"))
            .bind(if arg_str(args, "category").is_empty() { "material".into() } else { arg_str(args, "category") })
            .bind(arg_f64(args, "planned_amount", 0.0))
            .bind(arg_f64(args, "actual_amount", 0.0))
            .execute(&state.pool)
            .await
            {
                Ok(r) => json!({"id": r.last_insert_rowid()}).to_string(),
                Err(e) => format!("error: {e}"),
            }
        }
        "update_budget_line" => {
            let id = arg_i64(args, "id", 0);
            match sqlx::query(
                "UPDATE budget_lines SET planned_amount=CASE WHEN ?>=0 THEN ? ELSE planned_amount END, actual_amount=CASE WHEN ?>=0 THEN ? ELSE actual_amount END WHERE id=? AND organization_id=?",
            )
            .bind(arg_f64(args, "planned_amount", -1.0))
            .bind(arg_f64(args, "planned_amount", -1.0))
            .bind(arg_f64(args, "actual_amount", -1.0))
            .bind(arg_f64(args, "actual_amount", -1.0))
            .bind(id)
            .bind(org)
            .execute(&state.pool)
            .await
            {
                Ok(_) => format!("updated budget line {id}"),
                Err(e) => format!("error: {e}"),
            }
        }
        "create_resource" => {
            match sqlx::query(
                "INSERT INTO resources (organization_id, name, resource_type, cost_per_day, availability) VALUES (?,?,?,?,?)",
            )
            .bind(org)
            .bind(arg_str(args, "name"))
            .bind(if arg_str(args, "resource_type").is_empty() { "equipment".into() } else { arg_str(args, "resource_type") })
            .bind(arg_f64(args, "cost_per_day", 0.0))
            .bind(if arg_str(args, "availability").is_empty() { "available".into() } else { arg_str(args, "availability") })
            .execute(&state.pool)
            .await
            {
                Ok(r) => json!({"id": r.last_insert_rowid()}).to_string(),
                Err(e) => format!("error: {e}"),
            }
        }
        "allocate_resource" => {
            match sqlx::query(
                "INSERT INTO resource_allocations (organization_id, project_id, resource_id, start_date, end_date, qty, notes) VALUES (?,?,?,?,?,?,?)",
            )
            .bind(org)
            .bind(arg_i64(args, "project_id", 0))
            .bind(arg_i64(args, "resource_id", 0))
            .bind(arg_str(args, "start_date"))
            .bind(if arg_str(args, "end_date").is_empty() { None::<String> } else { Some(arg_str(args, "end_date")) })
            .bind(arg_f64(args, "qty", 1.0))
            .bind(arg_str(args, "notes"))
            .execute(&state.pool)
            .await
            {
                Ok(r) => json!({"id": r.last_insert_rowid()}).to_string(),
                Err(e) => format!("error: {e}"),
            }
        }
        "create_contractor" => {
            match sqlx::query(
                "INSERT INTO contractors (organization_id, name, trade, contact_name, phone, email, rating, status, description) VALUES (?,?,?,?,?,?,?,?,?)",
            )
            .bind(org)
            .bind(arg_str(args, "name"))
            .bind(arg_str(args, "trade"))
            .bind(arg_str(args, "contact_name"))
            .bind(arg_str(args, "phone"))
            .bind(arg_str(args, "email"))
            .bind(arg_f64(args, "rating", 0.0))
            .bind(if arg_str(args, "status").is_empty() { "active".into() } else { arg_str(args, "status") })
            .bind(arg_str(args, "description"))
            .execute(&state.pool)
            .await
            {
                Ok(r) => {
                    let id = r.last_insert_rowid();
                    let pids = project_ids_from_body(args);
                    if !pids.is_empty() {
                        let _ = set_project_links_tool(&state.pool, "project_contractors", "contractor_id", id, &pids).await;
                    }
                    json!({"id": id}).to_string()
                }
                Err(e) => format!("error: {e}"),
            }
        }
        "create_contract" => {
            match sqlx::query(
                "INSERT INTO contractor_contracts (organization_id, project_id, contractor_id, scope, contract_value, status, start_date, end_date) VALUES (?,?,?,?,?,?,?,?)",
            )
            .bind(org)
            .bind(arg_i64(args, "project_id", 0))
            .bind(arg_i64(args, "contractor_id", 0))
            .bind(arg_str(args, "scope"))
            .bind(arg_f64(args, "contract_value", 0.0))
            .bind(if arg_str(args, "status").is_empty() { "draft".into() } else { arg_str(args, "status") })
            .bind(if arg_str(args, "start_date").is_empty() { None::<String> } else { Some(arg_str(args, "start_date")) })
            .bind(if arg_str(args, "end_date").is_empty() { None::<String> } else { Some(arg_str(args, "end_date")) })
            .execute(&state.pool)
            .await
            {
                Ok(r) => json!({"id": r.last_insert_rowid()}).to_string(),
                Err(e) => format!("error: {e}"),
            }
        }
        "create_relation" => {
            let pid = arg_i64(args, "project_id", arg_i64(args, "active_project_id", 0));
            let name = arg_str(args, "name");
            if pid == 0 || name.is_empty() {
                "error: project_id and name required".into()
            } else {
                match sqlx::query(
                    "INSERT INTO public_relations (organization_id, project_id, name, role, contact, influence, sentiment, description) VALUES (?,?,?,?,?,?,?,?)",
                )
                .bind(org)
                .bind(pid)
                .bind(&name)
                .bind(arg_str(args, "role"))
                .bind(arg_str(args, "contact"))
                .bind(if arg_str(args, "influence").is_empty() { "medium".into() } else { arg_str(args, "influence") })
                .bind(if arg_str(args, "sentiment").is_empty() { "neutral".into() } else { arg_str(args, "sentiment") })
                .bind(arg_str(args, "description"))
                .execute(&state.pool)
                .await
                {
                    Ok(r) => json!({"id": r.last_insert_rowid()}).to_string(),
                    Err(e) => format!("error: {e}"),
                }
            }
        }
        "log_communication" => {
            match sqlx::query(
                "INSERT INTO communications (organization_id, project_id, relation_id, channel, subject, body, occurred_at) VALUES (?,?,?,?,?,?,?)",
            )
            .bind(org)
            .bind(if args.get("project_id").is_some() { Some(arg_i64(args, "project_id", 0)) } else { None::<i64> })
            .bind(if args.get("relation_id").is_some() { Some(arg_i64(args, "relation_id", 0)) } else { None::<i64> })
            .bind(if arg_str(args, "channel").is_empty() { "meeting".into() } else { arg_str(args, "channel") })
            .bind(arg_str(args, "subject"))
            .bind(arg_str(args, "body"))
            .bind(arg_str(args, "occurred_at"))
            .execute(&state.pool)
            .await
            {
                Ok(r) => json!({"id": r.last_insert_rowid()}).to_string(),
                Err(e) => format!("error: {e}"),
            }
        }
        _ => format!("unknown tool: {tool}"),
    };

    (result, None)
}

async fn list_projects_tool(pool: &sqlx::SqlitePool, org: i64, status: String) -> String {
    let rows = if status.is_empty() {
        sqlx::query("SELECT id, code, name, status, progress_pct, budget_total FROM projects WHERE organization_id = ? ORDER BY id DESC")
            .bind(org).fetch_all(pool).await
    } else {
        sqlx::query("SELECT id, code, name, status, progress_pct, budget_total FROM projects WHERE organization_id = ? AND status = ? ORDER BY id DESC")
            .bind(org).bind(status).fetch_all(pool).await
    };
    match rows {
        Ok(rs) => json!({"projects": rs.iter().map(|r| json!({
            "id": r.get::<i64,_>("id"),
            "code": r.get::<String,_>("code"),
            "name": r.get::<String,_>("name"),
            "status": r.get::<String,_>("status"),
            "progress_pct": r.get::<f64,_>("progress_pct"),
            "budget_total": r.get::<f64,_>("budget_total"),
        })).collect::<Vec<_>>()}).to_string(),
        Err(e) => format!("error: {e}"),
    }
}

async fn get_project_tool(pool: &sqlx::SqlitePool, org: i64, id: i64) -> String {
    match sqlx::query("SELECT * FROM projects WHERE id = ? AND organization_id = ?")
        .bind(id)
        .bind(org)
        .fetch_optional(pool)
        .await
    {
        Ok(Some(r)) => json!({
            "id": r.get::<i64,_>("id"),
            "code": r.get::<String,_>("code"),
            "name": r.get::<String,_>("name"),
            "client": r.get::<String,_>("client"),
            "location": r.get::<String,_>("location"),
            "status": r.get::<String,_>("status"),
            "budget_total": r.get::<f64,_>("budget_total"),
            "progress_pct": r.get::<f64,_>("progress_pct"),
            "description": r.get::<String,_>("description"),
        })
        .to_string(),
        Ok(None) => "not found".into(),
        Err(e) => format!("error: {e}"),
    }
}

async fn list_simple(pool: &sqlx::SqlitePool, table: &str, org: i64, project_id: i64) -> String {
    let q = if project_id > 0 {
        format!("SELECT * FROM {table} WHERE organization_id = ? AND project_id = ? ORDER BY id DESC LIMIT 100")
    } else {
        format!("SELECT * FROM {table} WHERE organization_id = ? ORDER BY id DESC LIMIT 100")
    };
    let rows = if project_id > 0 {
        sqlx::query(&q).bind(org).bind(project_id).fetch_all(pool).await
    } else {
        sqlx::query(&q).bind(org).fetch_all(pool).await
    };
    match rows {
        Ok(rs) => {
            let items: Vec<Value> = rs
                .iter()
                .map(|row| {
                    let cols = row.columns();
                    let mut obj = serde_json::Map::new();
                    for col in cols {
                        let name = col.name();
                        let val: Value = if let Ok(v) = row.try_get::<i64, _>(name) {
                            json!(v)
                        } else if let Ok(v) = row.try_get::<f64, _>(name) {
                            json!(v)
                        } else if let Ok(v) = row.try_get::<String, _>(name) {
                            json!(v)
                        } else {
                            json!(null)
                        };
                        obj.insert(name.to_string(), val);
                    }
                    Value::Object(obj)
                })
                .collect();
            json!({ table: items }).to_string()
        }
        Err(e) => format!("error: {e}"),
    }
}

fn project_ids_from_body(args: &Value) -> Vec<i64> {
    args.get("project_ids")
        .and_then(|v| v.as_array())
        .map(|arr| {
            arr.iter()
                .filter_map(|x| x.as_i64().or_else(|| x.as_str()?.parse().ok()))
                .filter(|id| *id > 0)
                .collect()
        })
        .unwrap_or_default()
}

async fn set_project_links_tool(
    pool: &sqlx::SqlitePool,
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

async fn list_resource_files_tool(pool: &sqlx::SqlitePool, org: i64, project_id: i64) -> String {
    let rows = if project_id > 0 {
        sqlx::query("SELECT rf.* FROM resource_files rf INNER JOIN project_resource_files prf ON prf.resource_file_id = rf.id WHERE rf.organization_id = ? AND prf.project_id = ? ORDER BY rf.id DESC")
            .bind(org).bind(project_id).fetch_all(pool).await
    } else {
        sqlx::query("SELECT * FROM resource_files WHERE organization_id = ? ORDER BY id DESC LIMIT 100")
            .bind(org).fetch_all(pool).await
    };
    match rows {
        Ok(rs) => json!({"resource_files": rs.iter().map(|r| json!({
            "id": r.get::<i64,_>("id"),
            "name": r.get::<String,_>("name"),
            "source_type": r.get::<String,_>("source_type"),
            "file_url": r.get::<String,_>("file_url"),
            "description": r.get::<String,_>("description"),
        })).collect::<Vec<_>>()}).to_string(),
        Err(e) => format!("error: {e}"),
    }
}

async fn list_contractors_tool(pool: &sqlx::SqlitePool, org: i64, project_id: i64) -> String {
    let rows = if project_id > 0 {
        sqlx::query("SELECT c.* FROM contractors c INNER JOIN project_contractors pc ON pc.contractor_id = c.id WHERE c.organization_id = ? AND pc.project_id = ? ORDER BY c.name")
            .bind(org).bind(project_id).fetch_all(pool).await
    } else {
        sqlx::query("SELECT * FROM contractors WHERE organization_id = ? ORDER BY name LIMIT 100")
            .bind(org).fetch_all(pool).await
    };
    match rows {
        Ok(rs) => json!({"contractors": rs.iter().map(|r| json!({
            "id": r.get::<i64,_>("id"),
            "name": r.get::<String,_>("name"),
            "trade": r.get::<String,_>("trade"),
            "description": r.try_get::<String,_>("description").unwrap_or_default(),
        })).collect::<Vec<_>>()}).to_string(),
        Err(e) => format!("error: {e}"),
    }
}

fn row_to_finance_tool(r: &sqlx::sqlite::SqliteRow) -> Value {
    json!({
        "id": r.get::<i64, _>("id"),
        "project_id": r.get::<i64, _>("project_id"),
        "summary": r.get::<String, _>("summary"),
        "total_planned": r.get::<f64, _>("total_planned"),
        "total_actual": r.get::<f64, _>("total_actual"),
        "notes": r.get::<String, _>("notes"),
    })
}
