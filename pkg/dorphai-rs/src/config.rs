use std::path::{Path, PathBuf};

const DEFAULT_MODEL: &str = "qwen3-max";
const DEFAULT_NVIDIA_MODEL: &str = "nvidia/nemotron-3-ultra-550b-a55b";
const DEFAULT_NVIDIA_BASE: &str = "https://integrate.api.nvidia.com/v1";
const DEFAULT_SILICONFLOW_BASE: &str = "https://api.siliconflow.cn/v1";
const DEFAULT_API_URL: &str =
    "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation";
const DEFAULT_BASE_URL: &str = "https://dashscope.aliyuncs.com/compatible-mode/v1";

/// DorphAI model settings loaded from the environment.
#[derive(Debug, Clone)]
pub struct Config {
    pub api_key: String,
    pub model: String,
    /// DashScope native text-generation endpoint (used when `base_url` is unset).
    pub api_url: String,
    /// OpenAI-compatible `/v1` base (e.g. SiliconFlow, DashScope compatible-mode).
    pub base_url: Option<String>,
    /// Named provider id when loaded from role env (e.g. `nvidia`, `siliconflow`).
    pub provider: Option<String>,
    pub nvidia_enable_thinking: bool,
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

/// DORPH_* first, then the matching MORPH_* name.
pub fn env_or_legacy(key: &str) -> Option<String> {
    std::env::var(key)
        .ok()
        .filter(|s| !s.trim().is_empty())
        .or_else(|| {
            let alias = key.replacen("DORPH_", "MORPH_", 1);
            if alias == key {
                None
            } else {
                std::env::var(&alias).ok().filter(|s| !s.trim().is_empty())
            }
        })
}

impl Config {
    pub fn from_env() -> Self {
        load_repo_dotenv();

        let api_key = env_or_legacy("DORPH_AI_API_KEY")
            .or_else(|| std::env::var("GEMINI_API_KEY").ok())
            .unwrap_or_default();

        let model = env_or_legacy("DORPH_AI_MODEL")
            .or_else(|| std::env::var("GEMINI_MODEL").ok())
            .unwrap_or_else(|| DEFAULT_MODEL.to_string());

        let base_url = env_or_legacy("DORPH_AI_BASE_URL")
            .map(|s| s.trim().trim_end_matches('/').to_string());

        let api_url = env_or_legacy("DORPH_AI_API_URL")
            .unwrap_or_else(|| DEFAULT_API_URL.to_string());

        Self {
            api_key: api_key.trim().to_string(),
            model: model.trim().to_string(),
            api_url: api_url.trim().trim_end_matches('/').to_string(),
            base_url,
            provider: None,
            nvidia_enable_thinking: nvidia_enable_thinking_from_env(),
        }
    }

    /// Primary agent profile: `DORPH_AI_ORCHESTRATOR_PROVIDER` when set, else `from_env()`.
    pub fn from_orchestrator_env() -> Self {
        load_repo_dotenv();
        let orch = env_or_legacy("DORPH_AI_ORCHESTRATOR_PROVIDER")
            .map(|s| s.trim().to_lowercase())
            .filter(|s| !s.is_empty());
        match orch.as_deref() {
            Some("nvidia") => Self::from_nvidia_orchestrator(),
            Some("siliconflow") => Self::from_siliconflow_orchestrator(),
            Some(other) => {
                let mut cfg = Self::from_env();
                cfg.provider = Some(other.to_string());
                cfg
            }
            None => Self::from_env(),
        }
    }

    fn from_nvidia_orchestrator() -> Self {
        let api_key = std::env::var("NVIDIA_API_KEY")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .unwrap_or_default();
        let model = std::env::var("NVIDIA_MODEL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .unwrap_or_else(|| DEFAULT_NVIDIA_MODEL.to_string());
        let base_url = std::env::var("NVIDIA_BASE_URL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .map(|s| s.trim().trim_end_matches('/').to_string())
            .or_else(|| Some(DEFAULT_NVIDIA_BASE.to_string()));
        Self {
            api_key: api_key.trim().to_string(),
            model: model.trim().to_string(),
            api_url: DEFAULT_API_URL.to_string(),
            base_url,
            provider: Some("nvidia".to_string()),
            nvidia_enable_thinking: nvidia_enable_thinking_from_env(),
        }
    }

    fn from_siliconflow_orchestrator() -> Self {
        let api_key = std::env::var("SILICONFLOW_API_KEY")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .unwrap_or_default();
        let model = std::env::var("SILICONFLOW_MODEL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .unwrap_or_else(|| "Qwen/Qwen3.6-27B".to_string());
        let base_url = std::env::var("SILICONFLOW_BASE_URL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .map(|s| s.trim().trim_end_matches('/').to_string())
            .or_else(|| Some(DEFAULT_SILICONFLOW_BASE.to_string()));
        Self {
            api_key: api_key.trim().to_string(),
            model: model.trim().to_string(),
            api_url: DEFAULT_API_URL.to_string(),
            base_url,
            provider: Some("siliconflow".to_string()),
            nvidia_enable_thinking: false,
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

    /// Message AI provider settings (`MESSAGE_AI_*`), falling back to `DORPH_AI_*`.
    pub fn from_message_ai_env() -> Self {
        let api_key = std::env::var("MESSAGE_AI_API_KEY")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .or_else(|| env_or_legacy("DORPH_AI_API_KEY"))
            .or_else(|| std::env::var("GEMINI_API_KEY").ok())
            .unwrap_or_default();

        let model = std::env::var("MESSAGE_AI_MODEL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .or_else(|| env_or_legacy("DORPH_AI_MODEL"))
            .or_else(|| std::env::var("GEMINI_MODEL").ok())
            .unwrap_or_else(|| DEFAULT_MODEL.to_string());

        let base_url = std::env::var("MESSAGE_AI_BASE_URL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .or_else(|| env_or_legacy("DORPH_AI_BASE_URL"))
            .map(|s| s.trim().trim_end_matches('/').to_string());

        let api_url = std::env::var("MESSAGE_AI_API_URL")
            .ok()
            .filter(|s| !s.trim().is_empty())
            .or_else(|| env_or_legacy("DORPH_AI_API_URL"))
            .unwrap_or_else(|| DEFAULT_API_URL.to_string());

        Self {
            api_key: api_key.trim().to_string(),
            model: model.trim().to_string(),
            api_url: api_url.trim().trim_end_matches('/').to_string(),
            base_url,
            provider: None,
            nvidia_enable_thinking: false,
        }
    }
}

fn nvidia_enable_thinking_from_env() -> bool {
    match std::env::var("NVIDIA_ENABLE_THINKING") {
        Ok(v) => {
            let v = v.trim().to_lowercase();
            v != "false" && v != "0" && v != "no"
        }
        Err(_) => true,
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
    fn env_or_legacy_reads_morph_when_dorph_is_empty() {
        std::env::remove_var("DORPH_TESTONLY_KEY");
        std::env::set_var("MORPH_TESTONLY_KEY", "from-old");
        assert_eq!(env_or_legacy("DORPH_TESTONLY_KEY").as_deref(), Some("from-old"));
        std::env::set_var("DORPH_TESTONLY_KEY", "from-new");
        assert_eq!(env_or_legacy("DORPH_TESTONLY_KEY").as_deref(), Some("from-new"));
        std::env::remove_var("DORPH_TESTONLY_KEY");
        std::env::remove_var("MORPH_TESTONLY_KEY");
    }

    #[test]
    fn finds_start_all_not_nested_env() {
        let nanos = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_nanos();
        let base = std::env::temp_dir().join(format!("dorphai-repoenv-{nanos}"));
        let nested = base.join("formx").join("backend");
        std::fs::create_dir_all(&nested).unwrap();
        std::fs::write(base.join("start-all.sh"), "#!/bin/sh\n").unwrap();
        std::fs::write(base.join(".env"), "DORPH_AI_API_KEY=from-root\n").unwrap();
        std::fs::write(nested.join(".env"), "DORPH_AI_API_KEY=\n").unwrap();
        let root = find_repo_root(&nested).expect("repo root");
        assert!(root.join("start-all.sh").is_file());
        let nested_env = std::fs::read_to_string(nested.join(".env")).unwrap();
        assert!(nested_env.contains("DORPH_AI_API_KEY="));
        let root_env = std::fs::read_to_string(root.join(".env")).unwrap();
        assert!(root_env.contains("from-root"));
        let _ = std::fs::remove_dir_all(&base);
    }

    #[test]
    fn orchestrator_nvidia_env() {
        std::env::set_var("DORPH_AI_ORCHESTRATOR_PROVIDER", "nvidia");
        std::env::set_var("NVIDIA_API_KEY", "nv-test");
        std::env::remove_var("NVIDIA_MODEL");
        let cfg = Config::from_orchestrator_env();
        assert_eq!(cfg.provider.as_deref(), Some("nvidia"));
        assert_eq!(cfg.api_key, "nv-test");
        assert_eq!(cfg.model, DEFAULT_NVIDIA_MODEL);
        assert!(cfg.nvidia_enable_thinking);
        std::env::remove_var("DORPH_AI_ORCHESTRATOR_PROVIDER");
        std::env::remove_var("NVIDIA_API_KEY");
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
