use reqwest::Client;
use serde::Deserialize;
use std::time::Duration;

const ROLE_ADMIN: &str = "Admin";
const ROLE_SHARP_REPORTS: &str = "Sharp Reports";
const PERM_VIEW_REPORTS: &str = "view_reports";
const PERM_EXPORT_REPORTS: &str = "export_reports";

#[derive(Debug, Clone)]
pub struct UsersPanelClient {
    base_url: String,
    http: Client,
}

#[derive(Debug, Deserialize)]
struct LoginResponse {
    token: String,
    user: UsersPanelUser,
}

#[derive(Debug, Clone, Deserialize)]
pub struct UsersPanelUser {
    pub email: String,
    pub username: String,
    pub roles: Vec<String>,
    #[serde(default)]
    pub default_channel_id: Option<String>,
}

#[derive(Debug, Deserialize)]
struct UserEnvelope {
    user: UsersPanelUser,
}

#[derive(Debug, Deserialize)]
struct PermissionsEnvelope {
    permissions: Vec<String>,
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
    Unavailable,
    BadResponse,
}

impl std::fmt::Display for UsersPanelError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::InvalidCredentials => write!(f, "invalid credentials"),
            Self::Forbidden => write!(f, "sharp reports access is restricted by admin policy"),
            Self::Unavailable => write!(f, "userspanel unavailable"),
            Self::BadResponse => write!(f, "invalid userspanel response"),
        }
    }
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

    pub async fn login(
        &self,
        email: &str,
        password: &str,
    ) -> Result<SessionContext, UsersPanelError> {
        let body = serde_json::json!({
            "email": email.trim(),
            "password": password,
        });
        let resp = self
            .http
            .post(format!("{}/api/auth/login", self.base_url))
            .json(&body)
            .send()
            .await
            .map_err(|e| {
                tracing::warn!(?e, url = %self.base_url, "UsersPanel login request failed");
                UsersPanelError::Unavailable
            })?;

        let status = resp.status();
        if status == reqwest::StatusCode::UNAUTHORIZED {
            return Err(UsersPanelError::InvalidCredentials);
        }
        if status == reqwest::StatusCode::FORBIDDEN {
            return Err(UsersPanelError::Forbidden);
        }
        if !status.is_success() {
            let body = resp.text().await.unwrap_or_default();
            tracing::warn!(
                status = %status,
                body = %body,
                url = %self.base_url,
                "UsersPanel login failed"
            );
            return Err(UsersPanelError::Unavailable);
        }

        let login: LoginResponse = resp.json().await.map_err(|e| {
            tracing::warn!(?e, "UsersPanel login JSON parse failed");
            UsersPanelError::BadResponse
        })?;
        if login.token.trim().is_empty() {
            return Err(UsersPanelError::BadResponse);
        }

        let permissions = self.permissions(&login.token).await.unwrap_or_default();
        if !has_sharp_report_access(&login.user.roles, &permissions) {
            return Err(UsersPanelError::Forbidden);
        }

        Ok(SessionContext {
            token: login.token,
            user: login.user,
            permissions,
        })
    }

    pub async fn session(&self, token: &str) -> Result<SessionContext, UsersPanelError> {
        let user = self.user(token).await?;
        let permissions = self.permissions(token).await.unwrap_or_default();
        if !has_sharp_report_access(&user.roles, &permissions) {
            return Err(UsersPanelError::Forbidden);
        }
        Ok(SessionContext {
            token: token.to_string(),
            user,
            permissions,
        })
    }

    async fn user(&self, token: &str) -> Result<UsersPanelUser, UsersPanelError> {
        let resp = self
            .http
            .get(format!("{}/api/auth/user", self.base_url))
            .bearer_auth(token)
            .send()
            .await
            .map_err(|e| {
                tracing::warn!(?e, url = %self.base_url, "UsersPanel login request failed");
                UsersPanelError::Unavailable
            })?;

        if resp.status() == reqwest::StatusCode::UNAUTHORIZED {
            return Err(UsersPanelError::InvalidCredentials);
        }
        if !resp.status().is_success() {
            return Err(UsersPanelError::Unavailable);
        }

        let envelope: UserEnvelope = resp
            .json()
            .await
            .map_err(|_| UsersPanelError::BadResponse)?;
        Ok(envelope.user)
    }

    async fn permissions(&self, token: &str) -> Result<Vec<String>, UsersPanelError> {
        let resp = self
            .http
            .get(format!("{}/api/auth/permissions", self.base_url))
            .bearer_auth(token)
            .send()
            .await
            .map_err(|e| {
                tracing::warn!(?e, url = %self.base_url, "UsersPanel login request failed");
                UsersPanelError::Unavailable
            })?;

        if !resp.status().is_success() {
            return Err(UsersPanelError::Unavailable);
        }

        let envelope: PermissionsEnvelope = resp
            .json()
            .await
            .map_err(|_| UsersPanelError::BadResponse)?;
        Ok(envelope.permissions)
    }
}

pub fn resolve_role(roles: &[String]) -> String {
    for r in roles {
        if r == ROLE_ADMIN {
            return "admin".to_string();
        }
    }
    for r in roles {
        if r == ROLE_SHARP_REPORTS {
            return "employee".to_string();
        }
    }
    "viewer".to_string()
}

pub fn has_sharp_report_access(_roles: &[String], _permissions: &[String]) -> bool {
    // Morph-hosted auth: any authenticated user may use DataX (no app permission gates).
    true
}
