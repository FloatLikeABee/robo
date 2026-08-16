//! Lightweight report-building assistant with clarification state.
//! Compatible with MorphAI contract (`assistant_message`, state round-trip).

use axum::Json;
use axum::extract::State;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

use crate::api::report_assistant_llm::{respond_data_ai_llm, respond_general_llm};
use crate::services::AppState;

#[derive(Debug, Deserialize, Serialize)]
pub struct ChatMessage {
    pub role: String,
    pub content: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct AttachedDataTableRef {
    pub id: String,
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct AssistantState {
    #[serde(default)]
    pub intent: String,
    #[serde(default)]
    pub fields: HashMap<String, String>,
    #[serde(default)]
    pub attached_data_tables: Vec<AttachedDataTableRef>,
}

#[derive(Debug, Deserialize)]
pub struct ChatRequest {
    pub messages: Vec<ChatMessage>,
    #[serde(default)]
    pub state: AssistantState,
}

#[derive(Debug, Serialize)]
pub struct ChatResponse {
    pub assistant_message: String,
    #[serde(skip_serializing_if = "String::is_empty")]
    pub intent: String,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub missing_fields: Vec<String>,
    pub state: AssistantState,
    pub completed: bool,
}

pub async fn data_ai_chat(
    State(state): State<AppState>,
    Json(req): Json<ChatRequest>,
) -> Json<ChatResponse> {
    let user_message = latest_user_message(&req.messages);
    let lower = user_message.to_lowercase();

    if is_off_topic_data_question(&lower) {
        return Json(ChatResponse {
            assistant_message: "I'm **Data AI** — I only help with your DataX data. Ask me to explore databases, data tables, run queries, analyze tables, or generate a data report.".to_string(),
            intent: "data_ai".to_string(),
            missing_fields: vec![],
            state: req.state,
            completed: true,
        });
    }

    match respond_data_ai_llm(&state, &req.messages, &req.state, &user_message).await {
        Ok(reply) => Json(ChatResponse {
            assistant_message: reply,
            intent: "data_ai".to_string(),
            missing_fields: vec![],
            state: req.state,
            completed: true,
        }),
        Err(e) => Json(ChatResponse {
            assistant_message: if e.contains("not configured") || e.contains("MORPH_AI_API_KEY") {
                data_ai_guidance_reply(&lower)
            } else {
                llm_error_reply(&e, &lower)
            },
            intent: "data_ai".to_string(),
            missing_fields: vec![],
            state: req.state,
            completed: true,
        }),
    }
}

pub async fn chat(
    State(state): State<AppState>,
    Json(mut req): Json<ChatRequest>,
) -> Json<ChatResponse> {
    let user_message = latest_user_message(&req.messages);
    let lower = user_message.to_lowercase();

    if req.state.intent.is_empty() {
        req.state.intent = detect_intent(&lower);
    }
    update_state_fields(&mut req.state.fields, &user_message);

    match req.state.intent.as_str() {
        "create_report" => {
            let missing = missing_required(&req.state.fields, &["report_name", "data_source"]);
            if !missing.is_empty() {
                return Json(ChatResponse {
                    assistant_message: format!(
                        "I can create the report. Please provide:\n\n{}",
                        missing
                            .iter()
                            .map(|f| format!("- `{f}`"))
                            .collect::<Vec<_>>()
                            .join("\n")
                    ),
                    intent: req.state.intent.clone(),
                    missing_fields: missing,
                    state: req.state,
                    completed: false,
                });
            }
            Json(ChatResponse {
                assistant_message: "✓ **Ready to build your report**\n\nOpen **Report Builder**, pick your data source, and run the query — I can help with SQL and charts along the way.".to_string(),
                intent: "create_report".to_string(),
                missing_fields: vec![],
                state: req.state,
                completed: true,
            })
        }
        _ => match respond_general_llm(&state, &req.messages, &req.state, &user_message).await {
            Ok(reply) => Json(ChatResponse {
                assistant_message: reply,
                intent: "general".to_string(),
                missing_fields: vec![],
                state: req.state,
                completed: true,
            }),
            Err(e) => Json(ChatResponse {
                assistant_message: llm_error_reply(&e, &lower),
                intent: "general".to_string(),
                missing_fields: vec![],
                state: req.state,
                completed: true,
            }),
        },
    }
}

fn latest_user_message(messages: &[ChatMessage]) -> String {
    messages
        .iter()
        .rev()
        .find(|m| m.role.eq_ignore_ascii_case("user"))
        .map(|m| m.content.trim().to_string())
        .unwrap_or_default()
}

fn detect_intent(lower: &str) -> String {
    if lower.contains("create report")
        || lower.contains("new report")
        || lower.contains("build report")
    {
        "create_report".to_string()
    } else {
        "general".to_string()
    }
}

fn update_state_fields(fields: &mut HashMap<String, String>, message: &str) {
    for token in ["report_name", "data_source", "sql", "chart_type"] {
        if let Some(value) = extract_token_value(message, token) {
            fields.insert(token.to_string(), value);
        }
    }
}

fn extract_token_value(message: &str, token: &str) -> Option<String> {
    let lower = message.to_lowercase();
    let idx = lower.find(token)?;
    let value = &message[idx + token.len()..];
    let cleaned = value
        .trim_start_matches(&[':', '=', '-', ' '][..])
        .split(&[',', ';', '\n'][..])
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

fn missing_required(fields: &HashMap<String, String>, required: &[&str]) -> Vec<String> {
    required
        .iter()
        .filter_map(|key| {
            let has_value = fields
                .get(*key)
                .map(|v| !v.trim().is_empty())
                .unwrap_or(false);
            if has_value {
                None
            } else {
                Some((*key).to_string())
            }
        })
        .collect()
}

fn is_off_topic_data_question(lower: &str) -> bool {
    if lower.is_empty() {
        return false;
    }
    let data_signals = [
        "database",
        "data",
        "sql",
        "table",
        "schema",
        "column",
        "query",
        "report",
        "dashboard",
        "metabase",
        "analyze",
        "analysis",
        "chart",
        "row",
        "count",
        "sum",
        "avg",
        "join",
        "select",
        "where",
        "aggregate",
        "insight",
        "metric",
        "connection",
        "postgres",
        "mysql",
        "sqlite",
    ];
    if data_signals.iter().any(|s| lower.contains(s)) {
        return false;
    }
    let off_topic = [
        "weather",
        "recipe",
        "joke",
        "poem",
        "who is",
        "what is the capital",
        "write code",
        "python tutorial",
        "javascript",
        "hello",
        "how are you",
        "tell me about yourself",
        "news",
        "sports",
    ];
    off_topic.iter().any(|s| lower.contains(s))
}

fn data_ai_guidance_reply(last: &str) -> String {
    if is_off_topic_data_question(last) {
        return "I'm **Data AI** — I only help with your DataX data. Ask me to explore databases, data tables, run queries, analyze tables, or generate a data report.".to_string();
    }
    if last.contains("schema") || last.contains("table") || last.contains("column") {
        "I can list your **connected databases** and show **schemas** when you sign in and set `MORPH_AI_API_KEY`.".to_string()
    } else if last.contains("analyze") || last.contains("report") {
        "Ask me to **analyze your data** or **generate a report** — I'll inspect your databases and summarize findings.".to_string()
    } else if last.contains("sql") || last.contains("query") || last.contains("select") {
        "I can run **read-only SQL** against your connected databases. Try: \"Run a sample query on my main database\".".to_string()
    } else {
        "Ask about your **databases**, **schemas**, **queries**, or request a **data analysis report**.\n\nSet `MORPH_AI_API_KEY` to enable live database access.".to_string()
    }
}

fn llm_error_reply(err: &str, last: &str) -> String {
    if err.contains("not configured") || err.contains("MORPH_AI_API_KEY") {
        return guidance_reply(last);
    }
    let hint = if err.contains("Api key is invalid") || err.contains("invalid_api_key") {
        "Verify your provider API key in `SharpReport/.env` (SiliconFlow keys must be used with `MORPH_AI_BASE_URL=https://api.siliconflow.cn/v1`)."
    } else if err.contains("model") && (err.contains("not exist") || err.contains("not found")) {
        "Check `MORPH_AI_MODEL` matches a model your provider account can access."
    } else if err.contains("failed while sending the request body")
        || err.contains("uploading the request")
        || err.contains("is_request")
    {
        "SiliconFlow uploads can fail intermittently (the backend now retries automatically). If errors persist, switch provider in `SharpReport/.env`: `MORPH_AI_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1` with a DashScope key, then `./start-all.sh restart sharpreport-api`. Check `GET /api/v1/ai/status`."
    } else {
        "Check `MORPH_AI_API_KEY`, `MORPH_AI_MODEL`, and `MORPH_AI_BASE_URL` in `SharpReport/.env`, then restart the DataX backend (`./start-all.sh restart sharpreport-api`). Open `/api/v1/ai/status` to verify the running process sees your config."
    };
    format!("**AI assistant error:** {err}\n\n{hint}")
}

fn guidance_reply(last: &str) -> String {
    if last.contains("metabase") {
        "**Metabase** is integrated in Data reports and dashboards. You can work with embedded Metabase while staying inside DataX.\n\nSet `MORPH_AI_API_KEY` to enable the full LLM assistant.".to_string()
    } else if last.contains("sql")
        || last.contains("join")
        || last.contains("where")
        || last.contains("select")
    {
        "Use **SQL mode** for multi-table logic and read-only analytics queries.\n\nSet `MORPH_AI_API_KEY` for live schema help.".to_string()
    } else if last.contains("csv")
        || last.contains("excel")
        || last.contains(".xlsx")
        || last.contains("json")
        || last.contains("file")
        || last.contains("import")
    {
        "**File mode** supports CSV, JSON, and XLSX for ad-hoc analysis.".to_string()
    } else if last.contains("visual") || last.contains("field") || last.contains("aggregate") {
        "**Visual mode** is ideal for one-table query building with guided aggregations."
            .to_string()
    } else if last.contains("chart") || last.contains("bar") || last.contains("pie") {
        "After running a query, switch the results panel between **table** and **chart** views."
            .to_string()
    } else {
        "I can help with report creation, SQL, data-source selection, chart setup, and general analytics questions.\n\nSet `MORPH_AI_API_KEY` to enable the full LLM assistant.".to_string()
    }
}
