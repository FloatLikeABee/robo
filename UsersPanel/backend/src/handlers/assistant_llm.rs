use morphai::{Client, Config, Message};

use super::assistant::{AssistantMessage, AssistantState};

const USERS_PANEL_INSTRUCTIONS: &str = r#"You are UsersPanel AI, an expert assistant for platform administration.

You help Admins with:
- Users (create, roles assignment, verification)
- Roles and permissions in the satellite apps (TranMail, FormsX, SharpReport, Booki)
- General questions about how UsersPanel fits the platform

Safe create flows use explicit phrases:
- "create user email: … username: … password: … last_name: … roles: Admin, Employee"
- "create role role_name: … description: …"

When you need live data, respond with ONLY one JSON object (no markdown):
{"tool":"<name>","args":{...}}

Tools:
- list_roles — all platform roles
- list_users — recent users (limit int, default 25)

After TOOL_RESULT, summarize in markdown. If no tool is needed, reply in markdown only."#;

const MAX_ROUNDS: usize = 8;
const MAX_TOOL_CHARS: usize = morphai::DEFAULT_TOOL_RESULT_MAX_CHARS;

#[derive(Debug, serde::Deserialize)]
struct ToolCall {
    tool: String,
    args: serde_json::Value,
}

pub async fn respond_general_llm(
    state: &crate::state::AppState,
    messages: &[AssistantMessage],
    assistant_state: &AssistantState,
    last_user: &str,
) -> Result<String, String> {
    let cfg = Config::from_env();
    if !cfg.configured() {
        return Err("MorphAI not configured".into());
    }
    let client = Client::new(cfg);

    let mut first = USERS_PANEL_INSTRUCTIONS.to_string();
    if let Some(hist) = format_history(messages) {
        first.push_str("\n\nRecent conversation:\n");
        first.push_str(&hist);
    }
    if let Some(st) = format_state(assistant_state) {
        first.push_str("\n\nActive assistant state:\n");
        first.push_str(&st);
    }
    first.push_str("\n\nLatest user message:\n");
    first.push_str(last_user);

    let mut llm_messages = vec![Message {
        role: "user".to_string(),
        content: first,
    }];

    for _ in 0..MAX_ROUNDS {
        let reply = client
            .chat_completion(&llm_messages)
            .await
            .map_err(|e| e.to_string())?;
        let reply = reply.trim();
        if reply.is_empty() {
            return Err("empty model response".into());
        }

        let Some(obj) = extract_json_object(reply) else {
            return Ok(reply.to_string());
        };

        let call: ToolCall = match serde_json::from_str(&obj) {
            Ok(c) => c,
            Err(_) => return Ok(reply.to_string()),
        };

        let result = exec_tool(state, &call).await;
        let tool_msg = match result {
            Ok(s) => format!("TOOL_RESULT\n{s}"),
            Err(e) => format!("TOOL_RESULT error={e}"),
        };

        let follow_up = morphai::tool_follow_up_prompt(&tool_msg);
        llm_messages.push(Message {
            role: "user".to_string(),
            content: follow_up,
        });
    }

    Ok("I hit the tool limit while gathering data. Please narrow your question.".to_string())
}

async fn exec_tool(
    state: &crate::state::AppState,
    call: &ToolCall,
) -> Result<String, String> {
    match call.tool.as_str() {
        "list_roles" => {
            let rows: Vec<crate::models::RoleRow> =
                sqlx::query_as("SELECT * FROM plat_roles ORDER BY name ASC")
                    .fetch_all(&state.pool)
                    .await
                    .map_err(|e| e.to_string())?;
            let names: Vec<String> = rows.into_iter().map(|r| r.name).collect();
            Ok(truncate(&names.join(", ")))
        }
        "list_users" => {
            let limit = call
                .args
                .get("limit")
                .and_then(|v| v.as_i64())
                .unwrap_or(25)
                .clamp(1, 100) as i64;
            let rows: Vec<crate::models::UserRow> = sqlx::query_as(
                "SELECT * FROM plat_users ORDER BY created_at DESC LIMIT ?",
            )
            .bind(limit)
            .fetch_all(&state.pool)
            .await
            .map_err(|e| e.to_string())?;
            let lines: Vec<String> = rows
                .iter()
                .map(|u| format!("{} <{}>", u.username, u.email))
                .collect();
            Ok(truncate(&lines.join("\n")))
        }
        other => Err(format!("unknown tool: {other}")),
    }
}

fn format_history(messages: &[AssistantMessage]) -> Option<String> {
    let start = messages.len().saturating_sub(morphai::DEFAULT_HISTORY_MAX_MESSAGES);
    let mut out = String::new();
    for m in &messages[start..] {
        let role = m.role.trim();
        if role != "user" && role != "assistant" {
            continue;
        }
        let content = morphai::truncate_chars(m.content.trim(), morphai::DEFAULT_HISTORY_MAX_CHARS);
        if content.is_empty() {
            continue;
        }
        out.push_str(role);
        out.push_str(": ");
        out.push_str(&content);
        out.push('\n');
    }
    if out.is_empty() {
        None
    } else {
        Some(out)
    }
}

fn format_state(state: &AssistantState) -> Option<String> {
    if state.intent.is_empty() && state.fields.is_empty() {
        return None;
    }
    Some(format!(
        "intent={} fields={:?}",
        state.intent, state.fields
    ))
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

fn truncate(s: &str) -> String {
    morphai::truncate_chars(s, MAX_TOOL_CHARS)
}
