//! AI project documents — file and/or paste → organized markdown + HTML + publish.

use axum::{
    body::Bytes,
    extract::{Multipart, Path, State},
    http::{header, StatusCode},
    response::IntoResponse,
    Json,
};
use morphai::{extract_json_object, Message};
use pulldown_cmark::{html, Options, Parser};
use serde::Deserialize;
use serde_json::{json, Value};
use sqlx::Row;
use std::path::PathBuf;
use std::sync::Arc;

use crate::api::extract::AuthUser;
use crate::api::record_validation::{project_code_or_generated, require_project_name};
use crate::services::doc_text;
use crate::services::AppState;

type ApiResult = Result<Json<Value>, (StatusCode, Json<Value>)>;

const MAX_FILES: usize = 5;
const MAX_FILE_BYTES: usize = 8 * 1024 * 1024;
const MAX_PASTE_CHARS: usize = 100_000;

const DOC_SYSTEM: &str = r#"You organize requirements, specifications, and project descriptions into clear project documentation.

Respond with ONLY one JSON object, no markdown fences:
{
  "title": "Short project title",
  "code": "SHORT-CODE",
  "markdown": "Full organized markdown document"
}

Markdown rules:
- Structure with headings: Overview, Goals/Requirements, Scope, Deliverables, Timeline or Milestones, Risks/Open Questions (omit empty sections).
- Be faithful to the source; do not invent major facts, dates, or names.
- Prefer concise, scannable bullets and short paragraphs.
- Aim under ~1500 words."#;

fn bad_request(msg: impl std::fmt::Display) -> (StatusCode, Json<Value>) {
    (StatusCode::BAD_REQUEST, Json(json!({"error": msg.to_string()})))
}

fn server_err(msg: impl std::fmt::Display) -> (StatusCode, Json<Value>) {
    (
        StatusCode::INTERNAL_SERVER_ERROR,
        Json(json!({"error": msg.to_string()})),
    )
}

fn db_err(e: sqlx::Error) -> (StatusCode, Json<Value>) {
    server_err(e)
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

fn slugify(title: &str) -> String {
    let mut s: String = title
        .chars()
        .map(|c| {
            if c.is_ascii_alphanumeric() {
                c.to_ascii_lowercase()
            } else if c.is_whitespace() || c == '-' || c == '_' {
                '-'
            } else {
                '\0'
            }
        })
        .filter(|c| *c != '\0')
        .collect();
    while s.contains("--") {
        s = s.replace("--", "-");
    }
    s = s.trim_matches('-').to_string();
    if s.is_empty() {
        s = "project".into();
    }
    if s.len() > 60 {
        s.truncate(60);
        s = s.trim_matches('-').to_string();
    }
    s
}

fn markdown_to_html_fragment(md: &str) -> String {
    let mut opts = Options::empty();
    opts.insert(Options::ENABLE_TABLES);
    opts.insert(Options::ENABLE_STRIKETHROUGH);
    let parser = Parser::new_ext(md, opts);
    let mut out = String::new();
    html::push_html(&mut out, parser);
    out
}

fn html_escape(s: &str) -> String {
    s.replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('"', "&quot;")
}

fn build_project_html(title: &str, markdown: &str) -> String {
    let body = markdown_to_html_fragment(markdown);
    format!(
        r#"<!doctype html>
<html lang="en"><head><meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>{title}</title>
<style>
:root {{ --ink:#e8eef7; --muted:#94a3b8; --line:#1e293b; --bg:#0b1220; --card:#111827; --accent:#2dd4bf; }}
*{{box-sizing:border-box}} body{{margin:0;font-family:Georgia,"Times New Roman",serif;color:var(--ink);background:radial-gradient(1200px 600px at 10% -10%,#134e4a 0%,var(--bg) 55%);line-height:1.55}}
.wrap{{max-width:760px;margin:0 auto;padding:2rem 1.25rem 3rem}}
.meta{{font:12px/1.4 system-ui,sans-serif;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin-bottom:.75rem}}
h1.page-title{{font-size:clamp(1.6rem,3vw,2.2rem);margin:0 0 1rem;line-height:1.2;color:#f8fafc}}
.prose{{font-size:1.05rem}}.prose h1,.prose h2,.prose h3{{color:#f8fafc;line-height:1.25;margin:1.35em 0 .55em}}
.prose p,.prose ul,.prose ol{{margin:.85em 0}}.prose ul,.prose ol{{padding-left:1.4em}}
.prose a{{color:var(--accent)}}.prose code{{font-family:ui-monospace,monospace;background:var(--card);padding:.1em .35em;border-radius:4px;border:1px solid var(--line)}}
.prose pre{{overflow:auto;padding:1em;border-radius:10px;background:var(--card);border:1px solid var(--line)}}
.foot{{font:11px/1.4 system-ui,sans-serif;color:var(--muted);margin-top:2rem}}
</style></head><body><div class="wrap">
<div class="meta">Project</div>
<h1 class="page-title">{title}</h1>
<div class="prose">{body}</div>
<p class="foot">Generated project document</p>
</div></body></html>"#,
        title = html_escape(title),
        body = body
    )
}

struct UploadedFile {
    name: String,
    mime: String,
    bytes: Bytes,
}

async fn unique_slug(pool: &sqlx::SqlitePool, base: &str) -> Result<String, sqlx::Error> {
    let base = slugify(base);
    for i in 0..50 {
        let candidate = if i == 0 {
            base.clone()
        } else {
            format!("{base}-{}", i + 1)
        };
        let exists: i64 = sqlx::query_scalar(
            "SELECT COUNT(*) FROM projects WHERE published_slug = ?",
        )
        .bind(&candidate)
        .fetch_one(pool)
        .await?;
        if exists == 0 {
            return Ok(candidate);
        }
    }
    Ok(format!(
        "{}-{}",
        base,
        chrono::Utc::now().timestamp() % 100_000
    ))
}

async fn insert_resource_file(
    pool: &sqlx::SqlitePool,
    org_id: i64,
    name: &str,
    source_type: &str,
    file_url: &str,
    file_name: &str,
    description: &str,
    project_id: i64,
) -> Result<i64, sqlx::Error> {
    let res = sqlx::query(
        "INSERT INTO resource_files (organization_id, name, source_type, file_url, file_name, description) VALUES (?,?,?,?,?,?)",
    )
    .bind(org_id)
    .bind(name)
    .bind(source_type)
    .bind(file_url)
    .bind(file_name)
    .bind(description)
    .execute(pool)
    .await?;
    let id = res.last_insert_rowid();
    sqlx::query(
        "INSERT OR IGNORE INTO project_resource_files (project_id, resource_file_id) VALUES (?,?)",
    )
    .bind(project_id)
    .bind(id)
    .execute(pool)
    .await?;
    Ok(id)
}

/// POST /projects/generate-document  multipart: file(s)?, paste?, title?
pub async fn generate_document(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    mut multipart: Multipart,
) -> ApiResult {
    let mut title_hint = String::new();
    let mut paste = String::new();
    let mut uploads: Vec<UploadedFile> = Vec::new();

    while let Some(field) = multipart.next_field().await.map_err(bad_request)? {
        let field_name = field.name().unwrap_or("").to_string();
        match field_name.as_str() {
            "title" => {
                title_hint = field.text().await.map_err(bad_request)?;
            }
            "paste" | "content" => {
                paste = field.text().await.map_err(bad_request)?;
            }
            "files" | "file" => {
                if uploads.len() >= MAX_FILES {
                    return Err(bad_request(format!(
                        "at most {MAX_FILES} files can be used together"
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
                        "{fname} exceeds the {} MB limit",
                        MAX_FILE_BYTES / 1024 / 1024
                    )));
                }
                let mime = if declared.is_empty() {
                    doc_text::mime_from_name(&fname)
                } else {
                    declared
                };
                if doc_text::classify(&fname, &mime) == doc_text::FileKind::Unsupported {
                    return Err(bad_request(format!(
                        "{fname} is not a supported file type; accepted: {}",
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

    let paste = paste.trim().to_string();
    let title_hint = title_hint.trim().to_string();
    if uploads.is_empty() && paste.is_empty() {
        return Err(bad_request(
            "provide at least one source: file upload or pasted content",
        ));
    }
    if paste.chars().count() > MAX_PASTE_CHARS {
        return Err(bad_request("paste content too long"));
    }

    let upload_dir = PathBuf::from("uploads").join(auth.org_id.to_string());
    tokio::fs::create_dir_all(&upload_dir)
        .await
        .map_err(server_err)?;

    let mut readable: Vec<(String, String)> = Vec::new();
    let mut summary_bits: Vec<String> = Vec::new();
    let mut stored_uploads: Vec<(String, String, String, usize)> = Vec::new();
    // (original_name, stored_name, url, size)

    for up in &uploads {
        let stored = format!(
            "pd_{}_{}",
            chrono::Utc::now().timestamp_millis(),
            safe_filename(&up.name)
        );
        tokio::fs::write(upload_dir.join(&stored), &up.bytes)
            .await
            .map_err(server_err)?;
        let url = format!("/api/v1/uploads/{}/{}", auth.org_id, stored);
        let (excerpt, err) = match doc_text::extract(&up.name, &up.mime, &up.bytes).await {
            Ok(text) => {
                let (capped, _) = doc_text::truncate_marked(&text, doc_text::MAX_PER_FILE_CHARS);
                (capped, String::new())
            }
            Err(e) => (String::new(), e),
        };
        if !err.is_empty() {
            return Err(bad_request(format!("{}: {err}", up.name)));
        }
        if excerpt.trim().is_empty() {
            return Err(bad_request(format!(
                "{} has no extractable text",
                up.name
            )));
        }
        readable.push((up.name.clone(), excerpt));
        summary_bits.push(format!("file:{}", up.name));
        stored_uploads.push((up.name.clone(), stored, url, up.bytes.len()));
    }

    if !paste.is_empty() {
        let (capped, _) = doc_text::truncate_marked(&paste, doc_text::MAX_PER_FILE_CHARS);
        readable.push(("Pasted content".into(), capped));
        summary_bits.push("paste".into());
    }

    if readable.is_empty() {
        return Err(bad_request("sources produced no usable text"));
    }

    let mut user_prompt = String::new();
    if !title_hint.is_empty() {
        user_prompt.push_str(&format!("Preferred title (use unless clearly wrong): {title_hint}\n\n"));
    }
    user_prompt.push_str(&format!(
        "## Source material ({} parts)\n",
        readable.len()
    ));
    user_prompt.push_str(&doc_text::combine_sections(&readable));

    let (title, code_raw, markdown) = if let Some(ai) = state.ai.as_ref().as_ref() {
        let messages = vec![
            Message {
                role: "system".to_string(),
                content: DOC_SYSTEM.to_string(),
            },
            Message {
                role: "user".to_string(),
                content: user_prompt,
            },
        ];
        let raw = ai.chat_completion(&messages).await.map_err(|e| {
            (
                StatusCode::BAD_GATEWAY,
                Json(json!({"error": format!("AI failed: {e}")})),
            )
        })?;
        let obj = extract_json_object(&raw)
            .and_then(|s| serde_json::from_str::<Value>(&s).ok())
            .ok_or_else(|| {
                (
                    StatusCode::BAD_GATEWAY,
                    Json(json!({"error": "AI did not return JSON for the project document", "raw": raw})),
                )
            })?;
        let mut title = obj
            .get("title")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .trim()
            .to_string();
        if !title_hint.is_empty() {
            title = title_hint.clone();
        }
        if title.is_empty() {
            title = "Untitled project".into();
        }
        let markdown = obj
            .get("markdown")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .trim()
            .to_string();
        if markdown.is_empty() {
            return Err((
                StatusCode::BAD_GATEWAY,
                Json(json!({"error": "AI returned empty markdown"})),
            ));
        }
        let code_raw = obj
            .get("code")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .trim()
            .to_string();
        (title, code_raw, markdown)
    } else {
        let title = if title_hint.is_empty() {
            readable
                .first()
                .map(|(n, _)| n.clone())
                .unwrap_or_else(|| "Untitled project".into())
        } else {
            title_hint.clone()
        };
        let mut markdown = format!("# {title}\n\n## Overview\n\nOrganized from source material (preview mode — AI not configured).\n\n");
        for (name, text) in &readable {
            markdown.push_str(&format!("## {name}\n\n{text}\n\n"));
        }
        (title, String::new(), markdown)
    };

    let name = require_project_name(&title).map_err(|e| bad_request(e))?;
    let code = project_code_or_generated(&code_raw, &name);
    let html_out = build_project_html(&name, &markdown);
    let summary = summary_bits.join(", ");

    let res = sqlx::query(
        "INSERT INTO projects (organization_id, code, name, status, description, markdown_content, html_content, source_summary)
         VALUES (?,?,?,?,?,?,?,?)",
    )
    .bind(auth.org_id)
    .bind(&code)
    .bind(&name)
    .bind("planning")
    .bind(&summary)
    .bind(&markdown)
    .bind(&html_out)
    .bind(&summary)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;
    let project_id = res.last_insert_rowid();

    for (orig, _stored, url, _size) in &stored_uploads {
        let _ = insert_resource_file(
            &state.pool,
            auth.org_id,
            orig,
            "upload",
            url,
            orig,
            "Source file for AI project document",
            project_id,
        )
        .await;
    }

    if !paste.is_empty() {
        let paste_name = format!(
            "paste-{}.md",
            chrono::Utc::now().format("%Y%m%d-%H%M%S")
        );
        let stored = format!(
            "pd_paste_{}_{}",
            chrono::Utc::now().timestamp_millis(),
            safe_filename(&paste_name)
        );
        tokio::fs::write(upload_dir.join(&stored), paste.as_bytes())
            .await
            .map_err(server_err)?;
        let url = format!("/api/v1/uploads/{}/{}", auth.org_id, stored);
        let _ = insert_resource_file(
            &state.pool,
            auth.org_id,
            &paste_name,
            "upload",
            &url,
            &paste_name,
            "Pasted content used to generate project document",
            project_id,
        )
        .await;
    }

    let row = sqlx::query("SELECT * FROM projects WHERE id = ? AND organization_id = ?")
        .bind(project_id)
        .bind(auth.org_id)
        .fetch_one(&state.pool)
        .await
        .map_err(db_err)?;

    Ok(Json(json!({
        "project": {
            "id": row.get::<i64, _>("id"),
            "code": row.get::<String, _>("code"),
            "name": row.get::<String, _>("name"),
            "markdown_content": row.try_get::<String, _>("markdown_content").unwrap_or_default(),
            "html_content": row.try_get::<String, _>("html_content").unwrap_or_default(),
            "source_summary": row.try_get::<String, _>("source_summary").unwrap_or_default(),
            "status": row.get::<String, _>("status"),
            "description": row.get::<String, _>("description"),
        }
    })))
}

#[derive(Deserialize)]
pub struct PublishBody {
    #[serde(default)]
    pub slug: Option<String>,
}

/// POST /projects/:id/publish
pub async fn publish_project(
    State(state): State<Arc<AppState>>,
    AuthUser(auth): AuthUser,
    Path(id): Path<i64>,
    Json(body): Json<PublishBody>,
) -> ApiResult {
    let row = sqlx::query("SELECT * FROM projects WHERE id = ? AND organization_id = ?")
        .bind(id)
        .bind(auth.org_id)
        .fetch_optional(&state.pool)
        .await
        .map_err(db_err)?
        .ok_or((StatusCode::NOT_FOUND, Json(json!({"error": "project not found"}))))?;

    let markdown = row
        .try_get::<String, _>("markdown_content")
        .unwrap_or_default();
    let mut html_content = row.try_get::<String, _>("html_content").unwrap_or_default();
    let name = row.get::<String, _>("name");
    if markdown.trim().is_empty() && html_content.trim().is_empty() {
        return Err(bad_request("project has no content to publish"));
    }
    if html_content.trim().is_empty() {
        html_content = build_project_html(&name, &markdown);
    }

    let existing = row
        .try_get::<Option<String>, _>("published_slug")
        .ok()
        .flatten()
        .unwrap_or_default();
    let slug = if let Some(s) = body.slug.filter(|s| !s.trim().is_empty()) {
        slugify(s.trim())
    } else if !existing.is_empty() {
        existing
    } else {
        unique_slug(&state.pool, &name)
            .await
            .map_err(db_err)?
    };
    let path = format!("/api/v1/public/projects/{slug}");

    sqlx::query(
        "UPDATE projects SET published_slug = ?, published_path = ?, html_content = ?, updated_at = datetime('now') WHERE id = ? AND organization_id = ?",
    )
    .bind(&slug)
    .bind(&path)
    .bind(&html_content)
    .bind(id)
    .bind(auth.org_id)
    .execute(&state.pool)
    .await
    .map_err(db_err)?;

    Ok(Json(json!({
        "id": id,
        "published_slug": slug,
        "published_path": path,
        "published_url": path,
        "html_content": html_content,
        "markdown_content": markdown,
        "name": name,
    })))
}

/// GET /public/projects/:slug — no auth
pub async fn serve_public_project(
    State(state): State<Arc<AppState>>,
    Path(slug): Path<String>,
) -> Result<impl IntoResponse, (StatusCode, Json<Value>)> {
    let row = sqlx::query(
        "SELECT name, markdown_content, html_content FROM projects WHERE published_slug = ? LIMIT 1",
    )
    .bind(slug.trim())
    .fetch_optional(&state.pool)
    .await
    .map_err(db_err)?
    .ok_or((
        StatusCode::NOT_FOUND,
        Json(json!({"error": "published project not found"})),
    ))?;

    let name = row.get::<String, _>("name");
    let markdown = row
        .try_get::<String, _>("markdown_content")
        .unwrap_or_default();
    let mut html_content = row.try_get::<String, _>("html_content").unwrap_or_default();
    if html_content.trim().is_empty() {
        if markdown.trim().is_empty() {
            return Err((
                StatusCode::NOT_FOUND,
                Json(json!({"error": "published project has no content"})),
            ));
        }
        html_content = build_project_html(&name, &markdown);
    }

    Ok((
        [(header::CONTENT_TYPE, "text/html; charset=utf-8")],
        html_content,
    ))
}
