use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use utoipa::ToSchema;

#[derive(Debug, Clone, FromRow)]
pub struct UserRow {
    pub id: String,
    pub email: String,
    pub username: String,
    pub password_hash: Option<String>,
    pub google_id: Option<String>,
    pub is_verified: bool,
    pub roles: String,
    pub permissions: String,
    pub default_channel_id: String,
    pub verification_token: Option<String>,
    pub verification_expires_at: Option<String>,
    pub reset_token: Option<String>,
    pub reset_expires_at: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize, ToSchema)]
pub struct UserPublic {
    pub email: String,
    pub username: String,
    pub roles: Vec<String>,
    pub permissions: Vec<String>,
    pub default_channel_id: String,
}

impl UserRow {
    pub fn roles_vec(&self) -> serde_json::Result<Vec<String>> {
        serde_json::from_str(&self.roles)
    }

    pub fn permissions_vec(&self) -> serde_json::Result<Vec<String>> {
        serde_json::from_str(&self.permissions)
    }

    pub fn to_public(&self) -> Result<UserPublic, serde_json::Error> {
        Ok(UserPublic {
            email: self.email.clone(),
            username: self.username.clone(),
            roles: self.roles_vec()?,
            permissions: self.permissions_vec()?,
            default_channel_id: self.default_channel_id.clone(),
        })
    }
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct RegisterBody {
    pub email: String,
    pub username: String,
    pub password: String,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct LoginBody {
    pub email: String,
    pub password: String,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct ForgotPasswordBody {
    pub email: String,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct ResetPasswordBody {
    pub token: String,
    pub password: String,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct UpdateUserAccessBody {
    pub is_admin: bool,
    #[serde(default)]
    pub permissions: Vec<String>,
}

#[derive(Debug, Clone, FromRow)]
pub struct PermissionRow {
    pub id: String,
    pub name: String,
    pub description: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct CreatePermissionBody {
    pub name: String,
    pub description: Option<String>,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct UpdatePermissionBody {
    pub description: Option<String>,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct PutRolePermissionsBody {
    pub permission_ids: Vec<String>,
}

#[derive(Debug, Clone, FromRow)]
pub struct RoleRow {
    pub id: String,
    pub name: String,
    pub description: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct CreateRoleBody {
    pub name: String,
    pub description: Option<String>,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct UpdateRoleBody {
    pub name: Option<String>,
    pub description: Option<String>,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct UpdateAdminUserBody {
    pub email: Option<String>,
    pub username: Option<String>,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct CreateAdminUserBody {
    pub email: String,
    pub username: String,
    pub password: String,
    #[serde(default)]
    pub is_verified: Option<bool>,
    #[serde(default)]
    pub is_admin: Option<bool>,
    #[serde(default)]
    pub permissions: Option<Vec<String>>,
    #[serde(default)]
    pub login_id: Option<String>,
    #[serde(default)]
    pub first_name: Option<String>,
    #[serde(default)]
    pub last_name: Option<String>,
    #[serde(default)]
    pub phone: Option<String>,
    #[serde(default)]
    pub title: Option<String>,
    #[serde(default)]
    pub administrator: Option<bool>,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct CreateThreadBody {
    pub title: Option<String>,
    pub member_user_ids: Vec<String>,
    #[serde(default)]
    pub initial_message: Option<String>,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct SendMessageBody {
    pub body: String,
}

#[derive(Debug, Deserialize, ToSchema)]
pub struct MarkReadBody {
    pub last_message_id: Option<String>,
}

#[derive(Debug, Serialize, ToSchema)]
pub struct MessageItem {
    pub id: String,
    pub thread_id: String,
    pub sender_user_id: String,
    pub body: String,
    #[serde(default = "default_message_kind")]
    pub message_kind: String,
    pub created_at: String,
}

fn default_message_kind() -> String {
    "human".to_string()
}

#[derive(Debug, Serialize, ToSchema)]
pub struct ThreadSummary {
    pub thread_id: String,
    pub title: String,
    pub is_group: bool,
    #[serde(default)]
    pub is_public_channel: bool,
    pub member_user_ids: Vec<String>,
    pub latest_message: Option<MessageItem>,
    pub unread_count: i64,
    pub updated_at: String,
}
