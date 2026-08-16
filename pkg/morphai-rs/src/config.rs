const DEFAULT_MODEL: &str = "qwen3-max";
const DEFAULT_API_URL: &str =
    "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation";
const DEFAULT_BASE_URL: &str = "https://dashscope.aliyuncs.com/compatible-mode/v1";

/// MorphAI model settings loaded from the environment.
#[derive(Debug, Clone)]
pub struct Config {
    pub api_key: String,
    pub model: String,
    /// DashScope native text-generation endpoint (used when `base_url` is unset).
    pub api_url: String,
    /// OpenAI-compatible `/v1` base (e.g. SiliconFlow, DashScope compatible-mode).
    pub base_url: Option<String>,
}

/// Load `.env` from cwd and parent (e.g. `SharpReport/.env` when run from `backend/`).
fn load_dotenv_files() {
    let _ = dotenvy::dotenv();
    if let Ok(cwd) = std::env::current_dir() {
        let local = cwd.join(".env");
        if local.is_file() {
            let _ = dotenvy::from_path(&local);
        }
        if let Some(parent) = cwd.parent() {
            let parent_env = parent.join(".env");
            if parent_env.is_file() {
                let _ = dotenvy::from_path(&parent_env);
            }
        }
    }
}

impl Config {
    pub fn from_env() -> Self {
        load_dotenv_files();

        let api_key = std::env::var("MORPH_AI_API_KEY")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .or_else(|| std::env::var("GEMINI_API_KEY").ok())
            .unwrap_or_default();

        let model = std::env::var("MORPH_AI_MODEL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .or_else(|| std::env::var("GEMINI_MODEL").ok())
            .unwrap_or_else(|| DEFAULT_MODEL.to_string());

        let base_url = std::env::var("MORPH_AI_BASE_URL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .map(|s| s.trim().trim_end_matches('/').to_string());

        let api_url = std::env::var("MORPH_AI_API_URL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .unwrap_or_else(|| DEFAULT_API_URL.to_string());

        Self {
            api_key: api_key.trim().to_string(),
            model: model.trim().to_string(),
            api_url: api_url.trim().trim_end_matches('/').to_string(),
            base_url,
        }
    }

    pub fn configured(&self) -> bool {
        !self.api_key.is_empty()
    }

    pub fn uses_openai_compatible(&self) -> bool {
        self.base_url.is_some()
    }

    pub fn chat_completions_url(&self) -> String {
        let base = self
            .base_url
            .as_deref()
            .filter(|s| !s.is_empty())
            .unwrap_or(DEFAULT_BASE_URL);
        format!("{}/chat/completions", base.trim_end_matches('/'))
    }

    /// Message AI provider settings (`MESSAGE_AI_*`), falling back to `MORPH_AI_*`.
    pub fn from_message_ai_env() -> Self {
        let api_key = std::env::var("MESSAGE_AI_API_KEY")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .or_else(|| std::env::var("MORPH_AI_API_KEY").ok())
            .or_else(|| std::env::var("GEMINI_API_KEY").ok())
            .unwrap_or_default();

        let model = std::env::var("MESSAGE_AI_MODEL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .or_else(|| std::env::var("MORPH_AI_MODEL").ok())
            .or_else(|| std::env::var("GEMINI_MODEL").ok())
            .unwrap_or_else(|| DEFAULT_MODEL.to_string());

        let base_url = std::env::var("MESSAGE_AI_BASE_URL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .or_else(|| std::env::var("MORPH_AI_BASE_URL").ok())
            .map(|s| s.trim().trim_end_matches('/').to_string());

        let api_url = std::env::var("MESSAGE_AI_API_URL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .or_else(|| std::env::var("MORPH_AI_API_URL").ok())
            .unwrap_or_else(|| DEFAULT_API_URL.to_string());

        Self {
            api_key: api_key.trim().to_string(),
            model: model.trim().to_string(),
            api_url: api_url.trim().trim_end_matches('/').to_string(),
            base_url,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn default_model_when_unset() {
        std::env::remove_var("MORPH_AI_MODEL");
        std::env::remove_var("GEMINI_MODEL");
        let cfg = Config::from_env();
        assert_eq!(cfg.model, DEFAULT_MODEL);
    }

    #[test]
    fn openai_url_from_base() {
        let mut cfg = Config::from_env();
        cfg.base_url = Some("https://api.siliconflow.cn/v1".to_string());
        assert_eq!(
            cfg.chat_completions_url(),
            "https://api.siliconflow.cn/v1/chat/completions"
        );
    }
}
