use std::env;

#[derive(Clone, Debug)]
pub struct Config {
    pub database_url: String,
    pub jwt_secret: String,
    pub jwt_expiry_hours: i64,
    pub host: String,
    pub port: u16,
    pub frontend_origin: String,
    /// Base URL for links in emails (verify, reset), e.g. http://127.0.0.1:5001
    pub api_public_url: String,
    pub google_client_id: Option<String>,
    pub google_client_secret: Option<String>,
    pub google_redirect_url: String,
    pub bootstrap_admin_email: Option<String>,
    pub bootstrap_admin_password: Option<String>,
}

impl Config {
    pub fn from_env() -> Self {
        // Dev default: MySQL on localhost, database `tran`. Override with DATABASE_URL in production.
        // @ in password is URL-encoded as %40
        let database_url = env::var("DATABASE_URL").unwrap_or_else(|_| {
            "mysql://root:Dafuq%40911@127.0.0.1:3306/tran?charset=utf8mb4".to_string()
        });
        let jwt_secret = env::var("JWT_SECRET").unwrap_or_else(|_| {
            tracing::warn!("JWT_SECRET not set; using default (not for production)");
            "dev-secret-change-me".to_string()
        });
        // Default: no practical session timeout (~100 years). Override with JWT_EXPIRY_HOURS.
        let jwt_expiry_hours: i64 = env::var("JWT_EXPIRY_HOURS")
            .ok()
            .and_then(|s| s.parse().ok())
            .unwrap_or(100 * 365 * 24);
        let host = env::var("HOST").unwrap_or_else(|_| "127.0.0.1".to_string());
        let port: u16 = env::var("PORT")
            .ok()
            .and_then(|s| s.parse().ok())
            .unwrap_or(5001);
        let frontend_origin =
            env::var("FRONTEND_ORIGIN").unwrap_or_else(|_| "http://localhost:5173".to_string());
        let api_public_url = env::var("API_PUBLIC_URL").unwrap_or_else(|_| {
            format!("http://{}:{}", host, port)
        });
        let google_client_id = env::var("GOOGLE_CLIENT_ID").ok();
        let google_client_secret = env::var("GOOGLE_CLIENT_SECRET").ok();
        let google_redirect_url = env::var("GOOGLE_REDIRECT_URL").unwrap_or_else(|_| {
            format!("http://{}:{}/api/auth/google/callback", host, port)
        });
        // When both bootstrap vars are unset, create a one-time dev Admin (opt-out: USERS_PANEL_NO_DEFAULT_ADMIN=1).
        // TranMail, TranForm, TranDemo, etc. authenticate via this service — use the same email/password everywhere locally.
        let bootstrap_email_env = env::var("BOOTSTRAP_ADMIN_EMAIL").ok();
        let bootstrap_password_env = env::var("BOOTSTRAP_ADMIN_PASSWORD").ok();
        let no_default = env::var("USERS_PANEL_NO_DEFAULT_ADMIN")
            .map(|v| v == "1" || v.eq_ignore_ascii_case("true"))
            .unwrap_or(false);

        let (bootstrap_admin_email, bootstrap_admin_password) = match (
            bootstrap_email_env.as_ref(),
            bootstrap_password_env.as_ref(),
        ) {
            (Some(email), Some(pass)) => (Some(email.clone()), Some(pass.clone())),
            (None, None) if !no_default => {
                tracing::info!(
                    "Using dev default bootstrap Admin: admin@example.com (set BOOTSTRAP_ADMIN_EMAIL/PASSWORD to override; USERS_PANEL_NO_DEFAULT_ADMIN=1 to disable)"
                );
                (
                    Some("admin@example.com".to_string()),
                    Some("AdminExample2026!".to_string()),
                )
            }
            (None, None) => (None, None),
            _ => {
                tracing::warn!(
                    "Set both BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_PASSWORD together, or leave both unset for dev defaults"
                );
                (bootstrap_email_env, bootstrap_password_env)
            }
        };

        Self {
            database_url,
            jwt_secret,
            jwt_expiry_hours,
            host,
            port,
            frontend_origin,
            api_public_url,
            google_client_id,
            google_client_secret,
            google_redirect_url,
            bootstrap_admin_email,
            bootstrap_admin_password,
        }
    }
}
