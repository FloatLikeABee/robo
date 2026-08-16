use chrono::Utc;
use morphai::{Client, Config, Message};
use reqwest::header::{AUTHORIZATION, CONTENT_TYPE};
use serde_json::{json, Value};
use sqlx::MySqlPool;
use std::time::Duration;
use uuid::Uuid;

use crate::config::Config as AppConfig;
use crate::public_channel::{AI_SENDER_USER_ID, PUBLIC_CHANNEL_THREAD_ID};

const MESSAGE_KIND_AI_DIGEST: &str = "ai_digest";
const SKIP_MARKER: &str = "[[NO_DIGEST]]";

#[derive(Clone)]
pub struct MessageAiSettings {
    pub ai: Config,
    pub enabled: bool,
    pub interval_minutes: u64,
    pub max_per_day: u32,
    pub formsx_base_url: String,
    pub service_email: String,
    pub service_password: String,
}

impl MessageAiSettings {
    pub fn from_app_config(cfg: &AppConfig) -> Self {
        let enabled = std::env::var("MESSAGE_AI_DIGEST_ENABLED")
            .map(|v| v == "1" || v.eq_ignore_ascii_case("true"))
            .unwrap_or(true);

        let interval_minutes = std::env::var("MESSAGE_AI_DIGEST_INTERVAL_MINUTES")
            .ok()
            .and_then(|s| s.parse().ok())
            .unwrap_or(90)
            .clamp(30, 240);

        let max_per_day = std::env::var("MESSAGE_AI_DIGEST_MAX_PER_DAY")
            .ok()
            .and_then(|s| s.parse().ok())
            .unwrap_or(30)
            .clamp(1, 30);

        let formsx_base_url = std::env::var("FORMSX_BASE_URL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .unwrap_or_else(|| "http://127.0.0.1:29909".to_string())
            .trim()
            .trim_end_matches('/')
            .to_string();

        let service_email = std::env::var("MESSAGE_AI_SERVICE_EMAIL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .or_else(|| cfg.bootstrap_admin_email.clone())
            .unwrap_or_else(|| "admin@example.com".to_string());

        let service_password = std::env::var("MESSAGE_AI_SERVICE_PASSWORD")
            .ok()
            .filter(|s| !s.is_empty())
            .or_else(|| cfg.bootstrap_admin_password.clone())
            .unwrap_or_else(|| "AdminExample2026!".to_string());

        Self {
            ai: Config::from_message_ai_env(),
            enabled,
            interval_minutes,
            max_per_day,
            formsx_base_url,
            service_email,
            service_password,
        }
    }

    pub fn configured(&self) -> bool {
        self.ai.configured()
    }
}

pub fn spawn_digest_worker(pool: MySqlPool, settings: MessageAiSettings) {
    if !settings.enabled {
        tracing::info!("Message AI digest worker disabled (MESSAGE_AI_DIGEST_ENABLED)");
        return;
    }
    if !settings.configured() {
        tracing::warn!(
            "Message AI digest worker not started: MESSAGE_AI_API_KEY / MORPH_AI_API_KEY is not set"
        );
        return;
    }

    let interval = Duration::from_secs(settings.interval_minutes * 60);
    tracing::info!(
        interval_minutes = settings.interval_minutes,
        max_per_day = settings.max_per_day,
        formsx = %settings.formsx_base_url,
        "Message AI digest worker started"
    );

    tokio::spawn(async move {
        // Stagger first run so startup isn't blocked.
        tokio::time::sleep(Duration::from_secs(45)).await;
        loop {
            if let Err(e) = run_digest_once(&pool, &settings).await {
                tracing::warn!(error = %e, "Message AI digest run failed");
            }
            tokio::time::sleep(interval).await;
        }
    });
}

async fn run_digest_once(pool: &MySqlPool, settings: &MessageAiSettings) -> Result<(), String> {
    if !can_post_digest_today(pool, settings.max_per_day).await? {
        tracing::debug!("Message AI digest skipped: daily cap reached");
        return Ok(());
    }

    let token = formsx_login(settings).await?;
    let _ = formsx_mcp_call(settings, &token, "sync_system_documents", json!({})).await;

    let state = load_digest_state(pool).await?;
    let now = Utc::now().to_rfc3339();
    update_last_run(pool, &now).await?;

    let is_first_run = state.last_events_at.is_none() && state.last_docs_sync_at.is_none();
    let events = fetch_recent_events(settings, &token, state.last_events_at.as_deref()).await?;
    let docs = fetch_recent_docs(settings, &token, state.last_docs_sync_at.as_deref()).await?;

    if is_first_run {
        update_cursors(pool, &events, &docs, &now).await?;
        tracing::info!("Message AI digest: seeded cursors (skipped historical backlog)");
        return Ok(());
    }

    if events.is_empty() && docs.is_empty() {
        tracing::debug!("Message AI digest: no new FormsX events or docs");
        return Ok(());
    }

    let summary = generate_digest_summary(settings, &events, &docs).await?;
    let trimmed = summary.trim();
    if trimmed.is_empty() || trimmed.contains(SKIP_MARKER) {
        tracing::debug!("Message AI digest: AI decided nothing important to broadcast");
        update_cursors(pool, &events, &docs, &now).await?;
        return Ok(());
    }

    post_digest_message(pool, trimmed).await?;
    increment_digest_count(pool).await?;
    update_cursors(pool, &events, &docs, &now).await?;
    tracing::info!("Message AI digest posted to public channel");
    Ok(())
}

async fn can_post_digest_today(pool: &MySqlPool, max_per_day: u32) -> Result<bool, String> {
    let today = Utc::now().format("%Y-%m-%d").to_string();
    let row: Option<(i32, String)> = sqlx::query_as(
        "SELECT digests_today, digest_day FROM msg_ai_digest_state WHERE id = 1",
    )
    .fetch_optional(pool)
    .await
    .map_err(|e| e.to_string())?;

    let (count, day) = row.unwrap_or((0, String::new()));
    if day != today {
        return Ok(true);
    }
    Ok((count as u32) < max_per_day)
}

async fn load_digest_state(pool: &MySqlPool) -> Result<DigestState, String> {
    let row: Option<(Option<String>, Option<String>)> = sqlx::query_as(
        "SELECT last_events_at, last_docs_sync_at FROM msg_ai_digest_state WHERE id = 1",
    )
    .fetch_optional(pool)
    .await
    .map_err(|e| e.to_string())?;
    Ok(DigestState {
        last_events_at: row.as_ref().and_then(|r| r.0.clone()),
        last_docs_sync_at: row.as_ref().and_then(|r| r.1.clone()),
    })
}

struct DigestState {
    last_events_at: Option<String>,
    last_docs_sync_at: Option<String>,
}

async fn update_last_run(pool: &MySqlPool, now: &str) -> Result<(), String> {
    sqlx::query("UPDATE msg_ai_digest_state SET last_run_at = ?, updated_at = ? WHERE id = 1")
        .bind(now)
        .bind(now)
        .execute(pool)
        .await
        .map_err(|e| e.to_string())?;
    Ok(())
}

async fn update_cursors(
    pool: &MySqlPool,
    events: &[Value],
    docs: &[Value],
    now: &str,
) -> Result<(), String> {
    let latest_event = events
        .iter()
        .filter_map(|e| e.get("created_at").or_else(|| e.get("time")).and_then(|v| v.as_str()))
        .max()
        .map(|s| s.to_string());

    let latest_doc = docs
        .iter()
        .filter_map(|d| d.get("updated_at").or_else(|| d.get("synced_at")).and_then(|v| v.as_str()))
        .max()
        .map(|s| s.to_string());

    if let Some(ts) = latest_event {
        sqlx::query("UPDATE msg_ai_digest_state SET last_events_at = ? WHERE id = 1")
            .bind(ts)
            .execute(pool)
            .await
            .map_err(|e| e.to_string())?;
    }
    if let Some(ts) = latest_doc {
        sqlx::query("UPDATE msg_ai_digest_state SET last_docs_sync_at = ? WHERE id = 1")
            .bind(ts)
            .execute(pool)
            .await
            .map_err(|e| e.to_string())?;
    }
    sqlx::query("UPDATE msg_ai_digest_state SET updated_at = ? WHERE id = 1")
        .bind(now)
        .execute(pool)
        .await
        .map_err(|e| e.to_string())?;
    Ok(())
}

async fn increment_digest_count(pool: &MySqlPool) -> Result<(), String> {
    let today = Utc::now().format("%Y-%m-%d").to_string();
    let row: Option<(i32, String)> = sqlx::query_as(
        "SELECT digests_today, digest_day FROM msg_ai_digest_state WHERE id = 1",
    )
    .fetch_optional(pool)
    .await
    .map_err(|e| e.to_string())?;

    let (count, day) = row.unwrap_or((0, String::new()));
    let next = if day == today { count + 1 } else { 1 };
    let now = Utc::now().to_rfc3339();
    sqlx::query(
        "UPDATE msg_ai_digest_state SET digests_today = ?, digest_day = ?, updated_at = ? WHERE id = 1",
    )
    .bind(next)
    .bind(&today)
    .bind(&now)
    .execute(pool)
    .await
    .map_err(|e| e.to_string())?;
    Ok(())
}

async fn formsx_login(settings: &MessageAiSettings) -> Result<String, String> {
    let url = format!("{}/api/v1/auth/login", settings.formsx_base_url);
    let client = reqwest::Client::new();
    let resp = client
        .post(&url)
        .header(CONTENT_TYPE, "application/json")
        .json(&json!({
            "email": settings.service_email,
            "password": settings.service_password,
        }))
        .send()
        .await
        .map_err(|e| format!("formsx login request failed: {e}"))?;

    if !resp.status().is_success() {
        let text = resp.text().await.unwrap_or_default();
        return Err(format!("formsx login failed: {text}"));
    }
    let body: Value = resp.json().await.map_err(|e| e.to_string())?;
    body.get("token")
        .and_then(|v| v.as_str())
        .map(|s| s.to_string())
        .ok_or_else(|| "formsx login missing token".into())
}

async fn formsx_mcp_call(
    settings: &MessageAiSettings,
    token: &str,
    tool: &str,
    args: Value,
) -> Result<Value, String> {
    let url = format!("{}/api/v1/ai/mongodb-mcp/call", settings.formsx_base_url);
    let client = reqwest::Client::new();
    let resp = client
        .post(&url)
        .header(AUTHORIZATION, format!("Bearer {token}"))
        .header(CONTENT_TYPE, "application/json")
        .json(&json!({ "tool": tool, "args": args }))
        .send()
        .await
        .map_err(|e| format!("formsx mcp call failed: {e}"))?;

    if !resp.status().is_success() {
        let text = resp.text().await.unwrap_or_default();
        return Err(format!("formsx mcp {tool} failed: {text}"));
    }
    let body: Value = resp.json().await.map_err(|e| e.to_string())?;
    Ok(body.get("result").cloned().unwrap_or(body))
}

async fn fetch_recent_events(
    settings: &MessageAiSettings,
    token: &str,
    since: Option<&str>,
) -> Result<Vec<Value>, String> {
    let result = formsx_mcp_call(
        settings,
        token,
        "list_events_info",
        json!({ "page": 1, "limit": 50 }),
    )
    .await?;
    let events = result
        .get("events")
        .and_then(|v| v.as_array())
        .cloned()
        .unwrap_or_default();

    Ok(filter_since(events, since, &["created_at", "time"]))
}

async fn fetch_recent_docs(
    settings: &MessageAiSettings,
    token: &str,
    since: Option<&str>,
) -> Result<Vec<Value>, String> {
    let result = formsx_mcp_call(
        settings,
        token,
        "list_system_documents",
        json!({ "page": 1, "limit": 50 }),
    )
    .await?;
    let docs = result
        .get("documents")
        .and_then(|v| v.as_array())
        .cloned()
        .unwrap_or_default();

    Ok(filter_since(docs, since, &["updated_at", "synced_at", "created_at"]))
}

fn filter_since(items: Vec<Value>, since: Option<&str>, fields: &[&str]) -> Vec<Value> {
    let Some(since) = since else {
        return items;
    };
    items
        .into_iter()
        .filter(|item| {
            fields.iter().any(|f| {
                item.get(*f)
                    .and_then(|v| v.as_str())
                    .map(|ts| ts > since)
                    .unwrap_or(false)
            })
        })
        .collect()
}

async fn generate_digest_summary(
    settings: &MessageAiSettings,
    events: &[Value],
    docs: &[Value],
) -> Result<String, String> {
    let client = Client::new(settings.ai.clone());
    let events_text = serde_json::to_string_pretty(events).unwrap_or_else(|_| "[]".into());
    let docs_text = serde_json::to_string_pretty(docs).unwrap_or_else(|_| "[]".into());

    let system = "You are a platform operations assistant. Summarize new FormsX events and system documents for a company-wide public channel. \
Only broadcast when there is genuinely useful operational news. \
If nothing is important enough, reply with exactly [[NO_DIGEST]] and nothing else. \
When you do broadcast, write clear Markdown with a short title (##), bullet highlights, and keep under 250 words. \
Focus on what changed, who it affects, and any action needed. Do not invent facts.";

    let user = format!(
        "New FormsX events since last check:\n{events_text}\n\nNew/updated system documents:\n{docs_text}\n\n\
Produce one public-channel digest message in Markdown, or {SKIP_MARKER} if nothing warrants posting."
    );

    client
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
}

async fn post_digest_message(pool: &MySqlPool, body: &str) -> Result<(), String> {
    let message_id = Uuid::new_v4().to_string();
    let now = Utc::now().to_rfc3339();
    let mut tx = pool.begin().await.map_err(|e| e.to_string())?;

    sqlx::query(
        r#"INSERT INTO msg_messages (id, thread_id, sender_user_id, body, message_kind, created_at)
           VALUES (?, ?, ?, ?, ?, ?)"#,
    )
    .bind(&message_id)
    .bind(PUBLIC_CHANNEL_THREAD_ID)
    .bind(AI_SENDER_USER_ID)
    .bind(body)
    .bind(MESSAGE_KIND_AI_DIGEST)
    .bind(&now)
    .execute(&mut *tx)
    .await
    .map_err(|e| e.to_string())?;

    let members: Vec<(String,)> =
        sqlx::query_as("SELECT user_id FROM msg_thread_members WHERE thread_id = ?")
            .bind(PUBLIC_CHANNEL_THREAD_ID)
            .fetch_all(&mut *tx)
            .await
            .map_err(|e| e.to_string())?;

    for (member_id,) in members {
        if member_id == AI_SENDER_USER_ID {
            sqlx::query(
                r#"INSERT INTO msg_message_reads (message_id, user_id, read_at)
                   VALUES (?, ?, ?)"#,
            )
            .bind(&message_id)
            .bind(&member_id)
            .bind(&now)
            .execute(&mut *tx)
            .await
            .map_err(|e| e.to_string())?;
            continue;
        }
        sqlx::query(
            r#"INSERT INTO msg_deliveries (message_id, recipient_user_id, delivered_at)
               VALUES (?, ?, ?)"#,
        )
        .bind(&message_id)
        .bind(&member_id)
        .bind(&now)
        .execute(&mut *tx)
        .await
        .map_err(|e| e.to_string())?;
    }

    sqlx::query("UPDATE msg_threads SET updated_at = ? WHERE id = ?")
        .bind(&now)
        .bind(PUBLIC_CHANNEL_THREAD_ID)
        .execute(&mut *tx)
        .await
        .map_err(|e| e.to_string())?;

    tx.commit().await.map_err(|e| e.to_string())?;
    Ok(())
}
