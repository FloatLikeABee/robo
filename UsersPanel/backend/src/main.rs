mod bootstrap;
mod config;
mod error;
mod extractors;
mod handlers;
mod importcol;
mod jwt;
mod message_ai;
mod models;
mod openapi;
mod permissions;
mod permissions_resolve;
mod public_channel;
mod roles_db;
mod state;

use axum::http::Method;
use axum::routing::{delete, get, patch, post};
use axum::Router;
use tower_http::cors::{Any, CorsLayer};
use tower_http::trace::TraceLayer;
use utoipa::OpenApi;
use utoipa_swagger_ui::SwaggerUi;

use crate::config::Config;
use crate::handlers::admin::{
    create_admin_user, deactivate_morph_user, delete_admin_user, list_morph_users, list_users,
    update_admin_user, update_morph_user, update_user_access,
};
use crate::handlers::assistant::chat as assistant_chat;
use crate::handlers::data_collector::{
    get_job, get_template, list_entities, start_import_job, validate_upload,
};
use crate::handlers::auth::{
    forgot_password, get_permissions, get_roles, get_user, google_callback, google_start, login,
    register, reset_password, verify_email,
};
use crate::handlers::integration::{email_compose, forms_create, main_panel, reports};
use crate::handlers::messages::{
    create_thread, delete_thread, list_inbox, list_message_users, list_thread_messages,
    mark_thread_read, send_message, unread_count,
};
use crate::state::AppState;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    dotenvy::dotenv().ok();
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| tracing_subscriber::EnvFilter::new("users_panel_api=info,tower_http=info")),
        )
        .init();

    let config = Config::from_env();
    let pool = sqlx::MySqlPool::connect(&config.database_url).await?;
    sqlx::migrate!("./migrations").run(&pool).await?;

    let state = AppState::new(pool.clone(), config.clone());
    bootstrap::ensure_bootstrap_admin(&state.pool, &state.config).await?;
    public_channel::ensure_platform_setup(&state.pool).await?;
    let message_ai_settings = message_ai::MessageAiSettings::from_app_config(&state.config);
    message_ai::spawn_digest_worker(state.pool.clone(), message_ai_settings);

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods([
            Method::GET,
            Method::POST,
            Method::PATCH,
            Method::PUT,
            Method::DELETE,
            Method::OPTIONS,
        ])
        .allow_headers(Any);

    let app = Router::new()
        .merge(
            SwaggerUi::new("/swagger-ui")
                .url("/openapi.json", crate::openapi::ApiDoc::openapi()),
        )
        .route("/api/auth/register", post(register))
        .route("/api/auth/verify-email", get(verify_email))
        .route("/api/auth/login", post(login))
        .route("/api/auth/google", get(google_start))
        .route("/api/auth/google/callback", get(google_callback))
        .route("/api/auth/forgot-password", post(forgot_password))
        .route("/api/auth/reset-password", post(reset_password))
        .route("/api/auth/roles", get(get_roles))
        .route("/api/auth/permissions", get(get_permissions))
        .route("/api/auth/user", get(get_user))
        .route("/api/assistant/chat", post(assistant_chat))
        .route("/api/data-collector/entities", get(list_entities))
        .route("/api/data-collector/templates/:entity", get(get_template))
        .route("/api/data-collector/validate", post(validate_upload))
        .route("/api/data-collector/jobs", post(start_import_job))
        .route("/api/data-collector/jobs/:job_id", get(get_job))
        .route("/api/admin/users", get(list_users).post(create_admin_user))
        .route(
            "/api/admin/users/:user_id/access",
            patch(update_user_access),
        )
        .route(
            "/api/admin/users/:user_id",
            patch(update_admin_user).delete(delete_admin_user),
        )
        .route("/api/admin/morph-users", get(list_morph_users))
        .route(
            "/api/admin/morph-users/:user_id",
            patch(update_morph_user).delete(deactivate_morph_user),
        )
        .route("/api/main-panel", get(main_panel))
        .route("/api/forms", post(forms_create))
        .route("/api/email/compose", post(email_compose))
        .route("/api/reports", get(reports))
        .route("/api/messages/inbox", get(list_inbox))
        .route("/api/messages/users", get(list_message_users))
        .route("/api/messages/unread-count", get(unread_count))
        .route("/api/messages/threads", post(create_thread))
        .route(
            "/api/messages/threads/:thread_id/messages",
            get(list_thread_messages).post(send_message),
        )
        .route("/api/messages/threads/:thread_id/delete", post(delete_thread))
        .route("/api/messages/threads/:thread_id/read", post(mark_thread_read))
        .route("/api/messages/threads/:thread_id", delete(delete_thread))
        .layer(cors)
        .layer(TraceLayer::new_for_http())
        .with_state(state);

    let addr = format!("{}:{}", config.host, config.port);
    let listener = tokio::net::TcpListener::bind(&addr).await?;
    tracing::info!("listening on http://{}", addr);
    axum::serve(listener, app).await?;

    Ok(())
}
