//! OpenAPI 3.0 spec for the Users Panel HTTP API.
//!
//! Interactive UI: `GET /swagger-ui` (served from the API process, same port as `HOST:PORT`).

use utoipa::openapi::security::{HttpAuthScheme, HttpBuilder, SecurityScheme};
use utoipa::{Modify, OpenApi, ToSchema};

use crate::models::{
    CreatePermissionBody, CreateRoleBody, ForgotPasswordBody, LoginBody, PutRolePermissionsBody,
    RegisterBody, ResetPasswordBody, UpdateAdminUserBody, UpdatePermissionBody, UpdateRoleBody,
    UpdateUserAccessBody, UserPublic, CreateThreadBody, SendMessageBody, MarkReadBody, MessageItem,
    ThreadSummary,
};

/// Standard JSON error payload from failed requests.
#[derive(Debug, ToSchema, serde::Serialize)]
pub struct ErrorBody {
    pub error: String,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct MessageBody {
    pub message: String,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct LoginResponse {
    pub token: String,
    #[schema(inline)]
    pub user: UserPublic,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct RolesInTokenBody {
    pub roles: Vec<String>,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct PermissionsForUserBody {
    pub permissions: Vec<String>,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct CurrentUserBody {
    pub user: UserPublic,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct InboxBody {
    pub threads: Vec<ThreadSummary>,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct ThreadMessagesBody {
    pub messages: Vec<MessageItem>,
    pub limit: u32,
    pub offset: u32,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct UnreadCountBody {
    pub unread_count: i64,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct CreateThreadResponseBody {
    pub thread_id: String,
    pub message_id: Option<String>,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct MessageUserItem {
    pub id: String,
    pub username: String,
    pub email: String,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct MessageUsersBody {
    pub users: Vec<MessageUserItem>,
}

/// Admin user row as returned by `GET /api/admin/users` (camelCase for some keys).
#[derive(Debug, ToSchema, serde::Serialize)]
#[serde(rename_all = "camelCase")]
pub struct AdminUserListItem {
    pub id: String,
    pub email: String,
    pub username: String,
    pub is_verified: bool,
    pub roles: Vec<String>,
    pub default_channel_id: String,
    pub created_at: String,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct AdminUsersListBody {
    pub users: Vec<AdminUserListItem>,
}

#[derive(Debug, ToSchema, serde::Serialize)]
#[serde(rename_all = "camelCase")]
pub struct AdminRoleItem {
    pub id: String,
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct AdminRolesListBody {
    pub roles: Vec<AdminRoleItem>,
}

#[derive(Debug, ToSchema, serde::Serialize)]
#[serde(rename_all = "camelCase")]
pub struct AdminCreateRoleResponse {
    pub id: String,
    pub name: String,
    pub description: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, ToSchema, serde::Serialize)]
#[serde(rename_all = "camelCase")]
pub struct AdminPermissionItem {
    pub id: String,
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, ToSchema, serde::Serialize)]
pub struct AdminPermissionsListBody {
    pub permissions: Vec<AdminPermissionItem>,
}

#[derive(Debug, ToSchema, serde::Serialize)]
#[serde(rename_all = "camelCase")]
pub struct PermissionsOverviewBody {
    pub permissions: Vec<AdminPermissionItem>,
    /// Distinct role names (display names, as in `plat_roles`).
    pub role_names: Vec<String>,
    /// Role display name → list of `plat_permissions.id` values.
    pub assignments: std::collections::HashMap<String, Vec<String>>,
}

#[derive(Debug, ToSchema, serde::Serialize)]
#[serde(rename_all = "camelCase")]
pub struct AdminCreatePermissionResponse {
    pub id: String,
    pub name: String,
    pub description: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

pub struct SecurityAddon;

impl Modify for SecurityAddon {
    fn modify(&self, openapi: &mut utoipa::openapi::OpenApi) {
        if let Some(components) = openapi.components.as_mut() {
            components.add_security_scheme(
                "bearer",
                SecurityScheme::Http(
                    HttpBuilder::new()
                        .scheme(HttpAuthScheme::Bearer)
                        .bearer_format("JWT")
                        .build(),
                ),
            );
        }
    }
}

#[allow(dead_code, unused_imports)]
mod doc {
    // Types referenced in `path` attributes for the OpenAPI doc bundle (macro uses outer scope).
    use super::*;

    #[utoipa::path(
        post,
        path = "/api/auth/register",
        tag = "auth",
        request_body = RegisterBody,
        responses(
            (status = 200, description = "User created; verify email before login. Verification URL is logged in dev.", body = MessageBody),
            (status = 400, body = ErrorBody),
            (status = 409, body = ErrorBody)
        )
    )]
    pub fn register() {}

    #[utoipa::path(
        get,
        path = "/api/auth/verify-email",
        tag = "auth",
        params(
            ("token" = String, Query, description = "Verification token from the registration / email link")
        ),
        responses(
            (status = 200, description = "Email verified", body = MessageBody),
            (status = 400, body = ErrorBody)
        )
    )]
    pub fn verify_email() {}

    #[utoipa::path(
        post,
        path = "/api/auth/login",
        tag = "auth",
        request_body = LoginBody,
        responses(
            (status = 200, body = LoginResponse),
            (status = 400, body = ErrorBody),
            (status = 401, body = ErrorBody, description = "Invalid credentials or unknown email")
        )
    )]
    pub fn login() {}

    #[utoipa::path(
        get,
        path = "/api/auth/google",
        tag = "auth",
        responses(
            (status = 302, description = "Redirect to Google OAuth (requires GOOGLE_CLIENT_ID/SECRET)"),
            (status = 400, body = ErrorBody, description = "Google OAuth not configured")
        )
    )]
    pub fn google_start() {}

    #[utoipa::path(
        get,
        path = "/api/auth/google/callback",
        tag = "auth",
        params(
            ("code" = Option<String>, Query, description = "Authorization code from Google"),
            ("state" = Option<String>, Query, description = "CSRF / PKCE state"),
            ("error" = Option<String>, Query, description = "OAuth error from provider")
        ),
        responses(
            (status = 302, description = "Redirect to FRONTEND_ORIGIN with JWT in hash or error query")
        )
    )]
    pub fn google_callback() {}

    #[utoipa::path(
        post,
        path = "/api/auth/forgot-password",
        tag = "auth",
        request_body = ForgotPasswordBody,
        responses(
            (status = 200, description = "Always same message; reset URL logged in dev", body = MessageBody)
        )
    )]
    pub fn forgot_password() {}

    #[utoipa::path(
        post,
        path = "/api/auth/reset-password",
        tag = "auth",
        request_body = ResetPasswordBody,
        responses(
            (status = 200, body = MessageBody),
            (status = 400, body = ErrorBody)
        )
    )]
    pub fn reset_password() {}

    #[utoipa::path(
        get,
        path = "/api/auth/roles",
        tag = "auth",
        security(("bearer" = [])),
        responses(
            (status = 200, body = RolesInTokenBody, description = "Role names from JWT for client convenience"),
            (status = 401, body = ErrorBody)
        )
    )]
    pub fn get_roles() {}

    #[utoipa::path(
        get,
        path = "/api/auth/permissions",
        tag = "auth",
        security(("bearer" = [])),
        responses(
            (status = 200, body = PermissionsForUserBody, description = "Resolved permission slugs for the current user's roles"),
            (status = 401, body = ErrorBody)
        )
    )]
    pub fn get_permissions() {}

    #[utoipa::path(
        get,
        path = "/api/auth/user",
        tag = "auth",
        security(("bearer" = [])),
        responses(
            (status = 200, body = CurrentUserBody),
            (status = 401, body = ErrorBody),
            (status = 404, body = ErrorBody)
        )
    )]
    pub fn get_user() {}

    #[utoipa::path(
        get,
        path = "/api/admin/users",
        tag = "admin-users",
        security(("bearer" = [])),
        responses(
            (status = 200, body = AdminUsersListBody, description = "JSON uses isVerified, defaultChannelId, createdAt keys from server"),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody, description = "Not Admin")
        )
    )]
    pub fn list_users() {}

    #[utoipa::path(
        patch,
        path = "/api/admin/users/{user_id}/access",
        tag = "admin-users",
        security(("bearer" = [])),
        params(
            ("user_id" = String, Path, description = "User UUID")
        ),
        request_body = UpdateUserAccessBody,
        responses(
            (status = 200, body = MessageBody),
            (status = 400, body = ErrorBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody),
            (status = 404, body = ErrorBody)
        )
    )]
    pub fn update_user_access() {}

    #[utoipa::path(
        patch,
        path = "/api/admin/users/{user_id}",
        tag = "admin-users",
        security(("bearer" = [])),
        params(
            ("user_id" = String, Path, description = "User UUID")
        ),
        request_body = UpdateAdminUserBody,
        responses(
            (status = 200, body = MessageBody),
            (status = 400, body = ErrorBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody),
            (status = 404, body = ErrorBody),
            (status = 409, body = ErrorBody)
        )
    )]
    pub fn update_admin_user() {}

    #[utoipa::path(
        delete,
        path = "/api/admin/users/{user_id}",
        tag = "admin-users",
        security(("bearer" = [])),
        params(
            ("user_id" = String, Path, description = "User UUID")
        ),
        responses(
            (status = 200, body = MessageBody),
            (status = 400, body = ErrorBody, description = "Cannot delete self"),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody),
            (status = 404, body = ErrorBody)
        )
    )]
    pub fn delete_admin_user() {}

    #[utoipa::path(
        get,
        path = "/api/admin/roles",
        tag = "admin-roles",
        security(("bearer" = [])),
        responses(
            (status = 200, body = AdminRolesListBody, description = "JSON uses createdAt, updatedAt on each role from server"),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn list_roles() {}

    #[utoipa::path(
        post,
        path = "/api/admin/roles",
        tag = "admin-roles",
        security(("bearer" = [])),
        request_body = CreateRoleBody,
        responses(
            (status = 200, body = AdminCreateRoleResponse, description = "JSON uses createdAt, updatedAt from server"),
            (status = 400, body = ErrorBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody),
            (status = 409, body = ErrorBody)
        )
    )]
    pub fn create_role() {}

    #[utoipa::path(
        patch,
        path = "/api/admin/roles/{id}",
        tag = "admin-roles",
        security(("bearer" = [])),
        params(
            ("id" = String, Path, description = "Role row UUID")
        ),
        request_body = UpdateRoleBody,
        responses(
            (status = 200, body = MessageBody),
            (status = 400, body = ErrorBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody),
            (status = 404, body = ErrorBody),
            (status = 409, body = ErrorBody)
        )
    )]
    pub fn update_role() {}

    #[utoipa::path(
        delete,
        path = "/api/admin/roles/{id}",
        tag = "admin-roles",
        security(("bearer" = [])),
        params(
            ("id" = String, Path, description = "Role row UUID")
        ),
        responses(
            (status = 200, body = MessageBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody),
            (status = 404, body = ErrorBody)
        )
    )]
    pub fn delete_role() {}

    #[utoipa::path(
        get,
        path = "/api/admin/permissions",
        tag = "admin-permissions",
        security(("bearer" = [])),
        responses(
            (status = 200, body = AdminPermissionsListBody, description = "JSON uses createdAt, updatedAt from server"),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn list_permissions() {}

    #[utoipa::path(
        get,
        path = "/api/admin/permissions/overview",
        tag = "admin-permissions",
        security(("bearer" = [])),
        responses(
            (status = 200, body = PermissionsOverviewBody, description = "Matrix of roles to permission ids; JSON uses roleNames, assignments from server"),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn permissions_overview() {}

    #[utoipa::path(
        post,
        path = "/api/admin/permissions",
        tag = "admin-permissions",
        security(("bearer" = [])),
        request_body = CreatePermissionBody,
        responses(
            (status = 200, body = AdminCreatePermissionResponse, description = "JSON uses createdAt, updatedAt from server"),
            (status = 400, body = ErrorBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody),
            (status = 409, body = ErrorBody)
        )
    )]
    pub fn create_permission() {}

    #[utoipa::path(
        patch,
        path = "/api/admin/permissions/{id}",
        tag = "admin-permissions",
        security(("bearer" = [])),
        params(
            ("id" = String, Path, description = "Permission row UUID")
        ),
        request_body = UpdatePermissionBody,
        responses(
            (status = 200, body = MessageBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody),
            (status = 404, body = ErrorBody)
        )
    )]
    pub fn update_permission() {}

    #[utoipa::path(
        delete,
        path = "/api/admin/permissions/{id}",
        tag = "admin-permissions",
        security(("bearer" = [])),
        params(
            ("id" = String, Path, description = "Permission row UUID")
        ),
        responses(
            (status = 200, body = MessageBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody),
            (status = 404, body = ErrorBody)
        )
    )]
    pub fn delete_permission() {}

    #[utoipa::path(
        put,
        path = "/api/admin/roles/{role_name}/permissions",
        tag = "admin-permissions",
        security(("bearer" = [])),
        params(
            ("role_name" = String, Path, description = "Role display name (URL-encoded), e.g. Main%20Panel")
        ),
        request_body = PutRolePermissionsBody,
        responses(
            (status = 200, body = MessageBody, description = "Replaces the full permission set for that role"),
            (status = 400, body = ErrorBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn put_role_permissions() {}

    #[utoipa::path(
        get,
        path = "/api/main-panel",
        tag = "integration",
        security(("bearer" = [])),
        responses(
            (status = 200, description = "Stub when Main Panel role is present"),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody, description = "Missing Main Panel role")
        )
    )]
    pub fn main_panel() {}

    #[utoipa::path(
        post,
        path = "/api/forms",
        tag = "integration",
        security(("bearer" = [])),
        responses(
            (status = 200, description = "Stub when Forms role is present"),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn forms_create() {}

    #[utoipa::path(
        post,
        path = "/api/email/compose",
        tag = "integration",
        security(("bearer" = [])),
        responses(
            (status = 200, description = "Stub when Email Composer role is present"),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn email_compose() {}

    #[utoipa::path(
        get,
        path = "/api/reports",
        tag = "integration",
        security(("bearer" = [])),
        responses(
            (status = 200, description = "Stub when Sharp Reports role is present"),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn reports() {}

    #[utoipa::path(
        get,
        path = "/api/messages/inbox",
        tag = "messages",
        security(("bearer" = [])),
        responses(
            (status = 200, body = InboxBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn list_inbox() {}

    #[utoipa::path(
        get,
        path = "/api/messages/users",
        tag = "messages",
        security(("bearer" = [])),
        responses(
            (status = 200, body = MessageUsersBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn list_message_users() {}

    #[utoipa::path(
        get,
        path = "/api/messages/unread-count",
        tag = "messages",
        security(("bearer" = [])),
        responses(
            (status = 200, body = UnreadCountBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn unread_count() {}

    #[utoipa::path(
        post,
        path = "/api/messages/threads",
        tag = "messages",
        security(("bearer" = [])),
        request_body = CreateThreadBody,
        responses(
            (status = 200, body = CreateThreadResponseBody),
            (status = 400, body = ErrorBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn create_thread() {}

    #[utoipa::path(
        delete,
        path = "/api/messages/threads/{thread_id}",
        tag = "messages",
        security(("bearer" = [])),
        params(
            ("thread_id" = String, Path, description = "Thread UUID")
        ),
        responses(
            (status = 200, body = MessageBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody),
            (status = 404, body = ErrorBody)
        )
    )]
    pub fn delete_thread() {}

    #[utoipa::path(
        post,
        path = "/api/messages/threads/{thread_id}/delete",
        tag = "messages",
        security(("bearer" = [])),
        params(
            ("thread_id" = String, Path, description = "Thread UUID")
        ),
        responses(
            (status = 200, body = MessageBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody),
            (status = 404, body = ErrorBody)
        )
    )]
    pub fn delete_thread_post() {}

    #[utoipa::path(
        get,
        path = "/api/messages/threads/{thread_id}/messages",
        tag = "messages",
        security(("bearer" = [])),
        params(
            ("thread_id" = String, Path, description = "Thread UUID"),
            ("limit" = Option<u32>, Query, description = "Page size, default 50"),
            ("offset" = Option<u32>, Query, description = "Offset, default 0")
        ),
        responses(
            (status = 200, body = ThreadMessagesBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn list_thread_messages() {}

    #[utoipa::path(
        post,
        path = "/api/messages/threads/{thread_id}/messages",
        tag = "messages",
        security(("bearer" = [])),
        params(
            ("thread_id" = String, Path, description = "Thread UUID")
        ),
        request_body = SendMessageBody,
        responses(
            (status = 200, description = "Message sent"),
            (status = 400, body = ErrorBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn send_message() {}

    #[utoipa::path(
        post,
        path = "/api/messages/threads/{thread_id}/read",
        tag = "messages",
        security(("bearer" = [])),
        params(
            ("thread_id" = String, Path, description = "Thread UUID")
        ),
        request_body = MarkReadBody,
        responses(
            (status = 200, body = MessageBody),
            (status = 400, body = ErrorBody),
            (status = 401, body = ErrorBody),
            (status = 403, body = ErrorBody)
        )
    )]
    pub fn mark_thread_read() {}
}

#[derive(OpenApi)]
#[openapi(
    info(
        title = "Users Panel API",
        version = "0.1.0",
        description = "User authentication, JWT sessions, role-based access, and admin APIs. \
            Use in microservices: issue JWTs from this service, validate them in each service, \
            or call this service to resolve `GET /api/auth/permissions` for a user. \
            All JSON error responses use `{\"error\": string}`.",
    ),
    paths(
        doc::register,
        doc::verify_email,
        doc::login,
        doc::google_start,
        doc::google_callback,
        doc::forgot_password,
        doc::reset_password,
        doc::get_roles,
        doc::get_permissions,
        doc::get_user,
        doc::list_users,
        doc::update_user_access,
        doc::update_admin_user,
        doc::delete_admin_user,
        doc::list_roles,
        doc::create_role,
        doc::update_role,
        doc::delete_role,
        doc::list_permissions,
        doc::permissions_overview,
        doc::create_permission,
        doc::update_permission,
        doc::delete_permission,
        doc::put_role_permissions,
        doc::main_panel,
        doc::forms_create,
        doc::email_compose,
        doc::reports,
        doc::list_inbox,
        doc::list_message_users,
        doc::unread_count,
        doc::create_thread,
        doc::delete_thread,
        doc::delete_thread_post,
        doc::list_thread_messages,
        doc::send_message,
        doc::mark_thread_read,
    ),
    components(
        schemas(
            ErrorBody,
            MessageBody,
            LoginResponse,
            UserPublic,
            RegisterBody,
            LoginBody,
            ForgotPasswordBody,
            ResetPasswordBody,
            RolesInTokenBody,
            PermissionsForUserBody,
            CurrentUserBody,
            AdminUsersListBody,
            AdminUserListItem,
            AdminRolesListBody,
            AdminCreateRoleResponse,
            AdminRoleItem,
            AdminPermissionItem,
            AdminPermissionsListBody,
            PermissionsOverviewBody,
            AdminCreatePermissionResponse,
            UpdateUserAccessBody,
            UpdateAdminUserBody,
            CreateRoleBody,
            UpdateRoleBody,
            CreatePermissionBody,
            UpdatePermissionBody,
            PutRolePermissionsBody,
            CreateThreadBody,
            SendMessageBody,
            MarkReadBody,
            MessageItem,
            ThreadSummary,
            InboxBody,
            ThreadMessagesBody,
            UnreadCountBody,
            CreateThreadResponseBody,
            MessageUserItem,
            MessageUsersBody,
        )
    ),
    tags(
        (name = "auth", description = "Registration, email verification, login, OAuth, password reset, and JWT profile"),
        (name = "admin-users", description = "List/update/delete users; assign role names to users (Admin role)"),
        (name = "admin-roles", description = "CRUD for role definitions in `plat_roles` (Admin role)"),
        (name = "admin-permissions", description = "CRUD permissions and assign permissions to roles (Admin role)"),
        (name = "integration", description = "Example feature endpoints gated by feature roles; extend as needed"),
        (name = "messages", description = "In-system direct/group messaging")
    ),
    modifiers(&SecurityAddon)
)]
pub struct ApiDoc;
