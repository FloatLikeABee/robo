use std::path::{Path, PathBuf};

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

/// Walk up from `start` until a directory containing `start-all.sh` is found.
pub fn find_repo_root(start: &Path) -> Option<PathBuf> {
    let mut dir = start.to_path_buf();
    if let Ok(canon) = dir.canonicalize() {
        dir = canon;
    }
    loop {
        if dir.join("start-all.sh").is_file() {
            return Some(dir);
        }
        if !dir.pop() {
            return None;
        }
    }
}

/// Load only the repository-root `.env` (next to `start-all.sh`). Nested `.env` files are ignored.
pub fn load_repo_dotenv() {
    let cwd = match std::env::current_dir() {
        Ok(d) => d,
        Err(_) => return,
    };
    if let Some(root) = find_repo_root(&cwd) {
        let env_path = root.join(".env");
        if env_path.is_file() {
            let _ = dotenvy::from_path(&env_path);
        }
    }
}

impl Config {
    pub fn from_env() -> Self {
        load_repo_dotenv();

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
    fn default_model_constant() {
        assert_eq!(DEFAULT_MODEL, "qwen3-max");
    }

    #[test]
    fn finds_start_all_not_nested_env() {
        let nanos = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_nanos();
        let base = std::env::temp_dir().join(format!("morphai-repoenv-{nanos}"));
        let nested = base.join("formx").join("backend");
        std::fs::create_dir_all(&nested).unwrap();
        std::fs::write(base.join("start-all.sh"), "#!/bin/sh\n").unwrap();
        std::fs::write(base.join(".env"), "MORPH_AI_API_KEY=from-root\n").unwrap();
        std::fs::write(nested.join(".env"), "MORPH_AI_API_KEY=\n").unwrap();
        let root = find_repo_root(&nested).expect("repo root");
        assert!(root.join("start-all.sh").is_file());
        let nested_env = std::fs::read_to_string(nested.join(".env")).unwrap();
        assert!(nested_env.contains("MORPH_AI_API_KEY="));
        let root_env = std::fs::read_to_string(root.join(".env")).unwrap();
        assert!(root_env.contains("from-root"));
        let _ = std::fs::remove_dir_all(&base);
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
