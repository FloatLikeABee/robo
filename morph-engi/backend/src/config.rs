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
    pub preview_demo: bool,
}

impl Settings {
    pub fn from_env() -> Result<Self, String> {
        let _ = dotenvy::dotenv();
        if let Ok(cwd) = env::current_dir() {
            for candidate in [
                cwd.join(".env"),
                cwd.join("..").join(".env"),
                cwd.join("../.env"),
            ] {
                if candidate.exists() {
                    let _ = dotenvy::from_path(&candidate);
                }
            }
        }

        Ok(Self {
            database_url: env::var("DATABASE_URL")
                .unwrap_or_else(|_| "sqlite://morph_engi.db".into()),
            jwt_secret: env::var("JWT_SECRET").unwrap_or_else(|_| "dev-morph-engi-secret".into()),
            jwt_access_expiry_min: env::var("JWT_ACCESS_EXPIRY_MIN")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(480),
            app_env: env::var("APP_ENV").unwrap_or_else(|_| "development".into()),
            app_port: env::var("APP_PORT")
                .ok()
                .or_else(|| env::var("PORT").ok())
                .and_then(|v| v.parse().ok())
                .unwrap_or(9096),
            cors_origin: env::var("CORS_ORIGIN")
                .unwrap_or_else(|_| "http://localhost:5179".into()),
            users_panel_base_url: env::var("USERS_PANEL_BASE_URL")
                .or_else(|_| env::var("MORPH_AUTH_BASE_URL"))
                .unwrap_or_else(|_| "http://127.0.0.1:9090".into()),
            static_dir: env::var("STATIC_DIR").unwrap_or_default(),
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
