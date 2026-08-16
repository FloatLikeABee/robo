//! Shared MorphAI configuration and DashScope client for Rust apps (SharpReport, UsersPanel).

pub mod client;
pub mod config;
pub mod context;

pub use client::Client;
pub use client::Message;
pub use config::Config;
pub use context::{
    extract_json_object, tool_follow_up_prompt, tool_follow_up_prompt_with_instruction,
    truncate_chars, DEFAULT_HISTORY_MAX_CHARS, DEFAULT_HISTORY_MAX_MESSAGES,
    DEFAULT_TOOL_MAX_ROUNDS, DEFAULT_TOOL_RESULT_MAX_CHARS, FAST_TOOL_FIRST_INSTRUCTIONS,
};
