//! Docs document HTML publish (Board replacement) with optional AI analysis.

use axum::{
    Json,
    extract::{Path, State},
    http::{StatusCode, header},
    response::{Html, IntoResponse, Response},
};
use morphai::{Client, Config, Message};
use serde::Deserialize;
use serde_json::{Value, json};
use uuid::Uuid;

use crate::api::error_response;
use crate::db::Database;
use crate::db::published_data_page_repository::slugify_publish_name;
use crate::services::AppState;

#[derive(Deserialize)]
pub struct PublishDocsBody {
    pub title: String,
    #[serde(default)]
    pub content: String,
    #[serde(default)]
    pub doc_id: Option<String>,
    /// Optional user prompt for AI analysis of the document.
    #[serde(default)]
    pub analysis_prompt: Option<String>,
}

async fn unique_slug(db: &Database, base: &str) -> Result<String, sqlx::Error> {
    let base = if base.trim().is_empty() {
        "doc"
    } else {
        base
    };
    let mut candidate = base.to_string();
    for i in 2..=5000 {
        let exists = match db {
            Database::Sqlite(pool) => {
                let row: Option<(i32,)> =
                    sqlx::query_as("SELECT 1 FROM docs_publishes WHERE slug = ? LIMIT 1")
                        .bind(&candidate)
                        .fetch_optional(pool)
                        .await?;
                row.is_some()
            }
            Database::Postgres(pool) => {
                let row: Option<(i32,)> =
                    sqlx::query_as("SELECT 1 FROM docs_publishes WHERE slug = $1 LIMIT 1")
                        .bind(&candidate)
                        .fetch_optional(pool)
                        .await?;
                row.is_some()
            }
        };
        if !exists {
            return Ok(candidate);
        }
        candidate = format!("{base}-{i}");
    }
    Err(sqlx::Error::Configuration(
        "could not find available docs publish slug".into(),
    ))
}

fn escape_html(s: &str) -> String {
    s.replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('"', "&quot;")
}

fn build_html(title: &str, content: &str, analysis: Option<&str>) -> String {
    let analysis_block = match analysis {
        Some(a) if !a.trim().is_empty() => format!(
            r#"<section class="analysis"><h2>AI analysis</h2><div class="analysis-body">{}</div></section>"#,
            escape_html(a).replace('\n', "<br/>")
        ),
        _ => String::new(),
    };
    format!(
        r#"<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>{title}</title>
<style>
  :root {{
    --bg: #13171f; --card: #1a2030; --text: #e9ecf1; --muted: #a3adbe; --accent: #6ea8fe; --border: rgba(255,255,255,.08);
  }}
  body {{ margin:0; font-family: system-ui, -apple-system, Segoe UI, Roboto, sans-serif; background:var(--bg); color:var(--text); line-height:1.55; padding:28px; }}
  main {{ max-width: 720px; margin: 0 auto; }}
  h1 {{ font-size: 1.75rem; margin: 0 0 16px; }}
  .doc {{ background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 20px; white-space: pre-wrap; }}
  .analysis {{ margin-top: 28px; background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 20px; }}
  .analysis h2 {{ margin: 0 0 12px; color: var(--accent); font-size: 1.1rem; }}
  .analysis-body {{ color: var(--muted); }}
  footer {{ margin-top: 24px; color: var(--muted); font-size: 12px; }}
</style>
</head>
<body>
<main>
  <h1>{title}</h1>
  <article class="doc">{content}</article>
  {analysis_block}
  <footer>Published from DataX Docs</footer>
</main>
</body>
</html>"#,
        title = escape_html(title),
        content = escape_html(content).replace('\n', "<br/>"),
        analysis_block = analysis_block,
    )
}

async fn run_analysis(title: &str, content: &str, prompt: &str) -> Result<String, String> {
    let cfg = Config::from_env();
    if !cfg.configured() {
        return Err("MorphAI is not configured (set MORPH_AI_API_KEY)".into());
    }
    let client = Client::new(cfg);
    let system = "You are Docs analysis AI. Analyze the document for the user's prompt. Reply in clear markdown prose. Do not invent facts not supported by the document.";
    let user = format!(
        "Document title: {title}\n\nDocument content:\n{content}\n\nUser analysis prompt:\n{prompt}"
    );
    let reply = client
        .chat_completion(&[
            Message {
                role: "system".into(),
                content: system.into(),
            },
            Message {
                role: "user".into(),
                content: user,
            },
        ])
        .await
        .map_err(|e| e.to_string())?;
    Ok(reply.trim().to_string())
}

pub async fn publish(
    State(state): State<AppState>,
    Json(body): Json<PublishDocsBody>,
) -> Result<Json<Value>, (StatusCode, String)> {
    let title = body.title.trim();
    if title.is_empty() {
        return Err((StatusCode::BAD_REQUEST, "title is required".into()));
    }
    let content = body.content.trim();
    if content.is_empty() {
        return Err((
            StatusCode::BAD_REQUEST,
            "content is required to publish a document".into(),
        ));
    }

    let prompt = body
        .analysis_prompt
        .as_deref()
        .map(str::trim)
        .filter(|s| !s.is_empty());

    let analysis = if let Some(p) = prompt {
        Some(
            run_analysis(title, content, p)
                .await
                .map_err(|e| (StatusCode::BAD_GATEWAY, e))?,
        )
    } else {
        None
    };

    let html = build_html(title, content, analysis.as_deref());
    let base = slugify_publish_name(title);
    let slug = unique_slug(&state.db_pool, &base)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
    let id = Uuid::new_v4();
    let doc_id = body.doc_id.clone().unwrap_or_default();

    match &state.db_pool {
        Database::Sqlite(pool) => {
            sqlx::query(
                r#"INSERT INTO docs_publishes (id, doc_id, title, slug, html_content, analysis_prompt, created_at, updated_at)
                   VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))"#,
            )
            .bind(id.to_string())
            .bind(&doc_id)
            .bind(title)
            .bind(&slug)
            .bind(&html)
            .bind(prompt.unwrap_or(""))
            .execute(pool)
            .await
            .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
        }
        Database::Postgres(pool) => {
            sqlx::query(
                r#"INSERT INTO docs_publishes (id, doc_id, title, slug, html_content, analysis_prompt, created_at, updated_at)
                   VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())"#,
            )
            .bind(id)
            .bind(&doc_id)
            .bind(title)
            .bind(&slug)
            .bind(&html)
            .bind(prompt.unwrap_or(""))
            .execute(pool)
            .await
            .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?;
        }
    }

    Ok(Json(json!({
        "id": id,
        "slug": slug,
        "title": title,
        "public_path": format!("/public/docs/{slug}"),
        "has_analysis": analysis.is_some(),
    })))
}

pub async fn list_publishes(State(state): State<AppState>) -> Result<Json<Value>, (StatusCode, String)> {
    let rows: Vec<(String, String, String, String)> = match &state.db_pool {
        Database::Sqlite(pool) => sqlx::query_as(
            "SELECT id, title, slug, created_at FROM docs_publishes ORDER BY created_at DESC LIMIT 100",
        )
        .fetch_all(pool)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?,
        Database::Postgres(pool) => sqlx::query_as(
            "SELECT id::text, title, slug, created_at::text FROM docs_publishes ORDER BY created_at DESC LIMIT 100",
        )
        .fetch_all(pool)
        .await
        .map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()))?,
    };
    let items: Vec<Value> = rows
        .into_iter()
        .map(|(id, title, slug, created_at)| {
            json!({
                "id": id,
                "title": title,
                "slug": slug,
                "created_at": created_at,
                "public_path": format!("/public/docs/{slug}"),
            })
        })
        .collect();
    Ok(Json(json!({ "publishes": items })))
}

pub async fn serve_published(Path(slug): Path<String>, State(state): State<AppState>) -> Response {
    let html: Option<String> = match &state.db_pool {
        Database::Sqlite(pool) => sqlx::query_scalar(
            "SELECT html_content FROM docs_publishes WHERE slug = ? LIMIT 1",
        )
        .bind(&slug)
        .fetch_optional(pool)
        .await
        .ok()
        .flatten(),
        Database::Postgres(pool) => sqlx::query_scalar(
            "SELECT html_content FROM docs_publishes WHERE slug = $1 LIMIT 1",
        )
        .bind(&slug)
        .fetch_optional(pool)
        .await
        .ok()
        .flatten(),
    };
    match html {
        Some(h) => Html(h).into_response(),
        None => error_response(StatusCode::NOT_FOUND, "Published doc not found").into_response(),
    }
}
