//! AI project import — read description files, propose projects, create on confirm.
//!
//! Analyze and confirm are deliberately separate: analysis only stores a draft,
//! and nothing reaches the domain tables until the user confirms an edited plan.
//!
//! Flow-log entries have no project column, so a created entry is associated with
//! its project by carrying the project code as a tag.

use axum::{
    body::Bytes,
    extract::{Multipart, Path, State},
    http::StatusCode,
    Json,
};
use morphai::{extract_json_object, Message};
use serde::Deserialize;
use serde_json::{json, Value};
use sqlx::Row;
use std::path::PathBuf;
use std::sync::Arc;

use crate::api::extract::AuthUser;
use crate::api::record_validation::{
    project_code_or_generated, require_person_name, require_project_name, validate_flow_entry,
    validate_site_log,
};
use crate::services::doc_text;
use crate::services::AppState;

type ApiResult = Result<Json<Value>, (StatusCode, Json<Value>)>;

const MAX_FILES: usize = 3;
const MAX_FILE_BYTES: usize = 8 * 1024 * 1024;

const IMPORT_SYSTEM: &str = r#"You are Morph Engi Project Import AI. You read construction and engineering description files and propose projects to create.

The files may contain project briefs, site logs, people or contractor lists, and money movements (income and expenses).

Respond with ONLY one JSON object, no markdown fences:
{
  "projects": [
    {
      "name": "Project name",
      "code": "SHORT-CODE",
      "client": "",
      "location": "",
      "status": "planning|active|on_hold|completed",
      "start_date": "YYYY-MM-DD",
      "end_date": "YYYY-MM-DD",
      "budget_total": 0,
      "description": "What this project covers, from the source",
      "site_logs": [
        {"log_date": "YYYY-MM-DD", "weather": "", "crew_count": 0, "summary": "What happened on site", "issues": ""}
      ],
      "people": [
        {"name": "Person or company", "trade": "", "contact_name": "", "phone": "", "email": "", "status": "active", "description": ""}
      ],
      "flow_log_entries": [
        {"entry_date": "YYYY-MM-DD", "direction": "income|expense", "amount": 0, "currency": "USD", "category": "", "title": "", "notes": ""}
      ]
    }
  ],
  "notes": "Anything you could not place, or assumptions a reviewer should check"
}

Rules:
- One entry in "projects" per distinct job, site, or contract in the sources. If the sources describe several jobs, propose several projects.
- "name" is required for every project. Never emit a project without a name.
- Attach each site log, person, and money movement to the project it belongs to. Put items you cannot attribute in "notes" rather than guessing.
- Omit an optional field, or leave it empty, when the sources do not state it. Never invent dates, amounts, names, or contact details.
- "amount" must be a positive number and "direction" must be income or expense. Drop a money movement you cannot express that way and mention it in "notes".
- Dates are ISO YYYY-MM-DD. If a source gives only a month or a year, leave the date empty and say so in "notes"."#;

fn db_err(e: sqlx::Error) -> (StatusCode, Json<Value>) {
    (
        StatusCode::INTERNAL_SERVER_ERROR,
        Json(json!({"error": e.to_string()})),
    )
}

fn bad_request(msg: impl std::fmt::Display) -> (StatusCode, Json<Value>) {
    (
        StatusCode::BAD_REQUEST,
        Json(json!({"error": msg.to_string()})),
    )
}

fn server_err(msg: impl std::fmt::Display) -> (StatusCode, Json<Value>) {
    (
        StatusCode::INTERNAL_SERVER_ERROR,
        Json(json!({"error": msg.to_string()})),
    )
}

fn safe_filename(name: &str) -> String {
    name.chars()
        .map(|c| {
            if c.is_ascii_alphanumeric() || c == '.' || c == '-' || c == '_' {
                c
            } else {
                '_'
            }
        })
        .collect()
}

struct UploadedFile {
    name: String,
    mime: String,
    bytes: Bytes,
}

/// POST /project-imports/analyze
pub async fn analyze(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    mut multipart: Multipart,
) -> ApiResult {
    let mut instruction = String::new();
    let mut uploads: Vec<UploadedFile> = Vec::new();

    while let Some(field) = multipart.next_field().await.map_err(bad_request)? {
        let field_name = field.name().unwrap_or("").to_string();
        match field_name.as_str() {
            "instruction" | "prompt" => {
                instruction = field.text().await.map_err(bad_request)?;
            }
            "files" | "file" => {
                if uploads.len() >= MAX_FILES {
                    return Err(bad_request(format!(
                        "at most {MAX_FILES} files can be analyzed together"
                    )));
                }
                let fname = field.file_name().unwrap_or("upload").to_string();
                let declared = field
                    .content_type()
                    .map(|s| s.to_string())
                    .unwrap_or_default();
                let bytes = field.bytes().await.map_err(bad_request)?;
                if bytes.len() > MAX_FILE_BYTES {
                    return Err(bad_request(format!(
                        "{} exceeds the {} MB limit",
                        fname,
                        MAX_FILE_BYTES / 1024 / 1024
                    )));
                }
                let mime = if declared.is_empty() {
                    doc_text::mime_from_name(&fname)
                } else {
                    declared
                };
                // Reject unsupported types before spending anything on AI.
                if doc_text::classify(&fname, &mime) == doc_text::FileKind::Unsupported {
                    return Err(bad_request(format!(
                        "{fname} is not a supported file type; accepted file types: {}",
                        doc_text::ACCEPTED_TYPES
                    )));
                }
                uploads.push(UploadedFile {
                    name: fname,
                    mime,
                    bytes,
                });
            }
            _ => {}
        }
    }

    if uploads.is_empty() {
        return Err(bad_request("at least one file is required"));
    }
    let instruction = instruction.trim().to_string();

    let ins = sqlx::query(
        "INSERT INTO project_import_sessions (organization_id, status, instruction) VALUES (?, 'analyzing', ?)",
    )
    .bind(auth.org_id)
    .bind(&instruction)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    let session_id = ins.last_insert_rowid();

    let upload_dir = PathBuf::from("uploads").join(auth.org_id.to_string());
    tokio::fs::create_dir_all(&upload_dir)
        .await
        .map_err(server_err)?;

    let mut file_results: Vec<Value> = Vec::new();
    let mut readable: Vec<(String, String)> = Vec::new();

    for (i, up) in uploads.iter().enumerate() {
        let stored = format!(
            "pi{}_{}_{}",
            session_id,
            chrono::Utc::now().timestamp_millis(),
            safe_filename(&up.name)
        );
        tokio::fs::write(upload_dir.join(&stored), &up.bytes)
            .await
            .map_err(server_err)?;

        let (excerpt, error) = match doc_text::extract(&up.name, &up.mime, &up.bytes).await {
            Ok(text) => {
                let (capped, _) = doc_text::truncate_marked(&text, doc_text::MAX_PER_FILE_CHARS);
                (capped, String::new())
            }
            Err(e) => (String::new(), e),
        };

        sqlx::query(
            "INSERT INTO project_import_files (session_id, organization_id, original_name, stored_name, mime_type, size_bytes, excerpt, error, sort_order)
             VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
        )
        .bind(session_id)
        .bind(auth.org_id)
        .bind(&up.name)
        .bind(&stored)
        .bind(&up.mime)
        .bind(up.bytes.len() as i64)
        .bind(&excerpt)
        .bind(&error)
        .bind(i as i64)
        .execute(&state.pool)
        .await
        .map_err(db_err)?;

        if error.is_empty() {
            readable.push((up.name.clone(), excerpt.clone()));
        }
        file_results.push(json!({
            "name": up.name,
            "mime_type": up.mime,
            "size_bytes": up.bytes.len(),
            "excerpt": excerpt,
            "error": error,
        }));
    }

    let Some(ai) = state.ai.as_ref().as_ref() else {
        let msg = "AI not configured — set MORPH_AI_API_KEY to import projects from files.";
        fail_session(&state.pool, session_id, msg).await;
        return Err((
            StatusCode::SERVICE_UNAVAILABLE,
            Json(json!({"error": msg, "session_id": session_id, "files": file_results})),
        ));
    };

    if readable.is_empty() {
        let msg = "No content could be read from the uploaded files.";
        fail_session(&state.pool, session_id, msg).await;
        return Err((
            StatusCode::UNPROCESSABLE_ENTITY,
            Json(json!({"error": msg, "session_id": session_id, "files": file_results})),
        ));
    }

    let mut user_prompt = String::new();
    if !instruction.is_empty() {
        user_prompt.push_str(&format!("## Instruction from the user\n{instruction}\n\n"));
    }
    user_prompt.push_str(&format!(
        "## Uploaded description files ({} readable)",
        readable.len()
    ));
    user_prompt.push_str(&doc_text::combine_sections(&readable));

    let messages = vec![
        Message {
            role: "system".to_string(),
            content: IMPORT_SYSTEM.to_string(),
        },
        Message {
            role: "user".to_string(),
            content: user_prompt,
        },
    ];

    let raw = match ai.chat_completion(&messages).await {
        Ok(raw) => raw,
        Err(e) => {
            let msg = format!("AI analysis failed: {e}");
            fail_session(&state.pool, session_id, &msg).await;
            return Err((
                StatusCode::BAD_GATEWAY,
                Json(json!({"error": msg, "session_id": session_id, "files": file_results})),
            ));
        }
    };

    let draft = match extract_json_object(&raw).and_then(|s| serde_json::from_str::<Value>(&s).ok()) {
        Some(v) => normalize_draft(v),
        None => {
            let msg = "The AI reply could not be read as a project plan.".to_string();
            fail_session(&state.pool, session_id, &msg).await;
            return Err((
                StatusCode::BAD_GATEWAY,
                Json(json!({"error": msg, "session_id": session_id, "raw": raw, "files": file_results})),
            ));
        }
    };

    let draft_json = draft.to_string();
    sqlx::query(
        "UPDATE project_import_sessions SET status = 'drafted', draft_json = ?, error = '', updated_at = datetime('now') WHERE id = ?",
    )
    .bind(&draft_json)
    .bind(session_id)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;

    Ok(Json(json!({
        "session": {
            "id": session_id,
            "status": "drafted",
            "instruction": instruction,
            "draft": draft,
        },
        "files": file_results,
    })))
}

async fn fail_session(pool: &sqlx::SqlitePool, session_id: i64, msg: &str) {
    let _ = sqlx::query(
        "UPDATE project_import_sessions SET status = 'failed', error = ?, updated_at = datetime('now') WHERE id = ?",
    )
    .bind(msg)
    .bind(session_id)
    .execute(pool)
    .await;
}

/// Fill in the arrays the frontend expects so it never has to guess a shape.
fn normalize_draft(v: Value) -> Value {
    let projects = v
        .get("projects")
        .and_then(|p| p.as_array())
        .cloned()
        .unwrap_or_default();

    let projects: Vec<Value> = projects
        .into_iter()
        .map(|mut p| {
            if let Some(obj) = p.as_object_mut() {
                for key in ["site_logs", "people", "flow_log_entries"] {
                    if !obj.get(key).map(|x| x.is_array()).unwrap_or(false) {
                        obj.insert(key.into(), json!([]));
                    }
                }
                if !obj.contains_key("code") {
                    obj.insert("code".into(), json!(""));
                }
            }
            p
        })
        .collect();

    json!({
        "projects": projects,
        "notes": v.get("notes").cloned().unwrap_or(json!("")),
    })
}

/// GET /project-imports
pub async fn list_sessions(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
) -> ApiResult {
    let rows = sqlx::query(
        "SELECT id, status, instruction, error, created_project_ids, created_at, updated_at
         FROM project_import_sessions WHERE organization_id = ? ORDER BY created_at DESC LIMIT 100",
    )
    .bind(auth.org_id)
    .fetch_all(&state.pool)
    .await
    .map_err(db_err)?;

    let sessions: Vec<Value> = rows
        .iter()
        .map(|r| {
            json!({
                "id": r.get::<i64, _>("id"),
                "status": r.get::<String, _>("status"),
                "instruction": r.get::<String, _>("instruction"),
                "error": r.get::<String, _>("error"),
                "created_project_ids": parse_json_or(r.get::<String, _>("created_project_ids"), json!([])),
                "created_at": r.get::<String, _>("created_at"),
                "updated_at": r.get::<String, _>("updated_at"),
            })
        })
        .collect();

    Ok(Json(json!({ "sessions": sessions })))
}

/// GET /project-imports/:id
pub async fn get_session(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Path(id): Path<i64>,
) -> ApiResult {
    let (row, files) = load_session(&state.pool, auth.org_id, id).await?;

    Ok(Json(json!({
        "session": {
            "id": row.get::<i64, _>("id"),
            "status": row.get::<String, _>("status"),
            "instruction": row.get::<String, _>("instruction"),
            "error": row.get::<String, _>("error"),
            "draft": parse_json_or(row.get::<String, _>("draft_json"), json!({"projects": []})),
            "created_project_ids": parse_json_or(row.get::<String, _>("created_project_ids"), json!([])),
            "created_at": row.get::<String, _>("created_at"),
            "updated_at": row.get::<String, _>("updated_at"),
        },
        "files": files,
    })))
}

/// Load a session scoped to the caller's organization, so another org's session
/// is indistinguishable from a missing one.
async fn load_session(
    pool: &sqlx::SqlitePool,
    org_id: i64,
    id: i64,
) -> Result<(sqlx::sqlite::SqliteRow, Vec<Value>), (StatusCode, Json<Value>)> {
    let row = sqlx::query("SELECT * FROM project_import_sessions WHERE id = ? AND organization_id = ?")
        .bind(id)
        .bind(org_id)
        .fetch_optional(pool)
        .await
        .map_err(db_err)?
        .ok_or_else(|| {
            (
                StatusCode::NOT_FOUND,
                Json(json!({"error": "import session not found"})),
            )
        })?;

    let file_rows = sqlx::query(
        "SELECT original_name, mime_type, size_bytes, excerpt, error, sort_order
         FROM project_import_files WHERE session_id = ? ORDER BY sort_order",
    )
    .bind(id)
    .fetch_all(pool)
    .await
    .map_err(db_err)?;

    let files = file_rows
        .iter()
        .map(|f| {
            json!({
                "name": f.get::<String, _>("original_name"),
                "mime_type": f.get::<String, _>("mime_type"),
                "size_bytes": f.get::<i64, _>("size_bytes"),
                "excerpt": f.get::<String, _>("excerpt"),
                "error": f.get::<String, _>("error"),
            })
        })
        .collect();

    Ok((row, files))
}

fn parse_json_or(raw: String, fallback: Value) -> Value {
    serde_json::from_str(&raw).unwrap_or(fallback)
}

// --- Confirm ---

fn default_true() -> bool {
    true
}

#[derive(Deserialize)]
pub struct ConfirmBody {
    #[serde(default)]
    pub projects: Vec<ProposedProject>,
}

#[derive(Deserialize)]
pub struct ProposedProject {
    #[serde(default)]
    pub name: String,
    #[serde(default)]
    pub code: String,
    #[serde(default)]
    pub client: String,
    #[serde(default)]
    pub location: String,
    #[serde(default)]
    pub status: String,
    #[serde(default)]
    pub start_date: Option<String>,
    #[serde(default)]
    pub end_date: Option<String>,
    #[serde(default)]
    pub budget_total: f64,
    #[serde(default)]
    pub description: String,
    #[serde(default = "default_true")]
    pub include: bool,
    #[serde(default)]
    pub site_logs: Vec<ProposedSiteLog>,
    #[serde(default)]
    pub people: Vec<ProposedPerson>,
    #[serde(default)]
    pub flow_log_entries: Vec<ProposedFlowEntry>,
}

#[derive(Deserialize)]
pub struct ProposedSiteLog {
    #[serde(default)]
    pub log_date: String,
    #[serde(default)]
    pub weather: String,
    #[serde(default)]
    pub crew_count: i64,
    #[serde(default)]
    pub summary: String,
    #[serde(default)]
    pub issues: String,
    #[serde(default = "default_true")]
    pub include: bool,
}

#[derive(Deserialize)]
pub struct ProposedPerson {
    #[serde(default)]
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
    pub status: String,
    #[serde(default)]
    pub description: String,
    #[serde(default = "default_true")]
    pub include: bool,
}

#[derive(Deserialize)]
pub struct ProposedFlowEntry {
    #[serde(default)]
    pub entry_date: String,
    #[serde(default)]
    pub direction: String,
    #[serde(default)]
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
    #[serde(default = "default_true")]
    pub include: bool,
}

/// POST /project-imports/:id/confirm
pub async fn confirm(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Path(id): Path<i64>,
    Json(body): Json<ConfirmBody>,
) -> ApiResult {
    let (row, _) = load_session(&state.pool, auth.org_id, id).await?;

    let status: String = row.get("status");
    if status == "completed" {
        return Err(bad_request(
            "this import was already completed; start a new import to create more projects",
        ));
    }

    let accepted: Vec<&ProposedProject> = body.projects.iter().filter(|p| p.include).collect();
    if accepted.is_empty() {
        return Err(bad_request("no projects were selected for creation"));
    }

    let mut created: Vec<Value> = Vec::new();
    let mut created_ids: Vec<i64> = Vec::new();
    let mut errors: Vec<Value> = Vec::new();

    for (pi, proposal) in accepted.iter().enumerate() {
        let name = match require_project_name(&proposal.name) {
            Ok(n) => n,
            Err(e) => {
                errors.push(json!({
                    "record": format!("projects[{pi}]"),
                    "name": proposal.name,
                    "error": e,
                }));
                continue;
            }
        };
        let code = project_code_or_generated(&proposal.code, &name);

        // One transaction per project: a project and its children land together.
        let mut tx = state.pool.begin().await.map_err(db_err)?;

        let res = sqlx::query(
            "INSERT INTO projects (organization_id, code, name, client, location, status, start_date, end_date, budget_total, progress_pct, description)
             VALUES (?,?,?,?,?,?,?,?,?,0,?)",
        )
        .bind(auth.org_id)
        .bind(&code)
        .bind(&name)
        .bind(&proposal.client)
        .bind(&proposal.location)
        .bind(if proposal.status.trim().is_empty() {
            "planning"
        } else {
            proposal.status.trim()
        })
        .bind(proposal.start_date.as_deref().filter(|s| !s.trim().is_empty()))
        .bind(proposal.end_date.as_deref().filter(|s| !s.trim().is_empty()))
        .bind(proposal.budget_total)
        .bind(&proposal.description)
        .execute(&mut *tx)
        .await;

        let project_id = match res {
            Ok(r) => r.last_insert_rowid(),
            Err(e) => {
                let _ = tx.rollback().await;
                errors.push(json!({
                    "record": format!("projects[{pi}]"),
                    "name": name,
                    "error": e.to_string(),
                }));
                continue;
            }
        };

        let mut log_count = 0usize;
        for (i, log) in proposal.site_logs.iter().enumerate().filter(|(_, l)| l.include) {
            match validate_site_log(&log.log_date, &log.summary) {
                Ok(fields) => {
                    let r = sqlx::query(
                        "INSERT INTO site_logs (organization_id, project_id, log_date, weather, crew_count, summary, issues) VALUES (?,?,?,?,?,?,?)",
                    )
                    .bind(auth.org_id)
                    .bind(project_id)
                    .bind(&fields.log_date)
                    .bind(&log.weather)
                    .bind(log.crew_count)
                    .bind(&fields.summary)
                    .bind(&log.issues)
                    .execute(&mut *tx)
                    .await;
                    match r {
                        Ok(_) => log_count += 1,
                        Err(e) => errors.push(json!({
                            "record": format!("projects[{pi}].site_logs[{i}]"),
                            "project": name,
                            "error": e.to_string(),
                        })),
                    }
                }
                Err(e) => errors.push(json!({
                    "record": format!("projects[{pi}].site_logs[{i}]"),
                    "project": name,
                    "error": e,
                })),
            }
        }

        let mut people_count = 0usize;
        for (i, person) in proposal.people.iter().enumerate().filter(|(_, p)| p.include) {
            let person_name = match require_person_name(&person.name) {
                Ok(n) => n,
                Err(e) => {
                    errors.push(json!({
                        "record": format!("projects[{pi}].people[{i}]"),
                        "project": name,
                        "error": e,
                    }));
                    continue;
                }
            };
            let r = sqlx::query(
                "INSERT INTO contractors (organization_id, name, trade, contact_name, phone, email, rating, status, description) VALUES (?,?,?,?,?,?,0,?,?)",
            )
            .bind(auth.org_id)
            .bind(&person_name)
            .bind(&person.trade)
            .bind(&person.contact_name)
            .bind(&person.phone)
            .bind(&person.email)
            .bind(if person.status.trim().is_empty() {
                "active"
            } else {
                person.status.trim()
            })
            .bind(&person.description)
            .execute(&mut *tx)
            .await;

            match r {
                Ok(ins) => {
                    let contractor_id = ins.last_insert_rowid();
                    let link = sqlx::query(
                        "INSERT OR IGNORE INTO project_contractors (project_id, contractor_id) VALUES (?, ?)",
                    )
                    .bind(project_id)
                    .bind(contractor_id)
                    .execute(&mut *tx)
                    .await;
                    match link {
                        Ok(_) => people_count += 1,
                        Err(e) => errors.push(json!({
                            "record": format!("projects[{pi}].people[{i}]"),
                            "project": name,
                            "error": e.to_string(),
                        })),
                    }
                }
                Err(e) => errors.push(json!({
                    "record": format!("projects[{pi}].people[{i}]"),
                    "project": name,
                    "error": e.to_string(),
                })),
            }
        }

        let mut flow_count = 0usize;
        for (i, entry) in proposal
            .flow_log_entries
            .iter()
            .enumerate()
            .filter(|(_, e)| e.include)
        {
            match validate_flow_entry(
                &entry.entry_date,
                &entry.direction,
                entry.amount,
                &entry.currency,
                &entry.status,
                &entry.title,
            ) {
                Ok(fields) => {
                    let tags = serde_json::to_string(&vec![code.clone()]).unwrap_or_else(|_| "[]".into());
                    let r = sqlx::query(
                        r#"INSERT INTO flow_log_entries
                           (organization_id, entry_date, direction, amount, currency, category, status, title, notes, tags, created_by)
                           VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"#,
                    )
                    .bind(auth.org_id)
                    .bind(&fields.entry_date)
                    .bind(fields.direction)
                    .bind(fields.amount)
                    .bind(&fields.currency)
                    .bind(entry.category.trim())
                    .bind(&fields.status)
                    .bind(&fields.title)
                    .bind(entry.notes.trim())
                    .bind(&tags)
                    .bind(auth.user_id)
                    .execute(&mut *tx)
                    .await;
                    match r {
                        Ok(_) => flow_count += 1,
                        Err(e) => errors.push(json!({
                            "record": format!("projects[{pi}].flow_log_entries[{i}]"),
                            "project": name,
                            "error": e.to_string(),
                        })),
                    }
                }
                Err(e) => errors.push(json!({
                    "record": format!("projects[{pi}].flow_log_entries[{i}]"),
                    "project": name,
                    "error": e,
                })),
            }
        }

        tx.commit().await.map_err(db_err)?;

        created_ids.push(project_id);
        created.push(json!({
            "id": project_id,
            "name": name,
            "code": code,
            "site_logs": log_count,
            "people": people_count,
            "flow_log_entries": flow_count,
        }));
    }

    let final_status = if created.is_empty() { "drafted" } else { "completed" };
    sqlx::query(
        "UPDATE project_import_sessions SET status = ?, created_project_ids = ?, updated_at = datetime('now') WHERE id = ? AND organization_id = ?",
    )
    .bind(final_status)
    .bind(serde_json::to_string(&created_ids).unwrap_or_else(|_| "[]".into()))
    .bind(id)
    .bind(auth.org_id)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;

    if created.is_empty() {
        return Err((
            StatusCode::UNPROCESSABLE_ENTITY,
            Json(json!({
                "error": "no projects could be created",
                "session_id": id,
                "errors": errors,
            })),
        ));
    }

    Ok(Json(json!({
        "session_id": id,
        "status": final_status,
        "created_projects": created,
        "errors": errors,
    })))
}

/// DELETE /project-imports/:id
pub async fn delete_session(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Path(id): Path<i64>,
) -> ApiResult {
    let n = sqlx::query("DELETE FROM project_import_sessions WHERE id = ? AND organization_id = ?")
        .bind(id)
        .bind(auth.org_id)
        .execute(&state.pool)
        .await
        .map_err(db_err)?
        .rows_affected();
    if n == 0 {
        return Err((
            StatusCode::NOT_FOUND,
            Json(json!({"error": "import session not found"})),
        ));
    }
    let _ = sqlx::query("DELETE FROM project_import_files WHERE session_id = ?")
        .bind(id)
        .execute(&state.pool)
        .await;
    Ok(Json(json!({"ok": true})))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::config::Settings;
    use crate::middleware::auth::AuthContext;
    use crate::services::{jwt, users_panel};
    use axum::body::Body;
    use axum::http::Request;
    use axum::routing::post;
    use axum::Router;
    use sqlx::SqlitePool;
    use tower::ServiceExt;

    const ORG: i64 = 1;
    const OTHER_ORG: i64 = 2;

    fn auth(org_id: i64) -> AuthUser {
        AuthUser(AuthContext {
            user_id: 7,
            org_id,
            role: "admin".into(),
            bearer: String::new(),
        })
    }

    async fn test_state() -> Arc<AppState> {
        let pool = SqlitePool::connect("sqlite::memory:").await.unwrap();
        crate::db::migrations::run(&pool).await.unwrap();

        // Projects reference organizations, so both orgs must exist.
        for id in [ORG, OTHER_ORG] {
            sqlx::query("INSERT OR IGNORE INTO organizations (id, name) VALUES (?, ?)")
                .bind(id)
                .bind(format!("Org {id}"))
                .execute(&pool)
                .await
                .unwrap();
        }

        let settings = Settings {
            database_url: "sqlite::memory:".into(),
            jwt_secret: "test-secret".into(),
            jwt_access_expiry_min: 60,
            app_env: "test".into(),
            app_port: 0,
            cors_origin: "*".into(),
            users_panel_base_url: "http://localhost".into(),
        };
        Arc::new(AppState {
            pool,
            jwt: jwt::JwtService::new(settings.jwt_secret.clone(), settings.jwt_access_expiry_min),
            users_panel: users_panel::UsersPanelClient::new(settings.users_panel_base_url.clone()),
            settings,
            ai: Arc::new(None),
        })
    }

    /// Insert a session directly, standing in for a completed analyze call.
    async fn seed_session(state: &AppState, org_id: i64, status: &str) -> i64 {
        sqlx::query(
            "INSERT INTO project_import_sessions (organization_id, status, instruction, draft_json) VALUES (?, ?, '', '{}')",
        )
        .bind(org_id)
        .bind(status)
        .execute(&state.pool)
        .await
        .unwrap()
        .last_insert_rowid()
    }

    /// A file part: (field name, filename, declared MIME, content).
    type FilePart<'a> = (&'a str, &'a str, &'a str, &'a str);

    fn multipart_body(fields: &[(&str, &str)], files: &[FilePart]) -> (String, Vec<u8>) {
        let boundary = "TESTBOUNDARY";
        let mut body = Vec::new();
        for (name, value) in fields {
            body.extend_from_slice(
                format!(
                    "--{boundary}\r\nContent-Disposition: form-data; name=\"{name}\"\r\n\r\n{value}\r\n"
                )
                .as_bytes(),
            );
        }
        for (name, filename, mime, content) in files {
            body.extend_from_slice(
                format!(
                    "--{boundary}\r\nContent-Disposition: form-data; name=\"{name}\"; filename=\"{filename}\"\r\nContent-Type: {mime}\r\n\r\n{content}\r\n"
                )
                .as_bytes(),
            );
        }
        body.extend_from_slice(format!("--{boundary}--\r\n").as_bytes());
        (
            format!("multipart/form-data; boundary={boundary}"),
            body,
        )
    }

    /// Post to analyze through a router, injecting an auth context.
    async fn post_analyze(
        state: Arc<AppState>,
        fields: &[(&str, &str)],
        files: &[FilePart<'_>],
    ) -> (StatusCode, Value) {
        let app = Router::new()
            .route("/analyze", post(analyze))
            .layer(axum::middleware::from_fn(|mut req: Request<Body>, next: axum::middleware::Next| async move {
                req.extensions_mut().insert(AuthContext {
                    user_id: 7,
                    org_id: ORG,
                    role: "admin".into(),
                    bearer: String::new(),
                });
                next.run(req).await
            }))
            .with_state(state);

        let (content_type, body) = multipart_body(fields, files);
        let res = app
            .oneshot(
                Request::builder()
                    .method("POST")
                    .uri("/analyze")
                    .header("content-type", content_type)
                    .body(Body::from(body))
                    .unwrap(),
            )
            .await
            .unwrap();

        let status = res.status();
        let bytes = axum::body::to_bytes(res.into_body(), 1024 * 1024)
            .await
            .unwrap();
        let value: Value = serde_json::from_slice(&bytes).unwrap_or(json!({}));
        (status, value)
    }

    #[tokio::test]
    async fn analyze_rejects_a_fourth_file() {
        let state = test_state().await;
        let (status, body) = post_analyze(
            state,
            &[],
            &[
                ("files", "a.txt", "text/plain", "one"),
                ("files", "b.txt", "text/plain", "two"),
                ("files", "c.txt", "text/plain", "three"),
                ("files", "d.txt", "text/plain", "four"),
            ],
        )
        .await;

        assert_eq!(status, StatusCode::BAD_REQUEST);
        let msg = body["error"].as_str().unwrap_or_default();
        assert!(msg.contains('3'), "error should state the limit: {msg}");
    }

    #[tokio::test]
    async fn analyze_rejects_unsupported_type_before_calling_ai() {
        let state = test_state().await;
        // AI is unconfigured here, so a 400 proves the type check ran first.
        let (status, body) =
            post_analyze(state, &[], &[("files", "plan.dwg", "application/acad", "junk")]).await;

        assert_eq!(status, StatusCode::BAD_REQUEST);
        let msg = body["error"].as_str().unwrap_or_default();
        assert!(msg.contains("plan.dwg"), "{msg}");
        assert!(msg.contains("PDF"), "{msg}");
    }

    #[tokio::test]
    async fn analyze_requires_a_file() {
        let state = test_state().await;
        let (status, body) = post_analyze(state, &[("instruction", "go")], &[]).await;

        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert!(body["error"].as_str().unwrap_or_default().contains("required"));
    }

    #[tokio::test]
    async fn analyze_reports_unconfigured_ai() {
        let state = test_state().await;
        let (status, body) = post_analyze(
            state.clone(),
            &[],
            &[("files", "brief.txt", "text/plain", "Build a bridge")],
        )
        .await;

        assert_eq!(status, StatusCode::SERVICE_UNAVAILABLE);
        assert!(body["error"]
            .as_str()
            .unwrap_or_default()
            .contains("MORPH_AI_API_KEY"));
        let session_id = body["session_id"]
            .as_i64()
            .expect("the failed session should be reported");

        // The file reached disk before the AI check, so remove what this test stored.
        let stored: Vec<String> =
            sqlx::query_scalar("SELECT stored_name FROM project_import_files WHERE session_id = ?")
                .bind(session_id)
                .fetch_all(&state.pool)
                .await
                .unwrap();
        for name in stored {
            let _ = tokio::fs::remove_file(
                PathBuf::from("uploads").join(ORG.to_string()).join(name),
            )
            .await;
        }
    }

    async fn confirm_call(
        state: &Arc<AppState>,
        org_id: i64,
        session_id: i64,
        body: Value,
    ) -> Result<Value, (StatusCode, Value)> {
        let parsed: ConfirmBody = serde_json::from_value(body).unwrap();
        match confirm(
            State(state.clone()),
            auth(org_id),
            Path(session_id),
            Json(parsed),
        )
        .await
        {
            Ok(Json(v)) => Ok(v),
            Err((s, Json(v))) => Err((s, v)),
        }
    }

    #[tokio::test]
    async fn confirm_creates_only_the_selected_projects_with_edited_values() {
        let state = test_state().await;
        let session_id = seed_session(&state, ORG, "drafted").await;

        let out = confirm_call(
            &state,
            ORG,
            session_id,
            json!({"projects": [
                {"name": "Harbour Bridge", "client": "City", "budget_total": 1000.0},
                {"name": "Skipped Depot", "include": false},
            ]}),
        )
        .await
        .expect("confirm should succeed");

        let created = out["created_projects"].as_array().unwrap();
        assert_eq!(created.len(), 1, "only the included project should be created");
        assert_eq!(created[0]["name"], json!("Harbour Bridge"));

        let rows = sqlx::query("SELECT name, client, budget_total FROM projects WHERE organization_id = ?")
            .bind(ORG)
            .fetch_all(&state.pool)
            .await
            .unwrap();
        assert_eq!(rows.len(), 1);
        assert_eq!(rows[0].get::<String, _>("name"), "Harbour Bridge");
        assert_eq!(rows[0].get::<String, _>("client"), "City");
        assert_eq!(rows[0].get::<f64, _>("budget_total"), 1000.0);
    }

    #[tokio::test]
    async fn confirm_generates_a_code_when_none_is_given() {
        let state = test_state().await;
        let session_id = seed_session(&state, ORG, "drafted").await;

        let out = confirm_call(
            &state,
            ORG,
            session_id,
            json!({"projects": [{"name": "Harbour Bridge"}]}),
        )
        .await
        .unwrap();

        assert_eq!(out["created_projects"][0]["code"], json!("HARBOUR-BRIDGE"));
    }

    #[tokio::test]
    async fn confirm_skips_excluded_children_and_creates_the_rest() {
        let state = test_state().await;
        let session_id = seed_session(&state, ORG, "drafted").await;

        let out = confirm_call(
            &state,
            ORG,
            session_id,
            json!({"projects": [{
                "name": "Harbour Bridge",
                "site_logs": [
                    {"log_date": "2026-01-02", "summary": "Poured slab"},
                    {"log_date": "2026-01-03", "summary": "Excluded log", "include": false},
                ],
                "people": [{"name": "Bo Steel Ltd", "trade": "steel"}],
                "flow_log_entries": [
                    {"entry_date": "2026-01-04", "direction": "expense", "amount": 120.5, "title": "Cement"},
                ],
            }]}),
        )
        .await
        .unwrap();

        let summary = &out["created_projects"][0];
        assert_eq!(summary["site_logs"], json!(1));
        assert_eq!(summary["people"], json!(1));
        assert_eq!(summary["flow_log_entries"], json!(1));

        let logs = sqlx::query("SELECT summary FROM site_logs WHERE organization_id = ?")
            .bind(ORG)
            .fetch_all(&state.pool)
            .await
            .unwrap();
        assert_eq!(logs.len(), 1);
        assert_eq!(logs[0].get::<String, _>("summary"), "Poured slab");

        // The person must be linked to the new project, not left dangling.
        let project_id = summary["id"].as_i64().unwrap();
        let links = sqlx::query("SELECT contractor_id FROM project_contractors WHERE project_id = ?")
            .bind(project_id)
            .fetch_all(&state.pool)
            .await
            .unwrap();
        assert_eq!(links.len(), 1);

        // Flow entries carry the project code so they can be traced back.
        let entries = sqlx::query("SELECT tags, amount FROM flow_log_entries WHERE organization_id = ?")
            .bind(ORG)
            .fetch_all(&state.pool)
            .await
            .unwrap();
        assert_eq!(entries.len(), 1);
        assert!(entries[0].get::<String, _>("tags").contains("HARBOUR-BRIDGE"));
    }

    #[tokio::test]
    async fn confirm_rejects_a_blank_project_name() {
        let state = test_state().await;
        let session_id = seed_session(&state, ORG, "drafted").await;

        let (status, body) = confirm_call(&state, ORG, session_id, json!({"projects": [{"name": "   "}]}))
            .await
            .expect_err("a blank name should not create anything");

        assert_eq!(status, StatusCode::UNPROCESSABLE_ENTITY);
        assert!(!body["errors"].as_array().unwrap().is_empty());

        let count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM projects")
            .fetch_one(&state.pool)
            .await
            .unwrap();
        assert_eq!(count, 0);
    }

    #[tokio::test]
    async fn confirm_reports_a_zero_amount_entry_but_keeps_the_project() {
        let state = test_state().await;
        let session_id = seed_session(&state, ORG, "drafted").await;

        let out = confirm_call(
            &state,
            ORG,
            session_id,
            json!({"projects": [{
                "name": "Harbour Bridge",
                "flow_log_entries": [
                    {"entry_date": "2026-01-04", "direction": "expense", "amount": 0.0, "title": "Unknown"},
                    {"entry_date": "2026-01-05", "direction": "income", "amount": 50.0, "title": "Deposit"},
                ],
            }]}),
        )
        .await
        .unwrap();

        assert_eq!(out["created_projects"][0]["flow_log_entries"], json!(1));
        let errors = out["errors"].as_array().unwrap();
        assert_eq!(errors.len(), 1);
        assert!(errors[0]["error"]
            .as_str()
            .unwrap_or_default()
            .contains("positive"));
        assert!(errors[0]["record"]
            .as_str()
            .unwrap_or_default()
            .contains("flow_log_entries[0]"));
    }

    #[tokio::test]
    async fn confirm_twice_is_rejected() {
        let state = test_state().await;
        let session_id = seed_session(&state, ORG, "drafted").await;
        let payload = json!({"projects": [{"name": "Harbour Bridge"}]});

        confirm_call(&state, ORG, session_id, payload.clone())
            .await
            .unwrap();
        let (status, body) = confirm_call(&state, ORG, session_id, payload)
            .await
            .expect_err("a second confirm must be refused");

        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert!(body["error"]
            .as_str()
            .unwrap_or_default()
            .contains("already completed"));

        let count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM projects")
            .fetch_one(&state.pool)
            .await
            .unwrap();
        assert_eq!(count, 1, "no duplicate project should exist");
    }

    #[tokio::test]
    async fn another_orgs_session_is_not_found() {
        let state = test_state().await;
        let session_id = seed_session(&state, OTHER_ORG, "drafted").await;

        let (status, _) = confirm_call(&state, ORG, session_id, json!({"projects": [{"name": "X"}]}))
            .await
            .expect_err("cross-org confirm must fail");
        assert_eq!(status, StatusCode::NOT_FOUND);

        let err = get_session(State(state.clone()), auth(ORG), Path(session_id))
            .await
            .expect_err("cross-org read must fail");
        assert_eq!(err.0, StatusCode::NOT_FOUND);
    }

    #[tokio::test]
    async fn confirm_requires_at_least_one_selected_project() {
        let state = test_state().await;
        let session_id = seed_session(&state, ORG, "drafted").await;

        let (status, body) = confirm_call(
            &state,
            ORG,
            session_id,
            json!({"projects": [{"name": "A", "include": false}]}),
        )
        .await
        .expect_err("nothing selected should be an error");

        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert!(body["error"].as_str().unwrap_or_default().contains("no projects"));
    }

    #[tokio::test]
    async fn sessions_are_listed_per_organization() {
        let state = test_state().await;
        seed_session(&state, ORG, "drafted").await;
        seed_session(&state, OTHER_ORG, "drafted").await;

        let Json(out) = list_sessions(State(state.clone()), auth(ORG)).await.unwrap();
        assert_eq!(out["sessions"].as_array().unwrap().len(), 1);
    }

    #[test]
    fn normalize_draft_fills_missing_arrays() {
        let v = json!({"projects": [{"name": "A"}]});
        let out = normalize_draft(v);
        let p = &out["projects"][0];
        assert!(p["site_logs"].is_array());
        assert!(p["people"].is_array());
        assert!(p["flow_log_entries"].is_array());
        assert_eq!(p["code"], json!(""));
    }

    #[test]
    fn normalize_draft_tolerates_a_missing_projects_key() {
        let out = normalize_draft(json!({"notes": "nothing found"}));
        assert_eq!(out["projects"].as_array().unwrap().len(), 0);
        assert_eq!(out["notes"], json!("nothing found"));
    }

    #[test]
    fn normalize_draft_keeps_supplied_children() {
        let v = json!({"projects": [{
            "name": "A",
            "site_logs": [{"summary": "poured slab"}],
            "people": [{"name": "Bo"}],
            "flow_log_entries": [{"amount": 5.0}],
        }]});
        let out = normalize_draft(v);
        let p = &out["projects"][0];
        assert_eq!(p["site_logs"].as_array().unwrap().len(), 1);
        assert_eq!(p["people"][0]["name"], json!("Bo"));
        assert_eq!(p["flow_log_entries"][0]["amount"], json!(5.0));
    }

    #[test]
    fn proposals_are_included_by_default() {
        let body: ConfirmBody =
            serde_json::from_value(json!({"projects": [{"name": "A"}]})).unwrap();
        assert!(body.projects[0].include);

        let body: ConfirmBody =
            serde_json::from_value(json!({"projects": [{"name": "A", "include": false}]})).unwrap();
        assert!(!body.projects[0].include);
    }

    #[test]
    fn nested_records_are_included_by_default() {
        let body: ConfirmBody = serde_json::from_value(json!({"projects": [{
            "name": "A",
            "site_logs": [{"summary": "s"}, {"summary": "t", "include": false}],
            "people": [{"name": "Bo"}],
            "flow_log_entries": [{"amount": 1.0, "direction": "expense", "entry_date": "2026-01-01"}],
        }]}))
        .unwrap();

        let p = &body.projects[0];
        assert!(p.site_logs[0].include);
        assert!(!p.site_logs[1].include);
        assert!(p.people[0].include);
        assert!(p.flow_log_entries[0].include);
    }

    #[test]
    fn confirm_body_tolerates_missing_optional_fields() {
        let body: ConfirmBody = serde_json::from_value(json!({"projects": [{"name": "A"}]})).unwrap();
        let p = &body.projects[0];
        assert_eq!(p.code, "");
        assert_eq!(p.budget_total, 0.0);
        assert!(p.start_date.is_none());
        assert!(p.site_logs.is_empty());
    }
}
