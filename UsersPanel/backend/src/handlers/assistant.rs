use axum::extract::State;
use axum::Json;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

use crate::error::AppResult;
use crate::extractors::JwtClaims;
use crate::handlers::admin::create_user_and_morph_profile;
use crate::importcol::entities::{all_specs, spec_for, EntityKind};
use crate::handlers::assistant_llm::respond_general_llm;
use crate::handlers::roles_admin::insert_role;
use crate::models::{CreateAdminUserBody, RoleRow, UserRow};
use crate::permissions::{require_role, ROLE_ADMIN};
use crate::state::AppState;
use crate::error::AppError;

#[derive(Debug, Deserialize)]
pub struct AssistantMessage {
    pub role: String,
    pub content: String,
}

#[derive(Debug, Clone, Default, Deserialize, Serialize)]
pub struct AssistantState {
    #[serde(default)]
    pub intent: String,
    #[serde(default)]
    pub fields: HashMap<String, String>,
}

#[derive(Debug, Deserialize)]
pub struct AssistantRequest {
    pub messages: Vec<AssistantMessage>,
    #[serde(default)]
    pub state: AssistantState,
}

#[derive(Debug, Serialize)]
pub struct AssistantResponse {
    pub assistant_message: String,
    pub intent: String,
    pub missing_fields: Vec<String>,
    pub state: AssistantState,
    pub completed: bool,
}

pub async fn chat(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Json(mut req): Json<AssistantRequest>,
) -> AppResult<Json<AssistantResponse>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(crate::error::AppError::Forbidden);
    }

    let user_msg = latest_user_message(&req.messages);
    reconcile_assistant_intent(&mut req.state, &user_msg);
    if req.state.intent.is_empty() {
        req.state.intent = detect_intent(&user_msg);
    }
    update_fields(&mut req.state.fields, &user_msg);

    match req.state.intent.as_str() {
        "list_users" => Ok(Json(AssistantResponse {
            assistant_message: list_users_markdown(&state).await?,
            intent: "list_users".to_string(),
            missing_fields: vec![],
            state: AssistantState::default(),
            completed: true,
        })),
        "list_roles" => Ok(Json(AssistantResponse {
            assistant_message: list_roles_markdown(&state).await?,
            intent: "list_roles".to_string(),
            missing_fields: vec![],
            state: AssistantState::default(),
            completed: true,
        })),
        "create_user" => {
            let missing = missing_required(
                &req.state.fields,
                &["email", "username", "password", "last_name"],
            );
            if !missing.is_empty() {
                return Ok(Json(AssistantResponse {
                    assistant_message: format!(
                        "I can create the user in Users Panel and MorphData. Please provide:\n\n{}",
                        missing
                            .iter()
                            .map(|f| format!("- `{f}`"))
                            .collect::<Vec<_>>()
                            .join("\n")
                    ),
                    intent: "create_user".to_string(),
                    missing_fields: missing,
                    state: req.state,
                    completed: false,
                }));
            }

            let create_body = match to_create_body(&req.state.fields) {
                Ok(body) => body,
                Err(msg) => {
                    return Ok(Json(AssistantResponse {
                        assistant_message: msg,
                        intent: "create_user".to_string(),
                        missing_fields: vec![],
                        state: req.state,
                        completed: false,
                    }));
                }
            };

            match create_user_and_morph_profile(&state, create_body).await {
                Ok(created) => Ok(Json(AssistantResponse {
                    assistant_message: format!(
                        "✓ **User created**\n\n- **Users Panel ID:** {}\n- **MorphData UserID:** {}",
                        created.user_id, created.morph_user_id
                    ),
                    intent: "create_user".to_string(),
                    missing_fields: vec![],
                    state: AssistantState::default(),
                    completed: true,
                })),
                Err(err) => Ok(Json(AssistantResponse {
                    assistant_message: format!(
                        "**Error:** I could not create the user yet: {err}\n\nCorrect the details and try again."
                    ),
                    intent: "create_user".to_string(),
                    missing_fields: vec![],
                    state: req.state,
                    completed: false,
                })),
            }
        }
        "create_role" => {
            let missing = missing_required(&req.state.fields, &["role_name"]);
            if !missing.is_empty() {
                return Ok(Json(AssistantResponse {
                    assistant_message: format!(
                        "I can create the role. Please provide:\n\n{}",
                        missing
                            .iter()
                            .map(|f| format!("- `{f}`"))
                            .collect::<Vec<_>>()
                            .join("\n")
                    ),
                    intent: "create_role".to_string(),
                    missing_fields: missing,
                    state: req.state,
                    completed: false,
                }));
            }
            let role_name = req
                .state
                .fields
                .get("role_name")
                .map(|v| v.trim().to_string())
                .unwrap_or_default();
            let description = optional_trim(req.state.fields.get("description"));
            match insert_role(&state.pool, &role_name, description.as_deref()).await {
                Ok((id, name)) => Ok(Json(AssistantResponse {
                    assistant_message: format!("✓ **Role created** — **{name}** (id: `{id}`)"),
                    intent: "create_role".to_string(),
                    missing_fields: vec![],
                    state: AssistantState::default(),
                    completed: true,
                })),
                Err(AppError::Conflict(_)) => Ok(Json(AssistantResponse {
                    assistant_message: format!(
                        "A role named \"{role_name}\" already exists. Choose a different name."
                    ),
                    intent: "create_role".to_string(),
                    missing_fields: vec![],
                    state: req.state,
                    completed: false,
                })),
                Err(err) => Ok(Json(AssistantResponse {
                    assistant_message: format!("I could not create the role: {err}"),
                    intent: "create_role".to_string(),
                    missing_fields: vec![],
                    state: req.state,
                    completed: false,
                })),
            }
        }
        "data_collector_help" => Ok(Json(AssistantResponse {
            assistant_message: data_collector_help_message(&user_msg),
            intent: "data_collector_help".to_string(),
            missing_fields: vec![],
            state: req.state,
            completed: true,
        })),
        _ => {
            let user_msg = latest_user_message(&req.messages);
            match respond_general_llm(&state, &req.messages, &req.state, &user_msg).await {
                Ok(reply) => Ok(Json(AssistantResponse {
                    assistant_message: reply,
                    intent: "general".to_string(),
                    missing_fields: vec![],
                    state: req.state,
                    completed: true,
                })),
                Err(_) => Ok(Json(AssistantResponse {
                    assistant_message: "I can help with **UsersPanel** users, roles, permissions, and **Data Collector** imports.\n\nTry:\n- `create user email: …`\n- `create role role_name: …`\n- `how do I import participants CSV?`\n\nSet `MORPH_AI_API_KEY` to enable the full LLM assistant.".to_string(),
                    intent: "general".to_string(),
                    missing_fields: vec![],
                    state: req.state,
                    completed: true,
                })),
            }
        }
    }
}

fn latest_user_message(messages: &[AssistantMessage]) -> String {
    messages
        .iter()
        .rev()
        .find(|m| m.role.eq_ignore_ascii_case("user"))
        .map(|m| m.content.trim().to_string())
        .unwrap_or_default()
}

fn detect_intent(msg: &str) -> String {
    let low = msg.to_lowercase();
    if low.contains("list users") || low.contains("show users") {
        "list_users".to_string()
    } else if low.contains("list roles") || low.contains("show roles") {
        "list_roles".to_string()
    } else if low.contains("create user") || low.contains("new user") {
        "create_user".to_string()
    } else if low.contains("create role") || low.contains("new role") {
        "create_role".to_string()
    } else if low.contains("data collector")
        || low.contains("import csv")
        || low.contains("import excel")
        || low.contains("import json")
        || low.contains("import template")
        || low.contains("morphdata import")
        || (low.contains("import") && (low.contains("member") || low.contains("facility") || low.contains("participant")))
    {
        "data_collector_help".to_string()
    } else {
        "general".to_string()
    }
}

fn reconcile_assistant_intent(st: &mut AssistantState, msg: &str) {
    let low = msg.to_lowercase();
    if st.fields.is_empty() && st.intent.is_empty() {
        return;
    }
    match () {
        _ if low.contains("list users") || low.contains("show users") => {
            st.intent = "list_users".to_string();
            st.fields.clear();
        }
        _ if low.contains("list roles") || low.contains("show roles") => {
            st.intent = "list_roles".to_string();
            st.fields.clear();
        }
        _ if low.contains("create user") || low.contains("new user") => {
            st.intent = "create_user".to_string();
        }
        _ if low.contains("create role") || low.contains("new role") => {
            st.intent = "create_role".to_string();
        }
        _ if st.intent == "create_user" || st.intent == "create_role" => {
            if low.contains("list ") || low.contains("show ") {
                st.intent = detect_intent(msg);
                st.fields.clear();
            }
        }
        _ => {}
    }
}

async fn list_users_markdown(state: &AppState) -> AppResult<String> {
    let total: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM plat_users")
        .fetch_one(&state.pool)
        .await
        .map_err(|_| AppError::Internal)?;

    let rows: Vec<UserRow> = sqlx::query_as(
        "SELECT * FROM plat_users ORDER BY created_at DESC LIMIT 25",
    )
    .fetch_all(&state.pool)
    .await
    .map_err(|_| AppError::Internal)?;

    if rows.is_empty() {
        return Ok("_No users yet._".to_string());
    }

    let mut items = Vec::with_capacity(rows.len());
    for row in rows {
        let roles = row.roles_vec().map_err(|_| AppError::Internal)?;
        let permissions = row.permissions_vec().map_err(|_| AppError::Internal)?;
        let label = row.username.trim();
        let line = if label.is_empty() {
            format!("`{}`", row.email)
        } else {
            format!("**{}** — `{}`", label, row.email)
        };
        let access_text = if crate::permissions::is_admin(&roles) {
            " · admin".to_string()
        } else if permissions.is_empty() {
            String::new()
        } else {
            format!(" · apps: {}", permissions.join(", "))
        };
        items.push(format!("{line}{access_text}"));
    }

    let intro = if total as usize > items.len() {
        format!("**{total} user(s)** _(showing {n} most recent)_", n = items.len())
    } else {
        format!("**{total} user(s)**")
    };

    Ok(bullet_list_markdown(&intro, &items))
}

async fn list_roles_markdown(state: &AppState) -> AppResult<String> {
    let rows: Vec<RoleRow> =
        sqlx::query_as("SELECT * FROM plat_roles ORDER BY name ASC")
            .fetch_all(&state.pool)
            .await
            .map_err(|_| AppError::Internal)?;

    if rows.is_empty() {
        return Ok("_No roles yet._".to_string());
    }

    let items: Vec<String> = rows
        .iter()
        .map(|row| {
            let desc = row.description.as_deref().unwrap_or("").trim();
            if desc.is_empty() {
                format!("**{}**", row.name)
            } else {
                format!("**{}** — {}", row.name, desc)
            }
        })
        .collect();

    Ok(bullet_list_markdown(
        &format!("**{} role(s)**", items.len()),
        &items,
    ))
}

fn bullet_list_markdown(intro: &str, items: &[String]) -> String {
    if items.is_empty() {
        return intro.to_string();
    }
    let mut out = String::new();
    if !intro.trim().is_empty() {
        out.push_str(intro.trim());
        out.push_str("\n\n");
    }
    for item in items {
        out.push_str("- ");
        out.push_str(item.trim());
        out.push('\n');
    }
    out.trim_end().to_string()
}

fn data_collector_help_message(msg: &str) -> String {
    let low = msg.to_lowercase();
    for kind in EntityKind::all() {
        if low.contains(kind.as_str())
            || low.contains(kind.label().to_lowercase().as_str())
        {
            let spec = spec_for(*kind);
            return format!(
                "**Data Collector — {}**\n\n{}\n\n**Required columns:** none\n\n**Record title:** add a `title` column for an overall title per record.\n\n**Template headers:** `{}`\n\n**CSV example (one row):**\n```\n{}\n```\n\n**JSON example:**\n```json\n{}\n```\n\nOpen **Data Collector** in the nav to validate and run a background import. If your file headers match the template exactly, import runs without AI.",
                spec.label,
                spec.description,
                spec.template_headers.join(", "),
                spec.csv_example.lines().nth(1).unwrap_or(""),
                spec.json_example.trim()
            );
        }
    }
    let mut lines = vec![
        "**Data Collector** imports MorphData records from CSV, Excel (.xlsx), or JSON.".to_string(),
        "Supported entities:".to_string(),
    ];
    for spec in all_specs() {
        lines.push(format!(
            "- **{}** (`{}`) — required: none",
            spec.label,
            spec.kind.as_str()
        ));
    }
    lines.push(
        "\nAsk e.g. *How do I import members?* for templates and examples.".to_string(),
    );
    lines.join("\n")
}

fn update_fields(fields: &mut HashMap<String, String>, msg: &str) {
    for token in [
        "email",
        "password",
        "role_name",
        "description",
        "username",
        "last_name",
        "first_name",
        "login_id",
        "phone",
        "title",
        "roles",
        "administrator",
        "is_verified",
    ] {
        if let Some(value) = extract_value(msg, token) {
            fields.insert(token.to_string(), value);
        }
    }

    if let Some((k, v)) = extract_alias(msg, "last name") {
        fields.insert(k, v);
    }
    if let Some((k, v)) = extract_alias(msg, "first name") {
        fields.insert(k, v);
    }
    if let Some((k, v)) = extract_alias(msg, "login id") {
        fields.insert(k, v);
    }
    if let Some((k, v)) = extract_alias(msg, "verified") {
        fields.insert(k, v);
    }
    if let Some((k, v)) = extract_alias(msg, "admin") {
        fields.insert(k, v);
    }
}

fn extract_value(msg: &str, token: &str) -> Option<String> {
    let low = msg.to_lowercase();
    let idx = low.find(token)?;
    let value = &msg[idx + token.len()..];
    let split_chars: &[char] = if token == "roles" {
        &[';', '\n']
    } else {
        &[',', ';', '\n']
    };
    let cleaned = value
        .trim_start_matches(&[':', '=', '-', ' '][..])
        .split(split_chars)
        .next()
        .unwrap_or_default()
        .trim()
        .to_string();
    if cleaned.is_empty() {
        None
    } else {
        Some(cleaned)
    }
}

fn extract_alias(msg: &str, token: &str) -> Option<(String, String)> {
    let mapped = match token {
        "last name" => "last_name",
        "first name" => "first_name",
        "login id" => "login_id",
        "verified" => "is_verified",
        "admin" => "administrator",
        _ => return None,
    };
    extract_value(msg, token).map(|v| (mapped.to_string(), v))
}

fn missing_required(fields: &HashMap<String, String>, required: &[&str]) -> Vec<String> {
    required
        .iter()
        .filter_map(|key| {
            let ok = fields
                .get(*key)
                .map(|v| !v.trim().is_empty())
                .unwrap_or(false);
            if ok {
                None
            } else {
                Some((*key).to_string())
            }
        })
        .collect()
}

fn parse_bool(value: Option<&String>) -> Option<bool> {
    let v = value?.trim().to_lowercase();
    if ["1", "true", "yes", "y"].contains(&v.as_str()) {
        Some(true)
    } else if ["0", "false", "no", "n"].contains(&v.as_str()) {
        Some(false)
    } else {
        None
    }
}


fn to_create_body(fields: &HashMap<String, String>) -> Result<CreateAdminUserBody, String> {
    let email = fields
        .get("email")
        .map(|v| v.trim().to_string())
        .ok_or_else(|| "email is missing".to_string())?;
    let username = fields
        .get("username")
        .map(|v| v.trim().to_string())
        .ok_or_else(|| "username is missing".to_string())?;
    let password = fields
        .get("password")
        .map(|v| v.trim().to_string())
        .ok_or_else(|| "password is missing".to_string())?;
    let last_name = fields
        .get("last_name")
        .map(|v| v.trim().to_string())
        .ok_or_else(|| "last_name is missing".to_string())?;
    let is_admin = fields
        .get("is_admin")
        .and_then(|v| parse_bool(Some(v)))
        .unwrap_or(false);
    let permissions = fields
        .get("permissions")
        .map(|v| {
            v.split(',')
                .map(|p| p.trim().to_string())
                .filter(|p| !p.is_empty())
                .collect()
        })
        .filter(|p: &Vec<String>| !p.is_empty());

    Ok(CreateAdminUserBody {
        email,
        username: username.clone(),
        password,
        is_verified: parse_bool(fields.get("is_verified")),
        is_admin: Some(is_admin),
        permissions,
        login_id: optional_trim(fields.get("login_id")),
        first_name: optional_trim(fields.get("first_name")),
        last_name: Some(last_name),
        phone: optional_trim(fields.get("phone")),
        title: optional_trim(fields.get("title")),
        administrator: parse_bool(fields.get("administrator")),
    })
}

fn optional_trim(value: Option<&String>) -> Option<String> {
    value.and_then(|v| {
        let trimmed = v.trim();
        if trimmed.is_empty() {
            None
        } else {
            Some(trimmed.to_string())
        }
    })
}
