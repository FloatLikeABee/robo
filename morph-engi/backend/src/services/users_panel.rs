//! Platform auth client — Morph Data `/api/auth/*` (UsersPanel-compatible).

use reqwest::Client;
use serde::Deserialize;
use std::time::Duration;

const PERM_ENGI: &str = "morph_engi";
const PLATFORM_ROLE_ADMIN: &str = "Admin";

#[derive(Debug, Clone)]
pub struct UsersPanelClient {
    base_url: String,
    http: Client,
}

#[derive(Debug, Clone, Deserialize)]
pub struct UsersPanelUser {
    pub email: String,
    pub username: String,
    pub roles: Vec<String>,
}

#[derive(Debug, Clone)]
pub struct SessionContext {
    pub token: String,
    pub user: UsersPanelUser,
    pub permissions: Vec<String>,
}

#[derive(Debug)]
pub enum UsersPanelError {
    InvalidCredentials,
    Forbidden,
    Unavailable(String),
}

#[derive(Deserialize)]
struct LoginResponse {
    token: String,
    user: UsersPanelUser,
}

#[derive(Deserialize)]
struct UserEnvelope {
    user: UsersPanelUser,
}

#[derive(Deserialize)]
struct PermissionsEnvelope {
    permissions: Vec<String>,
}

#[derive(Deserialize)]
struct ErrorEnvelope {
    error: Option<String>,
}

impl UsersPanelClient {
    pub fn new(base_url: impl Into<String>) -> Self {
        let http = Client::builder()
            .timeout(Duration::from_secs(15))
            .no_proxy()
            .build()
            .unwrap_or_else(|_| Client::new());
        Self {
            base_url: base_url.into().trim_end_matches('/').to_string(),
            http,
        }
    }

    pub async fn login(&self, email: &str, password: &str) -> Result<SessionContext, UsersPanelError> {
        let resp = self
            .http
            .post(format!("{}/api/auth/login", self.base_url))
            .json(&serde_json::json!({ "email": email.trim(), "password": password }))
            .send()
            .await
            .map_err(|e| UsersPanelError::Unavailable(format!("Morph auth unreachable at {} ({e})", self.base_url)))?;

        let status = resp.status();
        if status.is_success() {
            let body: LoginResponse = resp.json().await.map_err(|_| {
                UsersPanelError::Unavailable("Invalid auth response from Morph".into())
            })?;
            let permissions = self.fetch_permissions(&body.token).await?;
            if !has_engi_access(&body.user.roles, &permissions) {
                return Err(UsersPanelError::Forbidden);
            }
            return Ok(SessionContext {
                token: body.token,
                user: body.user,
                permissions,
            });
        }

        let msg = parse_error_body(resp).await;
        if status.as_u16() == 401 {
            return Err(UsersPanelError::InvalidCredentials);
        }
        if status.as_u16() == 403 {
            return Err(UsersPanelError::Forbidden);
        }
        Err(UsersPanelError::Unavailable(msg))
    }

    pub async fn session(&self, token: &str) -> Result<SessionContext, UsersPanelError> {
        let resp = self
            .http
            .get(format!("{}/api/auth/user", self.base_url))
            .header("Authorization", format!("Bearer {}", token))
            .send()
            .await
            .map_err(|e| UsersPanelError::Unavailable(format!("Morph auth unreachable ({e})")))?;

        if !resp.status().is_success() {
            return Err(UsersPanelError::InvalidCredentials);
        }
        let body: UserEnvelope = resp.json().await.map_err(|_| {
            UsersPanelError::Unavailable("Invalid user response from Morph".into())
        })?;
        let permissions = self.fetch_permissions(token).await?;
        if !has_engi_access(&body.user.roles, &permissions) {
            return Err(UsersPanelError::Forbidden);
        }
        Ok(SessionContext {
            token: token.to_string(),
            user: body.user,
            permissions,
        })
    }

    async fn fetch_permissions(&self, token: &str) -> Result<Vec<String>, UsersPanelError> {
        let resp = self
            .http
            .get(format!("{}/api/auth/permissions", self.base_url))
            .header("Authorization", format!("Bearer {}", token))
            .send()
            .await
            .map_err(|e| UsersPanelError::Unavailable(format!("Morph auth unreachable ({e})")))?;
        if !resp.status().is_success() {
            return Ok(vec![]);
        }
        let body: PermissionsEnvelope = resp.json().await.unwrap_or(PermissionsEnvelope {
            permissions: vec![],
        });
        Ok(body.permissions)
    }
}

async fn parse_error_body(resp: reqwest::Response) -> String {
    let status = resp.status();
    let text = resp.text().await.unwrap_or_default();
    if let Ok(w) = serde_json::from_str::<ErrorEnvelope>(&text) {
        if let Some(e) = w.error.filter(|s| !s.is_empty()) {
            return e;
        }
    }
    if text.trim().is_empty() {
        format!("Morph auth error ({status})")
    } else {
        text.chars().take(240).collect()
    }
}

pub fn has_engi_access(roles: &[String], permissions: &[String]) -> bool {
    for r in roles {
        if r.eq_ignore_ascii_case(PLATFORM_ROLE_ADMIN) || r.eq_ignore_ascii_case("admin") {
            return true;
        }
    }
    for p in permissions {
        if p.eq_ignore_ascii_case(PERM_ENGI) {
            return true;
        }
    }
    false
}

pub fn map_engi_role(platform_roles: &[String]) -> &'static str {
    for r in platform_roles {
        if r.eq_ignore_ascii_case(PLATFORM_ROLE_ADMIN) || r.eq_ignore_ascii_case("admin") {
            return "owner";
        }
    }
    "manager"
}
