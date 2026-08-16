//! LLM + tool loop for DataX general assistant queries.

use morphai::{Client, Config, Message};
use serde::Deserialize;
use uuid::Uuid;

use crate::api::data_table_query::{ParsedRow, RowQueryOptions, aggregate_rows, apply_row_query};
use crate::api::report_assistant::{AssistantState, ChatMessage};
use crate::db::data_table_repository::DataTableRepository;
use crate::db::repositories::DatabaseConnectionRepository;
use crate::services::AppState;
use serde_json::json;

const DATA_AI_INSTRUCTIONS: &str = r#"You are Data AI for DataX. Your ONLY job is to help users explore, analyze, and report on data already connected in their DataX workspace.

STRICT SCOPE — you MUST refuse any request that is not directly about the user's connected databases, imported data tables, schemas, tables, columns, SQL queries, query results, dashboards, or data analysis reports derived from that data. Do not answer general knowledge, coding help, product advice outside DataX data, chit-chat, or anything unrelated to their data.

When refusing off-topic requests, reply briefly: "I'm Data AI — I only help with your DataX data. Ask me to explore databases, data tables, run queries, analyze tables, or generate a data report."

When you need live data, respond with ONLY one JSON object (no markdown):
{"tool":"<name>","args":{...}}

Tools:
- list_databases — connected database list (id, name, engine)
- database_schema — args: database_id (uuid string) — tables and columns
- execute_query — args: database_id (uuid string), sql (read-only SELECT/WITH only) — run query and return up to 500 rows
- list_data_tables — imported DataX data tables (id, name, columns, row_count)
- data_table_schema — args: table_id (uuid string) — columns, row_count, sample rows
- query_data_table — args: table_id (uuid string); optional search, sort_by, sort_dir, limit, offset, group_by, aggregate_op (count|sum|avg|min|max), aggregate_column — search, sort, and aggregate imported table rows

When the user attached data tables to the conversation (see Active assistant state), prefer those table_id values for analysis.

After TOOL_RESULT, summarize findings in markdown with tables where helpful.
When the user asks to "analyze" or "generate a report", format the response as:
## Overview
## Key Findings
## Risks or Gaps
## Recommended Actions
If no tool is needed and the question is in scope, reply in markdown only."#;

const SHARP_REPORT_INSTRUCTIONS: &str = r#"You are DataX AI, an expert assistant for analytics, Metabase, SQL, and data reports.

You help staff with:
- Connected databases, imported data tables, schemas, tables, and columns
- SQL mode vs Visual mode vs file imports (CSV/JSON/XLSX)
- Metabase dashboards embedded in DataX
- Chart types and data reports workflow
- Markdown analysis reports for uploaded files and connected database structures

To start a guided create flow, users can say:
- "create report report_name: … data_source: …"

When you need live DataX data, respond with ONLY one JSON object (no markdown):
{"tool":"<name>","args":{...}}

Tools:
- list_databases — connected database list (id, name, engine)
- database_schema — args: database_id (uuid string) — tables and columns
- list_data_tables — imported DataX data tables (id, name, columns, row_count)
- data_table_schema — args: table_id (uuid string) — columns, row_count, sample rows
- query_data_table — args: table_id (uuid string); optional search, sort_by, sort_dir, limit, offset, group_by, aggregate_op (count|sum|avg|min|max), aggregate_column

When the user attached data tables to the conversation (see Active assistant state), prefer those table_id values for analysis.

After TOOL_RESULT, summarize in markdown with **Label:** values.
When the user asks to "analyze", format the response as a report with sections:
## Overview
## Key Findings
## Risks or Gaps
## Recommended Actions
If no tool is needed, reply in markdown only."#;

const MAX_ROUNDS: usize = morphai::DEFAULT_TOOL_MAX_ROUNDS;
const MAX_TOOL_CHARS: usize = morphai::DEFAULT_TOOL_RESULT_MAX_CHARS;
const MAX_TOOL_QUERY_ROWS: i32 = 300;

#[derive(Debug, Deserialize)]
struct ToolCall {
    tool: String,
    args: serde_json::Value,
}

#[derive(Copy, Clone)]
enum LlmMode {
    General,
    DataAi,
}

pub async fn respond_data_ai_llm(
    state: &AppState,
    messages: &[ChatMessage],
    assistant_state: &AssistantState,
    last_user: &str,
) -> Result<String, String> {
    respond_llm(state, messages, assistant_state, last_user, LlmMode::DataAi).await
}

pub async fn respond_general_llm(
    state: &AppState,
    messages: &[ChatMessage],
    assistant_state: &AssistantState,
    last_user: &str,
) -> Result<String, String> {
    respond_llm(
        state,
        messages,
        assistant_state,
        last_user,
        LlmMode::General,
    )
    .await
}

async fn respond_llm(
    state: &AppState,
    messages: &[ChatMessage],
    assistant_state: &AssistantState,
    last_user: &str,
    mode: LlmMode,
) -> Result<String, String> {
    crate::config::load_env_files();
    let cfg = Config::from_env();
    if !cfg.configured() {
        return Err("MorphAI not configured".into());
    }
    let client = Client::new(cfg);
    let last_user_trimmed = morphai::truncate_chars(last_user.trim(), 4_000);

    let mut first = match mode {
        LlmMode::General => SHARP_REPORT_INSTRUCTIONS.to_string(),
        LlmMode::DataAi => DATA_AI_INSTRUCTIONS.to_string(),
    };
    first.push_str("\n\nFast source selection:\n");
    first.push_str(morphai::FAST_TOOL_FIRST_INSTRUCTIONS);
    if let Some(hist) = format_history(messages) {
        first.push_str("\n\nRecent conversation:\n");
        first.push_str(&hist);
    }
    if let Some(st) = format_state(assistant_state) {
        first.push_str("\n\nActive assistant state:\n");
        first.push_str(&st);
    }
    first.push_str("\n\nLatest user message:\n");
    first.push_str(&last_user_trimmed);

    let mut llm_messages = vec![Message {
        role: "user".to_string(),
        content: first,
    }];
    let initial = llm_messages[0].clone();

    for _ in 0..MAX_ROUNDS {
        let reply = client
            .chat_completion(&llm_messages)
            .await
            .map_err(|e| e.to_string())?;
        let reply = reply.trim();
        if reply.is_empty() {
            return Err("empty model response".into());
        }

        let Some(obj) = morphai::extract_json_object(reply) else {
            return Ok(reply.to_string());
        };

        let call: ToolCall = match serde_json::from_str(&obj) {
            Ok(c) => c,
            Err(_) => return Ok(reply.to_string()),
        };

        let result = exec_tool(state, &call, mode).await;
        let tool_msg = match result {
            Ok(s) => format!("TOOL_RESULT\n{s}"),
            Err(e) => format!("TOOL_RESULT error={e}"),
        };

        let follow_up = morphai::tool_follow_up_prompt_with_instruction(
            &tool_msg,
            "Summarize findings in markdown with tables where helpful.",
        );
        llm_messages = vec![
            initial.clone(),
            Message {
                role: "user".to_string(),
                content: follow_up,
            },
        ];
    }

    Ok("I hit the tool limit while gathering data. Please narrow your question.".to_string())
}

async fn exec_tool(state: &AppState, call: &ToolCall, mode: LlmMode) -> Result<String, String> {
    match call.tool.as_str() {
        "list_databases" => {
            let repo = DatabaseConnectionRepository::new(state.db_pool.clone());
            let rows = repo.find_all().await.map_err(|e| e.to_string())?;
            if rows.is_empty() {
                return Ok("No database connections configured.".into());
            }
            let lines: Vec<String> = rows
                .into_iter()
                .map(|r| format!("{} ({}) engine={}", r.name, r.id, r.engine))
                .collect();
            Ok(truncate(&lines.join("\n")))
        }
        "database_schema" => {
            let id_str = call
                .args
                .get("database_id")
                .and_then(|v| v.as_str())
                .unwrap_or("")
                .trim();
            if id_str.is_empty() {
                return Err("database_id is required".into());
            }
            let uuid =
                Uuid::parse_str(id_str).map_err(|_| "invalid database_id uuid".to_string())?;
            let (_conn, tables) = crate::api::schema::load_schema_snapshot(state, &uuid).await?;
            if tables.is_empty() {
                return Ok("No tables found.".into());
            }
            let mut lines = Vec::new();
            for t in tables.iter().take(40) {
                let cols: Vec<String> = t
                    .columns
                    .iter()
                    .take(20)
                    .map(|c| format!("{} ({})", c.name, c.data_type))
                    .collect();
                lines.push(format!("{}.{}: {}", t.schema, t.name, cols.join(", ")));
            }
            Ok(truncate(&lines.join("\n")))
        }
        "list_data_tables" => {
            let repo = DataTableRepository::new(state.db_pool.clone());
            let rows = repo.list_all().await.map_err(|e| e.to_string())?;
            if rows.is_empty() {
                return Ok("No imported data tables.".into());
            }
            let lines: Vec<String> = rows
                .into_iter()
                .map(|r| {
                    let cols: Vec<String> =
                        serde_json::from_str(&r.column_schema).unwrap_or_default();
                    format!(
                        "{} ({}) rows={} cols={}",
                        r.name,
                        r.id,
                        r.row_count,
                        cols.join(", ")
                    )
                })
                .collect();
            Ok(truncate(&lines.join("\n")))
        }
        "data_table_schema" => {
            let id_str = call
                .args
                .get("table_id")
                .and_then(|v| v.as_str())
                .unwrap_or("")
                .trim();
            if id_str.is_empty() {
                return Err("table_id is required".into());
            }
            let uuid = Uuid::parse_str(id_str).map_err(|_| "invalid table_id uuid".to_string())?;
            let repo = DataTableRepository::new(state.db_pool.clone());
            let table = repo
                .find_by_id(uuid)
                .await
                .map_err(|e| e.to_string())?
                .ok_or_else(|| "data table not found".to_string())?;
            let columns: Vec<String> =
                serde_json::from_str(&table.column_schema).unwrap_or_default();
            let sample_rows = repo
                .fetch_rows(uuid, 5, 0)
                .await
                .map_err(|e| e.to_string())?;
            let samples: Vec<String> = sample_rows.into_iter().map(|r| truncate(&r.data)).collect();
            Ok(truncate(&format!(
                "name={} id={} rows={} columns={}\nsamples:\n{}",
                table.name,
                table.id,
                table.row_count,
                columns.join(", "),
                samples.join("\n")
            )))
        }
        "query_data_table" => {
            let id_str = call
                .args
                .get("table_id")
                .and_then(|v| v.as_str())
                .unwrap_or("")
                .trim();
            if id_str.is_empty() {
                return Err("table_id is required".into());
            }
            let uuid = Uuid::parse_str(id_str).map_err(|_| "invalid table_id uuid".to_string())?;
            let repo = DataTableRepository::new(state.db_pool.clone());
            let table = repo
                .find_by_id(uuid)
                .await
                .map_err(|e| e.to_string())?
                .ok_or_else(|| "data table not found".to_string())?;
            let columns: Vec<String> =
                serde_json::from_str(&table.column_schema).unwrap_or_default();
            let fetched = repo
                .fetch_rows(uuid, MAX_TOOL_QUERY_ROWS, 0)
                .await
                .map_err(|e| e.to_string())?;
            let parsed: Vec<ParsedRow> = fetched
                .into_iter()
                .filter_map(|r| {
                    let data: serde_json::Value = serde_json::from_str(&r.data).ok()?;
                    Some(ParsedRow {
                        row_index: r.row_index,
                        data,
                    })
                })
                .collect();
            let sampled_note = if table.row_count > MAX_TOOL_QUERY_ROWS {
                format!(
                    "\n(note: analyzed first {} of {} rows)",
                    MAX_TOOL_QUERY_ROWS, table.row_count
                )
            } else {
                String::new()
            };

            let search = call
                .args
                .get("search")
                .and_then(|v| v.as_str())
                .map(|s| s.to_string());
            let sort_by = call
                .args
                .get("sort_by")
                .and_then(|v| v.as_str())
                .map(|s| s.to_string());
            let sort_dir = call
                .args
                .get("sort_dir")
                .and_then(|v| v.as_str())
                .unwrap_or("asc")
                .to_string();
            let limit = call
                .args
                .get("limit")
                .and_then(|v| v.as_i64())
                .unwrap_or(100)
                .clamp(1, 500) as i32;
            let offset = call
                .args
                .get("offset")
                .and_then(|v| v.as_i64())
                .unwrap_or(0)
                .max(0) as i32;

            let filter_opts = RowQueryOptions {
                search: search.clone(),
                sort_by: sort_by.clone(),
                sort_dir: sort_dir.clone(),
                limit: 5000,
                offset: 0,
            };
            let (filtered, total) = apply_row_query(parsed, &columns, &filter_opts);

            let page_opts = RowQueryOptions {
                search: None,
                sort_by: None,
                sort_dir,
                limit,
                offset,
            };
            let (page, _) = apply_row_query(filtered.clone(), &columns, &page_opts);

            let aggregate = call
                .args
                .get("aggregate_op")
                .and_then(|v| v.as_str())
                .map(|op| {
                    aggregate_rows(
                        &filtered,
                        call.args.get("group_by").and_then(|v| v.as_str()),
                        op,
                        call.args.get("aggregate_column").and_then(|v| v.as_str()),
                    )
                });

            let rows_json: Vec<serde_json::Value> = page.into_iter().map(|r| r.data).collect();
            let body = json!({
                "table": table.name,
                "total": total,
                "rows": rows_json,
                "aggregate": aggregate
            });
            Ok(truncate(&(body.to_string() + &sampled_note)))
        }
        "execute_query" if matches!(mode, LlmMode::DataAi) => {
            let id_str = call
                .args
                .get("database_id")
                .and_then(|v| v.as_str())
                .unwrap_or("")
                .trim();
            let sql = call
                .args
                .get("sql")
                .and_then(|v| v.as_str())
                .unwrap_or("")
                .trim();
            if id_str.is_empty() {
                return Err("database_id is required".into());
            }
            if sql.is_empty() {
                return Err("sql is required".into());
            }
            let (row_count, results) =
                crate::api::queries::execute_readonly_sql(state, id_str, sql).await?;
            let body = serde_json::json!({
                "row_count": row_count,
                "results": results,
            });
            Ok(truncate(&body.to_string()))
        }
        other => Err(format!("unknown tool: {other}")),
    }
}

fn format_history(messages: &[ChatMessage]) -> Option<String> {
    let start = messages
        .len()
        .saturating_sub(morphai::DEFAULT_HISTORY_MAX_MESSAGES);
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
    if out.is_empty() { None } else { Some(out) }
}

fn format_state(state: &AssistantState) -> Option<String> {
    if state.intent.is_empty() && state.fields.is_empty() && state.attached_data_tables.is_empty() {
        return None;
    }
    let mut parts = Vec::new();
    if !state.intent.is_empty() {
        parts.push(format!("intent={}", state.intent));
    }
    if !state.fields.is_empty() {
        parts.push(format!("fields={:?}", state.fields));
    }
    if !state.attached_data_tables.is_empty() {
        let tables = state
            .attached_data_tables
            .iter()
            .map(|t| format!("- {} (table_id: {})", t.name, t.id))
            .collect::<Vec<_>>()
            .join("\n");
        parts.push(format!(
            "ATTACHED DATA TABLES — use these table_id values with data_table_schema / query_data_table:\n{tables}"
        ));
    }
    Some(parts.join("\n"))
}

fn truncate(s: &str) -> String {
    morphai::truncate_chars(s, MAX_TOOL_CHARS)
}
