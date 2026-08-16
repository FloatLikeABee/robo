//! Document verification sessions — upload files, AI review, saved reports.

use axum::{
    body::Bytes,
    extract::{Multipart, Path, Query, State},
    http::StatusCode,
    Json,
};
use morphai::{extract_json_object, truncate_chars, Message};
use serde::Deserialize;
use serde_json::{json, Value};
use sqlx::Row;
use std::path::PathBuf;
use std::sync::Arc;

use crate::api::extract::{json_err, AuthUser};
use crate::services::doc_text;
use crate::services::AppState;

type ApiResult = Result<Json<Value>, (StatusCode, Json<Value>)>;

const MAX_FILES: usize = 5;
const MAX_FILE_BYTES: usize = 8 * 1024 * 1024;
const MAX_EXCERPT_PER_FILE: usize = 10_000;

const VERIFY_SYSTEM: &str = r#"You are Morph Engi Verification AI — a senior civil engineering compliance reviewer, contract analyst, and technical document auditor.

Given:
1) A user "verification request" describing what must be checked or proven
2) Up to five uploaded documents (filename, type, and text excerpts where available)

Your job:
- Decide whether the user's requirement CAN be verified from the supplied evidence
- Cite specific evidence per file when possible
- Flag missing information, contradictions, and risks
- Be precise and conservative: "verified" only when evidence clearly supports the requirement

Respond with ONLY one JSON object (no markdown fences):
{
  "verdict": "verified|partial|not_verified|inconclusive",
  "summary": "2-4 sentence executive summary",
  "findings": [
    {"requirement": "...", "status": "met|partial|not_met|unclear", "evidence": "quote or reference", "file": "filename", "notes": "..."}
  ],
  "gaps": ["missing item or ambiguity"],
  "recommendations": ["actionable next step"],
  "report_html": "<article>...</article>"
}

report_html rules:
- Self-contained HTML fragment only (no html/head/body tags)
- Sections: Executive Summary, Verdict (with colored badge class verdict-*), Requirements Checklist, Evidence by File, Gaps, Recommendations
- Use h2/h3, ul/li, table where helpful; class names: verdict-verified, verdict-partial, verdict-not_verified, verdict-inconclusive
- Inline styles sparingly for readability on dark UI (light text, subtle borders)
- Do not invent document content not present in excerpts"#;

#[derive(Deserialize)]
pub struct SessionListQuery {
    pub project_id: Option<i64>,
}

fn db_err(e: sqlx::Error) -> (StatusCode, Json<Value>) {
    (
        StatusCode::INTERNAL_SERVER_ERROR,
        Json(json!({"error": e.to_string()})),
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

fn mime_from_name(name: &str) -> String {
    let lower = name.to_lowercase();
    if lower.ends_with(".pdf") {
        "application/pdf".into()
    } else if lower.ends_with(".png") {
        "image/png".into()
    } else if lower.ends_with(".jpg") || lower.ends_with(".jpeg") {
        "image/jpeg".into()
    } else if lower.ends_with(".csv") {
        "text/csv".into()
    } else if lower.ends_with(".json") {
        "application/json".into()
    } else if lower.ends_with(".md") {
        "text/markdown".into()
    } else {
        "application/octet-stream".into()
    }
}

async fn extract_text_excerpt(name: &str, bytes: &[u8]) -> String {
    let lower = name.to_lowercase();

    if lower.ends_with(".pdf") {
        return match doc_text::extract_pdf(bytes.to_vec()).await {
            Ok(text) => truncate_chars(&text, MAX_EXCERPT_PER_FILE),
            Err(e) => format!("[Could not read PDF — {e}]"),
        };
    }

    let text_ext = [
        ".txt", ".md", ".csv", ".json", ".xml", ".html", ".htm", ".log", ".yaml", ".yml",
    ];
    if text_ext.iter().any(|e| lower.ends_with(e)) {
        let s = String::from_utf8_lossy(bytes);
        return truncate_chars(s.trim(), MAX_EXCERPT_PER_FILE);
    }
    if bytes.len() < 4096 {
        let s = String::from_utf8_lossy(bytes);
        let printable = s.chars().filter(|c| !c.is_control() || *c == '\n').count();
        if printable * 2 > s.len() {
            return truncate_chars(s.trim(), MAX_EXCERPT_PER_FILE);
        }
    }
    format!(
        "[Binary or non-text file — {} bytes. Filename and metadata only; content not extracted server-side.]",
        bytes.len()
    )
}

fn verdict_badge_class(verdict: &str) -> &str {
    match verdict {
        "verified" => "verdict-verified",
        "partial" => "verdict-partial",
        "not_verified" => "verdict-not_verified",
        _ => "verdict-inconclusive",
    }
}

fn wrap_fallback_html(verdict: &str, summary: &str, body: &str) -> String {
    let cls = verdict_badge_class(verdict);
    format!(
        r#"<article class="verification-report"><h2>Verification Report</h2><p class="verdict-badge {cls}"><strong>Verdict:</strong> {verdict}</p><h3>Summary</h3><p>{summary}</p><div class="report-body">{body}</div></article>"#
    )
}

fn parse_ai_report(raw: &str) -> (String, String, String, String, Value) {
    let trimmed = raw.trim();
    if let Some(obj) = extract_json_object(trimmed) {
        if let Ok(v) = serde_json::from_str::<Value>(&obj) {
            let verdict = v
                .get("verdict")
                .and_then(|x| x.as_str())
                .unwrap_or("inconclusive")
                .to_string();
            let summary = v
                .get("summary")
                .and_then(|x| x.as_str())
                .unwrap_or("")
                .to_string();
            let report_html = v
                .get("report_html")
                .and_then(|x| x.as_str())
                .unwrap_or("")
                .to_string();
            let findings = v.get("findings").cloned().unwrap_or(json!([]));
            let html = if report_html.is_empty() {
                wrap_fallback_html(&verdict, &summary, &obj)
            } else {
                format!(r#"<article class="verification-report">{report_html}</article>"#)
            };
            let md = format!(
                "## Verification: {}\n\n{}\n\n```json\n{}\n```",
                verdict, summary, obj
            );
            return (verdict, summary, html, md, findings);
        }
    }
    let summary = truncate_chars(trimmed, 500);
    let html = wrap_fallback_html("inconclusive", &summary, &format!("<pre>{}</pre>", html_escape(trimmed)));
    let md = trimmed.to_string();
    (
        "inconclusive".into(),
        summary,
        html,
        md,
        json!([]),
    )
}

fn html_escape(s: &str) -> String {
    s.replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
}

pub async fn list_verification_sessions(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Query(q): Query<SessionListQuery>,
) -> ApiResult {
    let rows = if let Some(pid) = q.project_id.filter(|&p| p > 0) {
        sqlx::query(
            "SELECT id, project_id, title, verify_prompt, status, verdict, summary, created_at, updated_at
             FROM verification_sessions
             WHERE organization_id = ? AND (project_id = ? OR project_id IS NULL)
             ORDER BY created_at DESC LIMIT 100",
        )
        .bind(auth.org_id)
        .bind(pid)
        .fetch_all(&state.pool)
        .await
    } else {
        sqlx::query(
            "SELECT id, project_id, title, verify_prompt, status, verdict, summary, created_at, updated_at
             FROM verification_sessions WHERE organization_id = ? ORDER BY created_at DESC LIMIT 100",
        )
        .bind(auth.org_id)
        .fetch_all(&state.pool)
        .await
    }
    .map_err(db_err)?;

    let sessions: Vec<Value> = rows
        .iter()
        .map(|r| {
            json!({
                "id": r.get::<i64, _>("id"),
                "project_id": r.try_get::<i64, _>("project_id").ok(),
                "title": r.get::<String, _>("title"),
                "verify_prompt": r.get::<String, _>("verify_prompt"),
                "status": r.get::<String, _>("status"),
                "verdict": r.get::<String, _>("verdict"),
                "summary": r.get::<String, _>("summary"),
                "created_at": r.get::<String, _>("created_at"),
                "updated_at": r.get::<String, _>("updated_at"),
            })
        })
        .collect();

    Ok(Json(json!({ "sessions": sessions })))
}

pub async fn get_verification_session(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Path(id): Path<i64>,
) -> ApiResult {
    let row = sqlx::query(
        "SELECT * FROM verification_sessions WHERE id = ? AND organization_id = ?",
    )
    .bind(id)
    .bind(auth.org_id)
    .fetch_optional(&state.pool)
    .await
    .map_err(db_err)?
    .ok_or(json_err("session not found"))?;

    let files = sqlx::query(
        "SELECT id, original_name, stored_name, mime_type, size_bytes, excerpt, sort_order, created_at
         FROM verification_files WHERE session_id = ? ORDER BY sort_order",
    )
    .bind(id)
    .fetch_all(&state.pool)
    .await
    .map_err(db_err)?;

    let file_list: Vec<Value> = files
        .iter()
        .map(|f| {
            json!({
                "id": f.get::<i64, _>("id"),
                "original_name": f.get::<String, _>("original_name"),
                "stored_name": f.get::<String, _>("stored_name"),
                "mime_type": f.get::<String, _>("mime_type"),
                "size_bytes": f.get::<i64, _>("size_bytes"),
                "excerpt": f.get::<String, _>("excerpt"),
                "sort_order": f.get::<i64, _>("sort_order"),
                "file_url": format!("/api/v1/uploads/{}/{}", auth.org_id, f.get::<String, _>("stored_name")),
            })
        })
        .collect();

    Ok(Json(json!({
        "session": {
            "id": row.get::<i64, _>("id"),
            "project_id": row.try_get::<i64, _>("project_id").ok(),
            "title": row.get::<String, _>("title"),
            "verify_prompt": row.get::<String, _>("verify_prompt"),
            "status": row.get::<String, _>("status"),
            "verdict": row.get::<String, _>("verdict"),
            "summary": row.get::<String, _>("summary"),
            "result_html": row.get::<String, _>("result_html"),
            "result_markdown": row.get::<String, _>("result_markdown"),
            "findings_json": row.get::<String, _>("findings_json"),
            "created_at": row.get::<String, _>("created_at"),
            "updated_at": row.get::<String, _>("updated_at"),
        },
        "files": file_list,
    })))
}

struct UploadedFile {
    name: String,
    bytes: Bytes,
}

pub async fn run_verification(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    mut multipart: Multipart,
) -> ApiResult {
    let mut verify_prompt = String::new();
    let mut title = String::new();
    let mut project_id: Option<i64> = None;
    let mut uploads: Vec<UploadedFile> = Vec::new();

    while let Some(field) = multipart.next_field().await.map_err(|e| {
        (
            StatusCode::BAD_REQUEST,
            Json(json!({"error": e.to_string()})),
        )
    })? {
        let name = field.name().unwrap_or("").to_string();
        match name.as_str() {
            "verify_prompt" => {
                verify_prompt = field.text().await.map_err(|e| {
                    (
                        StatusCode::BAD_REQUEST,
                        Json(json!({"error": e.to_string()})),
                    )
                })?;
            }
            "title" => {
                title = field.text().await.map_err(|e| {
                    (
                        StatusCode::BAD_REQUEST,
                        Json(json!({"error": e.to_string()})),
                    )
                })?;
            }
            "project_id" => {
                let t = field.text().await.map_err(|e| {
                    (
                        StatusCode::BAD_REQUEST,
                        Json(json!({"error": e.to_string()})),
                    )
                })?;
                project_id = t.trim().parse().ok().filter(|&p| p > 0);
            }
            "files" | "file" => {
                if uploads.len() >= MAX_FILES {
                    return Err(json_err(&format!("Maximum {MAX_FILES} files allowed")));
                }
                let fname = field.file_name().unwrap_or("upload").to_string();
                let bytes = field.bytes().await.map_err(|e| {
                    (
                        StatusCode::BAD_REQUEST,
                        Json(json!({"error": e.to_string()})),
                    )
                })?;
                if bytes.len() > MAX_FILE_BYTES {
                    return Err(json_err(&format!(
                        "File {} exceeds {} MB limit",
                        fname,
                        MAX_FILE_BYTES / 1024 / 1024
                    )));
                }
                uploads.push(UploadedFile {
                    name: fname,
                    bytes,
                });
            }
            _ => {}
        }
    }

    let verify_prompt = verify_prompt.trim().to_string();
    if verify_prompt.is_empty() {
        return Err(json_err("verify_prompt is required"));
    }
    if uploads.is_empty() {
        return Err(json_err("At least one file is required"));
    }

    let title = if title.trim().is_empty() {
        truncate_chars(&verify_prompt, 80)
    } else {
        title.trim().to_string()
    };

    let ins = sqlx::query(
        "INSERT INTO verification_sessions (organization_id, project_id, title, verify_prompt, status)
         VALUES (?, ?, ?, ?, 'running')",
    )
    .bind(auth.org_id)
    .bind(project_id)
    .bind(&title)
    .bind(&verify_prompt)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;

    let session_id = ins.last_insert_rowid();
    let upload_dir = PathBuf::from("uploads").join(auth.org_id.to_string());
    tokio::fs::create_dir_all(&upload_dir).await.map_err(|e| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": e.to_string()})),
        )
    })?;

    let mut file_bundle = Vec::new();
    for (i, up) in uploads.iter().enumerate() {
        let safe = safe_filename(&up.name);
        let stored = format!("vf{}_{}_{}", session_id, chrono::Utc::now().timestamp_millis(), safe);
        let path = upload_dir.join(&stored);
        tokio::fs::write(&path, &up.bytes).await.map_err(|e| {
            (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(json!({"error": e.to_string()})),
            )
        })?;
        let excerpt = extract_text_excerpt(&up.name, &up.bytes).await;
        let mime = mime_from_name(&up.name);
        sqlx::query(
            "INSERT INTO verification_files (session_id, organization_id, original_name, stored_name, mime_type, size_bytes, excerpt, sort_order)
             VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
        )
        .bind(session_id)
        .bind(auth.org_id)
        .bind(&up.name)
        .bind(&stored)
        .bind(&mime)
        .bind(up.bytes.len() as i64)
        .bind(&excerpt)
        .bind(i as i64)
        .execute(&state.pool)
        .await
        .map_err(db_err)?;

        file_bundle.push(json!({
            "name": up.name,
            "mime_type": mime,
            "size_bytes": up.bytes.len(),
            "excerpt": excerpt,
        }));
    }

    let file_bundle_json = serde_json::to_string(&file_bundle).unwrap_or_else(|_| "[]".into());

    let Some(ai) = state.ai.as_ref().as_ref() else {
        let msg = "AI not configured — set MORPH_AI_API_KEY to run verification.";
        let _ = sqlx::query(
            "UPDATE verification_sessions SET status = 'failed', summary = ?, result_html = ?, updated_at = datetime('now') WHERE id = ?",
        )
        .bind(msg)
        .bind(format!("<p class=\"text-muted\">{}</p>", html_escape(msg)))
        .bind(session_id)
        .execute(&state.pool)
        .await;
        return Err((
            StatusCode::SERVICE_UNAVAILABLE,
            Json(json!({"error": msg, "session_id": session_id})),
        ));
    };

    let mut doc_section = String::new();
    for f in &file_bundle {
        doc_section.push_str(&format!(
            "\n\n### File: {}\nType: {}\nSize: {} bytes\n---\n{}\n",
            f.get("name").and_then(|v| v.as_str()).unwrap_or("?"),
            f.get("mime_type").and_then(|v| v.as_str()).unwrap_or("?"),
            f.get("size_bytes").and_then(|v| v.as_u64()).unwrap_or(0),
            f.get("excerpt").and_then(|v| v.as_str()).unwrap_or(""),
        ));
    }

    let user_prompt = format!(
        "## Verification request\n{}\n\n## Uploaded documents ({} files)\n{}",
        verify_prompt,
        file_bundle.len(),
        doc_section
    );

    let messages = vec![
        Message {
            role: "system".to_string(),
            content: VERIFY_SYSTEM.to_string(),
        },
        Message {
            role: "user".to_string(),
            content: user_prompt,
        },
    ];

    let ai_result = ai.chat_completion(&messages).await;
    let (status, verdict, summary, result_html, result_md, findings_json) = match ai_result {
        Ok(raw) => {
            let (verdict, summary, html, md, findings) = parse_ai_report(&raw);
            ("completed".to_string(), verdict, summary, html, md, findings.to_string())
        }
        Err(e) => {
            let msg = format!("AI verification failed: {e}");
            (
                "failed".to_string(),
                "inconclusive".to_string(),
                msg.clone(),
                wrap_fallback_html("inconclusive", &msg, ""),
                msg.clone(),
                "[]".to_string(),
            )
        }
    };

    sqlx::query(
        r#"UPDATE verification_sessions SET
            status = ?, verdict = ?, summary = ?, result_html = ?, result_markdown = ?,
            findings_json = ?, file_bundle_json = ?, updated_at = datetime('now')
         WHERE id = ?"#,
    )
    .bind(&status)
    .bind(&verdict)
    .bind(&summary)
    .bind(&result_html)
    .bind(&result_md)
    .bind(&findings_json)
    .bind(&file_bundle_json)
    .bind(session_id)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;

    Ok(Json(json!({
        "session": {
            "id": session_id,
            "title": title,
            "verify_prompt": verify_prompt,
            "project_id": project_id,
            "status": status,
            "verdict": verdict,
            "summary": summary,
            "result_html": result_html,
        },
        "files": file_bundle,
    })))
}
