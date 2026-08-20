pub mod assistant;
pub mod auth;
pub mod extract;
pub mod flow_log;
pub mod modules;
pub mod project_docs;
pub mod project_import;
pub mod record_validation;
pub mod verification;

use axum::{
    middleware,
    routing::{delete, get, patch, post},
    Router,
};
use std::sync::Arc;
use tower_http::cors::{Any, CorsLayer};

use crate::middleware::auth::auth_layer;
use crate::services::AppState;

pub fn router(state: Arc<AppState>) -> Router {
    let cors = CorsLayer::new()
        .allow_origin(tower_http::cors::AllowOrigin::mirror_request())
        .allow_methods(Any)
        .allow_headers(Any);

    if !state.settings.is_development() {
        let _ = state.settings.cors_origin.clone();
    }

    let api = Router::new()
        .route("/auth/login", post(auth::login))
        .route("/auth/platform-session", post(auth::platform_session))
        .route("/auth/dev-login", post(auth::dev_login))
        .route("/auth/me", get(auth::me))
        .route("/organization", get(auth::get_organization))
        .route("/organization", patch(auth::patch_organization))
        .route("/assistant/chat", post(assistant::assistant_chat))
        .route("/projects", get(modules::list_projects))
        .route("/projects", post(modules::create_project))
        .route("/projects/generate-document", post(project_docs::generate_document))
        .route("/projects/:id", patch(modules::update_project))
        .route("/projects/:id", delete(modules::delete_project))
        .route("/projects/:id/publish", post(project_docs::publish_project))
        .route("/tasks", get(modules::list_tasks))
        .route("/tasks", post(modules::create_task))
        .route("/site-logs", get(modules::list_site_logs))
        .route("/site-logs", post(modules::create_site_log))
        .route("/materials", get(modules::list_materials))
        .route("/materials", post(modules::create_material))
        .route("/materials/:id", patch(modules::update_material))
        .route("/material-usages", get(modules::list_material_usages))
        .route("/material-usages", post(modules::create_material_usage))
        .route("/resource-files", get(modules::list_resource_files))
        .route("/resource-files", post(modules::create_resource_file))
        .route("/resource-files/:id", patch(modules::update_resource_file))
        .route("/resource-files/:id", delete(modules::delete_resource_file))
        .route("/resource-files/upload", post(modules::upload_resource_file))
        .route("/finance", get(modules::get_project_finance))
        .route("/finance", post(modules::upsert_project_finance))
        .route("/budget-lines", get(modules::list_budget_lines))
        .route("/budget-lines", post(modules::create_budget_line))
        .route("/budget-lines/:id", patch(modules::update_budget_line))
        .route("/resources", get(modules::list_resources))
        .route("/resources", post(modules::create_resource))
        .route("/resource-allocations", get(modules::list_resource_allocations))
        .route("/resource-allocations", post(modules::create_resource_allocation))
        .route("/contractors", get(modules::list_contractors))
        .route("/contractors", post(modules::create_contractor))
        .route("/contracts", get(modules::list_contracts))
        .route("/contracts", post(modules::create_contract))
        .route("/relations", get(modules::list_relations))
        .route("/relations", post(modules::create_relation))
        .route("/communications", get(modules::list_communications))
        .route("/communications", post(modules::create_communication))
        .route("/project-imports", get(project_import::list_sessions))
        .route("/project-imports/analyze", post(project_import::analyze))
        .route("/project-imports/:id", get(project_import::get_session))
        .route("/project-imports/:id", delete(project_import::delete_session))
        .route("/project-imports/:id/confirm", post(project_import::confirm))
        .route("/verification/sessions", get(verification::list_verification_sessions))
        .route("/verification/sessions/:id", get(verification::get_verification_session))
        .route("/verification/run", post(verification::run_verification))
        .route("/flow-log/entries", get(flow_log::list_entries))
        .route("/flow-log/entries", post(flow_log::create_entry))
        .route("/flow-log/entries/:id", delete(flow_log::delete_entry))
        .route("/flow-log/summary", get(flow_log::summary))
        .layer(middleware::from_fn_with_state(state.clone(), auth_layer));

    Router::new()
        .route("/health", get(|| async { axum::Json(serde_json::json!({"status":"ok","service":"morph-engi-api"})) }))
        .route("/api/v1/health", get(|| async { axum::Json(serde_json::json!({"status":"ok"})) }))
        .route("/api/v1/uploads/:org_id/:filename", get(modules::serve_upload))
        .route("/api/v1/public/projects/:slug", get(project_docs::serve_public_project))
        .nest("/api/v1", api)
        .layer(cors)
        .with_state(state)
}
