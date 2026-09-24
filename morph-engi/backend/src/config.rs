use std::env;

#[derive(Clone, Debug)]
pub struct Settings {
    pub database_url: String,
    pub jwt_secret: String,
    pub jwt_access_expiry_min: i64,
    pub app_env: String,
    pub app_port: u16,
    pub cors_origin: String,
    pub users_panel_base_url: String,
    pub static_dir: String,
    pub upload_dir: String,
    pub preview_demo: bool,
}

impl Settings {
    pub fn from_env() -> Result<Self, String> {
        morphai::load_repo_dotenv();

        Ok(Self {
            database_url: env::var("MORPH_ENGI_DATABASE_URL")
                .unwrap_or_else(|_| "sqlite://morph_engi.db".into()),
            jwt_secret: env::var("JWT_SECRET").unwrap_or_else(|_| "dev-morph-engi-secret".into()),
            jwt_access_expiry_min: env::var("MORPH_ENGI_JWT_ACCESS_EXPIRY_MIN")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(480),
            app_env: env::var("APP_ENV").unwrap_or_else(|_| "development".into()),
            app_port: env::var("MORPH_ENGI_PORT")
                .ok()
                .or_else(|| env::var("PORT").ok())
                .and_then(|v| v.parse().ok())
                .unwrap_or(9096),
            cors_origin: env::var("MORPH_ENGI_CORS_ORIGIN")
                .unwrap_or_else(|_| "http://localhost:5179".into()),
            users_panel_base_url: env::var("USERS_PANEL_BASE_URL")
                .or_else(|_| env::var("MORPH_AUTH_BASE_URL"))
                .unwrap_or_else(|_| "http://127.0.0.1:9090".into()),
            static_dir: env::var("STATIC_DIR").unwrap_or_default(),
            upload_dir: upload_dir_from(env::var("MORPH_ENGI_UPLOAD_DIR").ok().as_deref()),
            preview_demo: env::var("PREVIEW_DEMO")
                .map(|v| v == "1" || v.eq_ignore_ascii_case("true"))
                .unwrap_or(false)
                || env::var("APP_ENV")
                    .map(|v| v.eq_ignore_ascii_case("preview"))
                    .unwrap_or(false),
        })
    }

    pub fn is_development(&self) -> bool {
        self.app_env.eq_ignore_ascii_case("development")
    }
}

/// Upload directory from `MORPH_ENGI_UPLOAD_DIR`. Blank means the local `uploads` folder.
pub fn upload_dir_from(value: Option<&str>) -> String {
    match value.map(str::trim).filter(|s| !s.is_empty()) {
        Some(dir) => dir.to_string(),
        None => "uploads".into(),
    }
}

fn sqlite_file_path(database_url: &str) -> &str {
    let rest = database_url.strip_prefix("sqlite://").unwrap_or(database_url);
    rest.split('?').next().unwrap_or(rest)
}

fn under_data(path: &str) -> bool {
    path == "/data" || path.starts_with("/data/")
}

fn origin_host(origin: &str) -> &str {
    let trimmed = origin.trim();
    let rest = trimmed.split_once("://").map(|(_, host)| host).unwrap_or(trimmed);
    let hostport = rest.split('/').next().unwrap_or(rest);
    if let Some(inside) = hostport.strip_prefix('[') {
        return inside.split(']').next().unwrap_or(inside);
    }
    hostport.split(':').next().unwrap_or(hostport)
}

fn loopback_host(host: &str) -> bool {
    matches!(
        host.trim().to_ascii_lowercase().as_str(),
        "" | "localhost" | "127.0.0.1" | "::1" | "0.0.0.0"
    )
}

fn jwt_rejected(secret: &str) -> bool {
    let secret = secret.trim();
    secret.is_empty()
        || secret.len() < 32
        || secret == "dev-morph-engi-secret"
        || secret == "morph-dev-jwt-secret-change-me"
}

/// Why a production process must not listen. `None` means start.
pub fn production_block_reason(settings: &Settings) -> Option<&'static str> {
    if !settings.app_env.eq_ignore_ascii_case("production") {
        return None;
    }
    if jwt_rejected(&settings.jwt_secret) {
        return Some("JWT_SECRET is missing or is a development secret");
    }
    if loopback_host(origin_host(&settings.users_panel_base_url)) {
        return Some("USERS_PANEL_BASE_URL must be the public Morph API origin");
    }
    if !under_data(sqlite_file_path(&settings.database_url)) {
        return Some("MORPH_ENGI_DATABASE_URL must be an absolute sqlite path under /data");
    }
    if !under_data(settings.upload_dir.trim()) {
        return Some("MORPH_ENGI_UPLOAD_DIR must be an absolute path under /data");
    }
    None
}

#[cfg(test)]
mod tests {
    use super::*;

    fn production() -> Settings {
        Settings {
            database_url: "sqlite:///data/morph_engi.db".into(),
            jwt_secret: "x".repeat(32),
            jwt_access_expiry_min: 480,
            app_env: "production".into(),
            app_port: 9096,
            cors_origin: String::new(),
            users_panel_base_url: "https://morph.example.com".into(),
            static_dir: "/app/frontend/dist".into(),
            upload_dir: "/data/uploads".into(),
            preview_demo: false,
        }
    }

    #[test]
    fn upload_dir_defaults_when_unset_or_blank() {
        assert_eq!(upload_dir_from(None), "uploads");
        assert_eq!(upload_dir_from(Some("  ")), "uploads");
    }

    #[test]
    fn upload_dir_uses_configured_value() {
        assert_eq!(upload_dir_from(Some("/data/uploads")), "/data/uploads");
    }

    #[test]
    fn production_rejects_dev_jwt() {
        for secret in ["dev-morph-engi-secret", "morph-dev-jwt-secret-change-me", "short"] {
            let mut settings = production();
            settings.jwt_secret = secret.into();
            let reason = production_block_reason(&settings).expect(secret);
            assert!(reason.contains("JWT_SECRET"), "{secret}: {reason}");
        }
    }

    #[test]
    fn production_rejects_loopback_morph_api() {
        for origin in ["", "http://127.0.0.1:9090", "http://localhost:9090", "http://[::1]:9090"] {
            let mut settings = production();
            settings.users_panel_base_url = origin.into();
            let reason = production_block_reason(&settings).expect(origin);
            assert!(reason.contains("USERS_PANEL_BASE_URL"), "{origin}: {reason}");
        }
    }

    #[test]
    fn production_rejects_storage_outside_data() {
        let mut settings = production();
        settings.database_url = "sqlite://morph_engi.db".into();
        let reason = production_block_reason(&settings).expect("relative db");
        assert!(reason.contains("MORPH_ENGI_DATABASE_URL"), "{reason}");

        settings.database_url = "sqlite:///data/morph_engi.db".into();
        settings.upload_dir = "uploads".into();
        let reason = production_block_reason(&settings).expect("relative uploads");
        assert!(reason.contains("MORPH_ENGI_UPLOAD_DIR"), "{reason}");
    }

    #[test]
    fn development_allows_local_defaults() {
        let mut settings = production();
        settings.app_env = "development".into();
        settings.jwt_secret = "dev-morph-engi-secret".into();
        settings.database_url = "sqlite://morph_engi.db".into();
        settings.upload_dir = "uploads".into();
        settings.users_panel_base_url = "http://127.0.0.1:9090".into();
        assert_eq!(production_block_reason(&settings), None);
    }

    #[test]
    fn production_accepts_public_morph_and_data_paths() {
        assert_eq!(production_block_reason(&production()), None);
    }
}
