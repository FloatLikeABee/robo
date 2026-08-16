use crate::config::Config;
use crate::context::truncate_chars;
use reqwest::header::{AUTHORIZATION, CONTENT_TYPE};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize)]
pub struct Message {
    pub role: String,
    pub content: String,
}

#[derive(Serialize)]
struct DashScopeRequest<'a> {
    model: &'a str,
    input: DashScopeInput<'a>,
}

#[derive(Serialize)]
struct DashScopeInput<'a> {
    messages: &'a [Message],
}

#[derive(Deserialize)]
struct DashScopeResponse {
    output: Option<DashScopeOutput>,
    code: Option<String>,
    message: Option<String>,
}

#[derive(Deserialize)]
struct DashScopeOutput {
    choices: Vec<DashScopeChoice>,
}

#[derive(Deserialize)]
struct DashScopeChoice {
    message: DashScopeChoiceMessage,
}

#[derive(Deserialize)]
struct DashScopeChoiceMessage {
    content: String,
}

#[derive(Serialize)]
struct OpenAIRequest<'a> {
    model: &'a str,
    messages: &'a [Message],
}

#[derive(Deserialize)]
struct OpenAIResponse {
    choices: Option<Vec<OpenAIChoice>>,
    error: Option<OpenAIErrorBody>,
}

#[derive(Deserialize)]
struct OpenAIChoice {
    message: OpenAIChoiceMessage,
}

#[derive(Deserialize)]
struct OpenAIChoiceMessage {
    content: String,
}

#[derive(Deserialize)]
struct OpenAIErrorBody {
    message: Option<String>,
}

pub struct Client {
    cfg: Config,
    http: reqwest::Client,
}

/// Cap total message content sent to the provider (avoids upload failures on large tool loops).
pub const MAX_CHAT_REQUEST_CHARS: usize = 12_000;

const MAX_CHAT_RETRIES: usize = 3;

fn trim_messages_for_api(messages: &[Message], max_total_chars: usize) -> Vec<Message> {
    if messages.is_empty() {
        return vec![];
    }
    let mut out: Vec<Message> = messages.to_vec();
    while message_chars(&out) > max_total_chars && out.len() > 1 {
        out.remove(1);
    }
    if message_chars(&out) > max_total_chars {
        let excess = message_chars(&out) - max_total_chars;
        let first_len = out[0].content.chars().count();
        let keep = first_len.saturating_sub(excess + 64);
        out[0].content = truncate_chars(&out[0].content, keep);
    }
    out
}

fn message_chars(messages: &[Message]) -> usize {
    messages.iter().map(|m| m.content.chars().count()).sum()
}

impl Client {
    pub fn new(cfg: Config) -> Self {
        Self {
            cfg,
            http: reqwest::Client::builder()
                .timeout(std::time::Duration::from_secs(180))
                .connect_timeout(std::time::Duration::from_secs(30))
                .tcp_keepalive(std::time::Duration::from_secs(30))
                .pool_max_idle_per_host(2)
                .build()
                .expect("reqwest client"),
        }
    }

    pub fn from_env() -> Self {
        Self::new(Config::from_env())
    }

    pub async fn chat_completion(&self, messages: &[Message]) -> Result<String, String> {
        if messages.is_empty() {
            return Err("no messages".into());
        }
        if !self.cfg.configured() {
            return Err("MORPH_AI_API_KEY is not configured".into());
        }

        let messages = trim_messages_for_api(messages, MAX_CHAT_REQUEST_CHARS);

        if self.cfg.uses_openai_compatible() {
            return self.chat_with_retries(&messages, true).await;
        }
        self.chat_with_retries(&messages, false).await
    }

    async fn chat_with_retries(
        &self,
        messages: &[Message],
        openai_compatible: bool,
    ) -> Result<String, String> {
        let mut last_err = String::new();
        for attempt in 0..MAX_CHAT_RETRIES {
            let result = if openai_compatible {
                self.chat_openai_compatible(messages).await
            } else {
                self.chat_dashscope_native(messages).await
            };
            match result {
                Ok(reply) => return Ok(reply),
                Err(e) => {
                    last_err = e;
                    if attempt + 1 < MAX_CHAT_RETRIES && is_retriable_error(&last_err) {
                        let delay_ms = 750 * (attempt + 1) as u64;
                        tokio::time::sleep(std::time::Duration::from_millis(delay_ms)).await;
                        continue;
                    }
                    break;
                }
            }
        }
        Err(last_err)
    }

    async fn chat_openai_compatible(&self, messages: &[Message]) -> Result<String, String> {
        let url = self.cfg.chat_completions_url();
        let body = OpenAIRequest {
            model: &self.cfg.model,
            messages,
        };

        let resp = self
            .http
            .post(&url)
            .header(AUTHORIZATION, format!("Bearer {}", self.cfg.api_key))
            .header(CONTENT_TYPE, "application/json")
            .json(&body)
            .send()
            .await
            .map_err(format_request_error)?;

        let status = resp.status();
        let text = resp.text().await.map_err(|e| e.to_string())?;
        if !status.is_success() {
            return Err(format!("API status {}: {}", status, text));
        }

        let parsed: OpenAIResponse = serde_json::from_str(&text).map_err(|e| e.to_string())?;
        if let Some(err) = parsed.error {
            return Err(err.message.unwrap_or_else(|| "API error".to_string()));
        }
        parsed
            .choices
            .and_then(|c| c.into_iter().next())
            .map(|c| c.message.content)
            .filter(|s| !s.trim().is_empty())
            .ok_or_else(|| "no response from AI model".into())
    }

    async fn chat_dashscope_native(&self, messages: &[Message]) -> Result<String, String> {
        let body = DashScopeRequest {
            model: &self.cfg.model,
            input: DashScopeInput { messages },
        };

        let resp = self
            .http
            .post(&self.cfg.api_url)
            .header(AUTHORIZATION, format!("Bearer {}", self.cfg.api_key))
            .header(CONTENT_TYPE, "application/json")
            .json(&body)
            .send()
            .await
            .map_err(format_request_error)?;

        let status = resp.status();
        let text = resp.text().await.map_err(|e| e.to_string())?;
        if !status.is_success() {
            return Err(format!("API status {}: {}", status, text));
        }

        let parsed: DashScopeResponse = serde_json::from_str(&text).map_err(|e| e.to_string())?;
        if let Some(code) = parsed.code {
            if code != "Success" {
                return Err(format!(
                    "API error: {} - {}",
                    code,
                    parsed.message.unwrap_or_default()
                ));
            }
        }
        parsed
            .output
            .and_then(|o| o.choices.into_iter().next())
            .map(|c| c.message.content)
            .ok_or_else(|| "no response from AI model".into())
    }
}

fn is_retriable_error(err: &str) -> bool {
    let lower = err.to_lowercase();
    lower.contains("uploading the request")
        || lower.contains("failed while sending the request body")
        || lower.contains("connection closed")
        || lower.contains("connection reset")
        || lower.contains("broken pipe")
        || lower.contains("timed out")
        || lower.contains("timeout")
        || lower.contains("connection refused")
        || lower.contains("error sending request")
}

fn format_request_error(e: reqwest::Error) -> String {
    let url = e.url().map(|u| u.to_string()).unwrap_or_default();
    let mut parts = vec![format!("request failed: {e}")];
    if !url.is_empty() {
        parts.push(format!("url: {url}"));
    }
    if e.is_timeout() {
        parts.push("hint: request timed out — check network or try again".into());
    } else if e.is_connect() {
        parts.push(
            "hint: could not connect — check internet/VPN, firewall, and MORPH_AI_BASE_URL".into(),
        );
    } else if e.is_request() {
        parts.push(
            "hint: connection dropped while uploading the request — usually oversized chat/tool payload or network instability; retry after backend restart"
                .into(),
        );
    }
    parts.join("; ")
}
