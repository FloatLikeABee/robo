//! Shared context limits and truncation for Rust AI assistants.

pub const DEFAULT_HISTORY_MAX_MESSAGES: usize = 4;
pub const DEFAULT_HISTORY_MAX_CHARS: usize = 500;
pub const DEFAULT_TOOL_RESULT_MAX_CHARS: usize = 8_000;
pub const DEFAULT_TOOL_MAX_ROUNDS: usize = 8;
pub const FAST_TOOL_FIRST_INSTRUCTIONS: &str = "Use the fastest grounded source available. Prefer live MCP-style catalogs, product APIs, repositories, or read-only database/schema lookups before broad reasoning. List or search first, then fetch details by id. Use read-only SQL unless the user explicitly asks to create, update, or delete.";

/// Truncate at a UTF-8 char boundary (not byte index).
pub fn truncate_chars(s: &str, max_chars: usize) -> String {
    if max_chars == 0 {
        return String::new();
    }
    if s.chars().count() <= max_chars {
        return s.to_string();
    }
    let mut out: String = s.chars().take(max_chars).collect();
    out.push('…');
    out
}

pub fn tool_follow_up_prompt(tool_result: &str) -> String {
    format!(
        "{tool_result}\n\nSummarize for the user in markdown. If you need another tool, reply with only one JSON object."
    )
}

pub fn tool_follow_up_prompt_with_instruction(tool_result: &str, instruction: &str) -> String {
    let instruction = instruction.trim();
    if instruction.is_empty() {
        return tool_follow_up_prompt(tool_result);
    }
    format!("{tool_result}\n\n{instruction} If another tool is needed, reply with only one JSON object.")
}

pub fn extract_json_object(s: &str) -> Option<String> {
    let mut trimmed = s.trim();
    for prefix in ["```json", "```JSON", "```"] {
        if let Some(rest) = trimmed.strip_prefix(prefix) {
            trimmed = rest.trim();
            break;
        }
    }
    if let Some(rest) = trimmed.strip_suffix("```") {
        trimmed = rest.trim();
    }
    let start = trimmed.find('{')?;
    let mut depth = 0i32;
    for (i, ch) in trimmed[start..].char_indices() {
        match ch {
            '{' => depth += 1,
            '}' => {
                depth -= 1;
                if depth == 0 {
                    return Some(trimmed[start..start + i + 1].to_string());
                }
            }
            _ => {}
        }
    }
    None
}
