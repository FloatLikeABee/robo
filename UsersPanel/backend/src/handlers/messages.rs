use axum::extract::{Path, Query, State};
use axum::Json;
use chrono::Utc;
use serde::Deserialize;
use serde_json::{json, Value};
use std::collections::HashSet;
use uuid::Uuid;

use crate::error::{AppError, AppResult};
use crate::extractors::JwtClaims;
use crate::models::{CreateThreadBody, MarkReadBody, MessageItem, SendMessageBody};
use crate::permissions_resolve::resolve_permission_names;
use crate::public_channel::{is_public_channel, PUBLIC_CHANNEL_THREAD_ID};
use crate::state::AppState;

const PERMISSION_INBOX_MESSAGE: &str = "inbox_message";

#[derive(Debug, Deserialize)]
pub struct MessagesQuery {
    pub limit: Option<u32>,
    pub offset: Option<u32>,
}

async fn require_messaging_access(state: &AppState, user_id: &str, roles: &[String]) -> AppResult<()> {
    let perms = resolve_permission_names(&state.pool, user_id, roles).await;
    if perms.iter().any(|p| p == PERMISSION_INBOX_MESSAGE) {
        return Ok(());
    };
    Err(AppError::Forbidden)
}

async fn require_member(state: &AppState, thread_id: &str, user_id: &str) -> AppResult<()> {
    let exists: Option<(String,)> = sqlx::query_as(
        "SELECT thread_id FROM msg_thread_members WHERE thread_id = ? AND user_id = ?",
    )
    .bind(thread_id)
    .bind(user_id)
    .fetch_optional(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;
    if exists.is_none() {
        return Err(AppError::Forbidden);
    }
    Ok(())
}

pub async fn list_inbox(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<Value>> {
    require_messaging_access(&state, &claims.sub, &claims.roles).await?;

    let rows: Vec<(String, Option<String>, i8, i8, String)> = sqlx::query_as(
        r#"SELECT t.id, t.title, t.is_group, t.is_public_channel, t.updated_at
           FROM msg_threads t
           INNER JOIN msg_thread_members tm ON tm.thread_id = t.id
           WHERE tm.user_id = ?
           ORDER BY t.is_public_channel DESC, t.updated_at DESC"#,
    )
    .bind(&claims.sub)
    .fetch_all(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;

    let mut threads = Vec::new();
    for (thread_id, title, is_group, is_public_channel, updated_at) in rows {
        let members: Vec<(String,)> = sqlx::query_as(
            "SELECT user_id FROM msg_thread_members WHERE thread_id = ? ORDER BY user_id ASC",
        )
        .bind(&thread_id)
        .fetch_all(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;
        let member_user_ids: Vec<String> = members.into_iter().map(|(id,)| id).collect();

        let latest: Option<(String, String, String, String, String)> = sqlx::query_as(
            r#"SELECT id, sender_user_id, body, message_kind, created_at
               FROM msg_messages
               WHERE thread_id = ?
               ORDER BY created_at DESC
               LIMIT 1"#,
        )
        .bind(&thread_id)
        .fetch_optional(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;

        let unread: Option<(i64,)> = sqlx::query_as(
            r#"SELECT COUNT(*)
               FROM msg_deliveries d
               INNER JOIN msg_messages m ON m.id = d.message_id
               LEFT JOIN msg_message_reads r
                 ON r.message_id = d.message_id AND r.user_id = d.recipient_user_id
               WHERE d.recipient_user_id = ?
                 AND m.thread_id = ?
                 AND r.message_id IS NULL"#,
        )
        .bind(&claims.sub)
        .bind(&thread_id)
        .fetch_optional(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;

        let latest_message = latest.map(|(id, sender_user_id, body, message_kind, created_at)| {
            MessageItem {
                id,
                thread_id: thread_id.clone(),
                sender_user_id,
                body,
                message_kind,
                created_at,
            }
        });

        let display_title = if is_public_channel == 1 {
            "Public Channel".to_string()
        } else {
            title.unwrap_or_default()
        };

        threads.push(json!({
            "thread_id": thread_id,
            "title": display_title,
            "is_group": is_group == 1,
            "is_public_channel": is_public_channel == 1,
            "member_user_ids": member_user_ids,
            "latest_message": latest_message,
            "unread_count": unread.map(|(n,)| n).unwrap_or(0),
            "updated_at": updated_at,
        }));
    }

    Ok(Json(json!({ "threads": threads })))
}

pub async fn list_message_users(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<Value>> {
    require_messaging_access(&state, &claims.sub, &claims.roles).await?;
    let rows: Vec<(String, String, String)> = sqlx::query_as(
        "SELECT id, username, email FROM plat_users ORDER BY username ASC",
    )
    .fetch_all(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;
    let users: Vec<Value> = rows
        .into_iter()
        .map(|(id, username, email)| json!({ "id": id, "username": username, "email": email }))
        .collect();
    Ok(Json(json!({ "users": users })))
}

pub async fn list_thread_messages(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(thread_id): Path<String>,
    Query(q): Query<MessagesQuery>,
) -> AppResult<Json<Value>> {
    require_messaging_access(&state, &claims.sub, &claims.roles).await?;
    require_member(&state, &thread_id, &claims.sub).await?;

    let limit = q.limit.unwrap_or(50).clamp(1, 200);
    let offset = q.offset.unwrap_or(0);
    let rows: Vec<(String, String, String, String, String)> = sqlx::query_as(
        r#"SELECT id, sender_user_id, body, message_kind, created_at
           FROM msg_messages
           WHERE thread_id = ?
           ORDER BY created_at DESC
           LIMIT ? OFFSET ?"#,
    )
    .bind(&thread_id)
    .bind(limit)
    .bind(offset)
    .fetch_all(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;

    let messages: Vec<MessageItem> = rows
        .into_iter()
        .map(|(id, sender_user_id, body, message_kind, created_at)| MessageItem {
            id,
            thread_id: thread_id.clone(),
            sender_user_id,
            body,
            message_kind,
            created_at,
        })
        .collect();

    Ok(Json(json!({ "messages": messages, "limit": limit, "offset": offset })))
}

pub async fn create_thread(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Json(body): Json<CreateThreadBody>,
) -> AppResult<Json<Value>> {
    require_messaging_access(&state, &claims.sub, &claims.roles).await?;
    let initial_message = body
        .initial_message
        .as_deref()
        .map(str::trim)
        .filter(|s| !s.is_empty())
        .map(|s| s.to_string());

    let mut member_set: HashSet<String> = body
        .member_user_ids
        .into_iter()
        .map(|s| s.trim().to_string())
        .filter(|s| !s.is_empty())
        .collect();
    member_set.insert(claims.sub.clone());

    if member_set.is_empty() {
        return Err(AppError::BadRequest("at least one member is required".into()));
    }

    let mut tx = state.pool.begin().await.map_err(|_| AppError::Internal)?;

    for member_id in &member_set {
        let exists: Option<(String,)> = sqlx::query_as("SELECT id FROM plat_users WHERE id = ?")
            .bind(member_id)
            .fetch_optional(&mut *tx)
            .await
            .map_err(|_| AppError::Internal)?;
        if exists.is_none() {
            return Err(AppError::BadRequest(format!(
                "unknown user id: {member_id}"
            )));
        }
    }

    let thread_id = Uuid::new_v4().to_string();
    let message_id = Uuid::new_v4().to_string();
    let now = Utc::now().to_rfc3339();
    let title = body.title.as_deref().map(str::trim).filter(|s| !s.is_empty());
    let is_group = if member_set.len() > 2 { 1 } else { 0 };

    sqlx::query(
        r#"INSERT INTO msg_threads (id, title, created_by, is_group, created_at, updated_at)
           VALUES (?, ?, ?, ?, ?, ?)"#,
    )
    .bind(&thread_id)
    .bind(title)
    .bind(&claims.sub)
    .bind(is_group)
    .bind(&now)
    .bind(&now)
    .execute(&mut *tx)
    .await
    .map_err(|_| AppError::Internal)?;

    for member_id in &member_set {
        sqlx::query(
            r#"INSERT INTO msg_thread_members (thread_id, user_id, joined_at)
               VALUES (?, ?, ?)"#,
        )
        .bind(&thread_id)
        .bind(member_id)
        .bind(&now)
        .execute(&mut *tx)
        .await
        .map_err(|_| AppError::Internal)?;
    }

    let created_message_id = if let Some(initial_text) = initial_message {
        sqlx::query(
            r#"INSERT INTO msg_messages (id, thread_id, sender_user_id, body, message_kind, created_at)
               VALUES (?, ?, ?, ?, 'human', ?)"#,
        )
        .bind(&message_id)
        .bind(&thread_id)
        .bind(&claims.sub)
        .bind(initial_text)
        .bind(&now)
        .execute(&mut *tx)
        .await
        .map_err(|_| AppError::Internal)?;

        for member_id in &member_set {
            if member_id == &claims.sub {
                sqlx::query(
                    r#"INSERT INTO msg_message_reads (message_id, user_id, read_at)
                       VALUES (?, ?, ?)"#,
                )
                .bind(&message_id)
                .bind(member_id)
                .bind(&now)
                .execute(&mut *tx)
                .await
                .map_err(|_| AppError::Internal)?;
                continue;
            }
            sqlx::query(
                r#"INSERT INTO msg_deliveries (message_id, recipient_user_id, delivered_at)
                   VALUES (?, ?, ?)"#,
            )
            .bind(&message_id)
            .bind(member_id)
            .bind(&now)
            .execute(&mut *tx)
            .await
            .map_err(|_| AppError::Internal)?;
        }

        Some(message_id)
    } else {
        None
    };

    tx.commit().await.map_err(|_| AppError::Internal)?;
    Ok(Json(json!({ "thread_id": thread_id, "message_id": created_message_id })))
}

pub async fn send_message(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(thread_id): Path<String>,
    Json(body): Json<SendMessageBody>,
) -> AppResult<Json<Value>> {
    require_messaging_access(&state, &claims.sub, &claims.roles).await?;
    require_member(&state, &thread_id, &claims.sub).await?;
    let text = body.body.trim();
    if text.is_empty() {
        return Err(AppError::BadRequest("body is required".into()));
    }

    let message_id = Uuid::new_v4().to_string();
    let now = Utc::now().to_rfc3339();
    let mut tx = state.pool.begin().await.map_err(|_| AppError::Internal)?;
    sqlx::query(
        r#"INSERT INTO msg_messages (id, thread_id, sender_user_id, body, message_kind, created_at)
           VALUES (?, ?, ?, ?, 'human', ?)"#,
    )
    .bind(&message_id)
    .bind(&thread_id)
    .bind(&claims.sub)
    .bind(text)
    .bind(&now)
    .execute(&mut *tx)
    .await
    .map_err(|_| AppError::Internal)?;

    let members: Vec<(String,)> =
        sqlx::query_as("SELECT user_id FROM msg_thread_members WHERE thread_id = ?")
            .bind(&thread_id)
            .fetch_all(&mut *tx)
            .await
            .map_err(|_| AppError::Internal)?;

    for (member_id,) in members {
        if member_id == claims.sub {
            sqlx::query(
                r#"INSERT INTO msg_message_reads (message_id, user_id, read_at)
                   VALUES (?, ?, ?)"#,
            )
            .bind(&message_id)
            .bind(&member_id)
            .bind(&now)
            .execute(&mut *tx)
            .await
            .map_err(|_| AppError::Internal)?;
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
        .map_err(|_| AppError::Internal)?;
    }

    sqlx::query("UPDATE msg_threads SET updated_at = ? WHERE id = ?")
        .bind(&now)
        .bind(&thread_id)
        .execute(&mut *tx)
        .await
        .map_err(|_| AppError::Internal)?;
    tx.commit().await.map_err(|_| AppError::Internal)?;

    Ok(Json(json!({
        "message": {
            "id": message_id,
            "thread_id": thread_id,
            "sender_user_id": claims.sub,
            "body": text,
            "message_kind": "human",
            "created_at": now
        }
    })))
}

pub async fn mark_thread_read(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(thread_id): Path<String>,
    Json(body): Json<MarkReadBody>,
) -> AppResult<Json<Value>> {
    require_messaging_access(&state, &claims.sub, &claims.roles).await?;
    require_member(&state, &thread_id, &claims.sub).await?;
    let now = Utc::now().to_rfc3339();

    let mut query = String::from(
        r#"INSERT IGNORE INTO msg_message_reads (message_id, user_id, read_at)
           SELECT d.message_id, ?, ?
           FROM msg_deliveries d
           INNER JOIN msg_messages m ON m.id = d.message_id
           LEFT JOIN msg_message_reads r
             ON r.message_id = d.message_id AND r.user_id = d.recipient_user_id
           WHERE d.recipient_user_id = ?
             AND m.thread_id = ?
             AND r.message_id IS NULL"#,
    );

    let mut cutoff: Option<String> = None;
    if let Some(last_message_id) = body.last_message_id.as_deref() {
        let c: Option<(String,)> =
            sqlx::query_as("SELECT created_at FROM msg_messages WHERE id = ? AND thread_id = ?")
                .bind(last_message_id)
                .bind(&thread_id)
                .fetch_optional(&state.pool)
                .await
                .map_err(|_| AppError::Internal)?;
        if let Some((created_at,)) = c {
            query.push_str(" AND m.created_at <= ?");
            cutoff = Some(created_at);
        }
    }

    let mut q = sqlx::query(&query)
        .bind(&claims.sub)
        .bind(&now)
        .bind(&claims.sub)
        .bind(&thread_id);
    if let Some(c) = cutoff {
        q = q.bind(c);
    }
    q.execute(&state.pool).await.map_err(|_| AppError::Internal)?;

    Ok(Json(json!({ "message": "thread marked read" })))
}

pub async fn delete_thread(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(thread_id): Path<String>,
) -> AppResult<Json<Value>> {
    require_messaging_access(&state, &claims.sub, &claims.roles).await?;
    let thread_id = thread_id.trim().to_string();
    if thread_id.is_empty() {
        return Err(AppError::BadRequest("thread_id is required".into()));
    }
    require_member(&state, &thread_id, &claims.sub).await?;
    if thread_id == PUBLIC_CHANNEL_THREAD_ID
        || is_public_channel(&state.pool, &thread_id)
            .await
            .map_err(|_| AppError::Internal)?
    {
        return Err(AppError::BadRequest(
            "the public channel cannot be deleted".into(),
        ));
    }

    let res = sqlx::query("DELETE FROM msg_threads WHERE id = ?")
        .bind(&thread_id)
        .execute(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;
    if res.rows_affected() == 0 {
        return Err(AppError::NotFound);
    }
    Ok(Json(json!({ "message": "thread deleted" })))
}

pub async fn unread_count(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
) -> AppResult<Json<Value>> {
    require_messaging_access(&state, &claims.sub, &claims.roles).await?;
    let row: Option<(i64,)> = sqlx::query_as(
        r#"SELECT COUNT(*)
           FROM msg_deliveries d
           LEFT JOIN msg_message_reads r
             ON r.message_id = d.message_id AND r.user_id = d.recipient_user_id
           WHERE d.recipient_user_id = ? AND r.message_id IS NULL"#,
    )
    .bind(&claims.sub)
    .fetch_optional(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;
    Ok(Json(json!({ "unread_count": row.map(|(n,)| n).unwrap_or(0) })))
}
