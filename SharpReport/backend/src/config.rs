use std::path::{Path, PathBuf};

use config::{Config, Environment, File};
use serde::Deserialize;

#[derive(Debug, Deserialize, Clone)]
pub struct Settings {
    pub server: ServerSettings,
    pub database: DatabaseSettings,
    pub jwt: JwtSettings,
    pub users_panel: UsersPanelSettings,
    pub metabase: MetabaseSettings,
    pub cors: CorsSettings,
    #[serde(default)]
    pub academi: AcademiSettings,
}

#[derive(Debug, Deserialize, Clone, Default)]
pub struct AcademiSettings {
    /// Base URL for Academi API (e.g. http://127.0.0.1:8978) — no `/api/v1` suffix.
    #[serde(default = "default_academi_base")]
    pub base_url: String,
}

fn default_academi_base() -> String {
    "http://127.0.0.1:8978".into()
}

#[derive(Debug, Deserialize, Clone)]
pub struct ServerSettings {
    pub host: String,
    pub port: u16,
}

#[derive(Debug, Deserialize, Clone)]
pub struct DatabaseSettings {
    pub url: String,
}

#[derive(Debug, Deserialize, Clone)]
pub struct JwtSettings {
    pub secret: String,
    pub expiry_hours: u64,
}

#[derive(Debug, Deserialize, Clone)]
pub struct UsersPanelSettings {
    /// Base URL for UsersPanel API (e.g. http://127.0.0.1:5001).
    pub base_url: String,
}

#[derive(Debug, Deserialize, Clone)]
pub struct MetabaseSettings {
    /// When true, download the JAR if missing and spawn `java -jar` on startup.
    /// When false, use an already-running Metabase at `host`:`port` (proxy only).
    #[serde(default)]
    pub autostart: bool,
    pub jar_path: String,
    pub download_url: String,
    pub host: String,
    pub port: u16,
    pub db_type: String,
    pub db_url: Option<String>,
    pub jvm_opts: String,
    pub health_check_interval: u64,
    pub restart_on_failure: bool,
    pub max_restarts: u32,
}

#[derive(Debug, Deserialize, Clone)]
pub struct CorsSettings {
    pub allowed_origins: Vec<String>,
}

/// Load `.env` from the backend directory and parent app directory (`SharpReport/.env`).
pub fn load_env_files() {
    if let Ok(cwd) = std::env::current_dir() {
        if let Some(parent) = cwd.parent() {
            let parent_env = parent.join(".env");
            if parent_env.is_file() {
                let _ = dotenvy::from_path(&parent_env);
            }
        }
        let local_env = cwd.join(".env");
        if local_env.is_file() {
            let _ = dotenvy::from_path(&local_env);
        }
    } else {
        let _ = dotenvy::dotenv();
    }
}

impl Settings {
    pub fn new() -> Result<Self, config::ConfigError> {
        load_env_files();
        let env = std::env::var("RUN_ENV").unwrap_or_else(|_| "development".into());
        let mut settings: Settings = Config::builder()
            .add_source(File::with_name("config/default").required(false))
            .add_source(File::with_name(&format!("config/{}", env)).required(false))
            .add_source(Environment::with_prefix("app").separator("_"))
            .build()?
            .try_deserialize()?;

        // Morph hosts auth (legacy name USERS_PANEL_BASE_URL kept for compatibility).
        if let Ok(url) = std::env::var("USERS_PANEL_BASE_URL") {
            let url = url.trim();
            if !url.is_empty() {
                settings.users_panel.base_url = url.trim_end_matches('/').to_string();
            }
        } else if let Ok(url) = std::env::var("MORPH_API_BASE_URL") {
            let url = url.trim();
            if !url.is_empty() {
                settings.users_panel.base_url = url.trim_end_matches('/').to_string();
            }
        }

        if let Ok(url) = std::env::var("ACADEMI_API_BASE_URL") {
            let url = url.trim();
            if !url.is_empty() {
                settings.academi.base_url = url.trim_end_matches('/').to_string();
            }
        }

        Ok(settings)
    }
}

/// Turn relative `sqlite:` URLs into absolute paths so the DB file is stable regardless of how the
/// binary is launched. Set `DATAPULSE_WORK_DIR` to pin the base directory (defaults to [`std::env::current_dir`]).
pub fn resolve_database_url(url: &str) -> String {
    let trimmed = url.trim();
    let lower = trimmed.to_lowercase();
    if !lower.starts_with("sqlite:") {
        return trimmed.to_string();
    }
    let (main, query) = match trimmed.split_once('?') {
        Some((a, b)) => (a, Some(b)),
        None => (trimmed, None),
    };
    let path_part = main
        .strip_prefix("sqlite://")
        .or_else(|| main.strip_prefix("sqlite:"))
        .unwrap_or(main);
    let path_part = path_part.trim_start_matches('/');
    if path_part.is_empty()
        || path_part.eq_ignore_ascii_case(":memory:")
        || lower.contains("mode=memory")
    {
        return trimmed.to_string();
    }
    let path = Path::new(path_part);
    if path.is_absolute() {
        return trimmed.to_string();
    }
    let base = std::env::var("DATAPULSE_WORK_DIR")
        .map(PathBuf::from)
        .or_else(|_| std::env::current_dir())
        .unwrap_or_else(|_| PathBuf::from("."));
    let resolved = base.join(path_part);
    let mut out = format!("sqlite://{}", resolved.display());
    if let Some(q) = query {
        out.push('?');
        out.push_str(q);
    }
    out
}

// Default configuration that will be merged with others
impl Default for Settings {
    fn default() -> Self {
        Self {
            server: ServerSettings {
                host: "0.0.0.0".into(),
                port: 3050,
            },
            database: DatabaseSettings {
                url: "sqlite://./data/datapulse.db".into(),
            },
            jwt: JwtSettings {
                secret: "change-me-in-production".into(),
                expiry_hours: 24,
            },
            users_panel: UsersPanelSettings {
                base_url: "http://127.0.0.1:9090".into(),
            },
            metabase: MetabaseSettings {
                autostart: false,
                jar_path: "./metabase/metabase.jar".into(),
                download_url: "https://downloads.metabase.com/latest/metabase.jar".into(),
                host: "127.0.0.1".into(),
                port: 8001,
                db_type: "h2".into(),
                db_url: None,
                jvm_opts: "-Xmx2g -Xms512m".into(),
                health_check_interval: 30,
                restart_on_failure: true,
                max_restarts: 5,
            },
            cors: CorsSettings {
                allowed_origins: vec![
                    "http://localhost:5173".into(),
                    "http://localhost:5178".into(),
                    "http://127.0.0.1:5173".into(),
                    "http://127.0.0.1:5178".into(),
                    "http://localhost:3050".into(),
                    "http://127.0.0.1:3050".into(),
                ],
            },
            academi: AcademiSettings {
                base_url: default_academi_base(),
            },
        }
    }
}
