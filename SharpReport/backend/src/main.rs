mod api;
mod config;
mod db;
mod metabase;
mod middleware;
mod services;
mod utils;

use crate::metabase::orchestrator::Orchestrator;
use axum::{
    Router,
    routing::{any, delete, get, post, put},
};
use std::net::SocketAddr;
use std::sync::Arc;
use tracing::{Level, info};
use tracing_subscriber::FmtSubscriber;

#[tokio::main]
async fn main() {
    // Initialize tracing
    let subscriber = FmtSubscriber::builder()
        .with_max_level(Level::INFO)
        .finish();
    tracing::subscriber::set_global_default(subscriber).expect("setting default subscriber failed");

    // Load configuration
    let settings = config::Settings::new().expect("Failed to load settings");
    let morph_cfg = morphai::Config::from_env();
    info!("Starting DataPulse backend...");
    info!("UsersPanel auth: {}", settings.users_panel.base_url);
    if morph_cfg.configured() {
        if morph_cfg.uses_openai_compatible() {
            info!(
                "MorphAI: model={} base={}",
                morph_cfg.model,
                morph_cfg.base_url.as_deref().unwrap_or("(default)")
            );
        } else {
            info!("MorphAI: model={} (DashScope native)", morph_cfg.model);
        }
    } else {
        info!("MorphAI: not configured (set MORPH_AI_API_KEY in SharpReport/.env)");
    }
    if settings.metabase.autostart {
        info!("Metabase JAR path: {}", settings.metabase.jar_path);
    } else {
        info!(
            "Metabase autostart off; proxy target http://{}:{}",
            settings.metabase.host, settings.metabase.port
        );
    }

    // Initialize database (resolve relative sqlite paths against cwd or DATAPULSE_WORK_DIR)
    let database_url = config::resolve_database_url(&settings.database.url);
    if database_url.to_lowercase().starts_with("sqlite:") {
        info!(
            "Using SQLite database at {}",
            database_url
                .strip_prefix("sqlite://")
                .unwrap_or(&database_url)
        );
    }
    let pool = db::initialize_db(&database_url)
        .await
        .expect("Failed to initialize database");

    // Initialize Metabase orchestrator
    let server_port = settings.server.port;
    let metabase_handle = Arc::new(
        Orchestrator::new(settings.metabase.clone())
            .await
            .expect("Failed to initialize Metabase orchestrator"),
    );

    let app_state = services::AppState::new(pool, metabase_handle, settings);

    // Build our application with routes
    let app = Router::new()
        // Health check
        .route("/api/v1/health", get(|| async { "OK" }))
        .route("/api/v1/ai/status", get(api::ai_status::status))
        .route("/api/v1/ai/mcp-tools", get(api::mcp_tools::tools))
        // Auth routes
        .route("/api/v1/auth/login", post(api::auth::login))
        .route("/api/v1/auth/logout", post(api::auth::logout))
        .route("/api/v1/auth/me", get(api::auth::me))
        .route("/api/v1/settings", get(api::settings::get))
        // Setup routes
        .route("/api/v1/setup/status", get(api::setup::status))
        .route("/api/v1/setup/initialize", post(api::setup::initialize))
        .route("/api/v1/setup/admin", post(api::setup::create_admin))
        .route("/api/v1/setup/database", post(api::setup::add_database))
        // Database routes
        .route("/api/v1/databases", get(api::databases::list))
        .route("/api/v1/databases", post(api::databases::create))
        .route("/api/v1/databases/:id", get(api::databases::get))
        .route("/api/v1/databases/:id", put(api::databases::update))
        .route("/api/v1/databases/:id", delete(api::databases::delete))
        .route(
            "/api/v1/databases/:id/test",
            post(api::databases::test_connection),
        )
        .route("/api/v1/databases/:id/schema", get(api::schema::get_schema))
        // Data tables (imported flat files)
        .route("/api/v1/data-tables", get(api::data_tables::list))
        .route(
            "/api/v1/data-tables/analyze",
            post(api::data_tables::analyze),
        )
        .route(
            "/api/v1/data-tables/import",
            post(api::data_tables::import_table),
        )
        .route("/api/v1/data-tables/:id", get(api::data_tables::get))
        .route(
            "/api/v1/data-tables/:id",
            delete(api::data_tables::delete_table),
        )
        .route(
            "/api/v1/data-tables/:id/rows",
            get(api::data_tables::get_rows),
        )
        .route(
            "/api/v1/data-tables/:id/query",
            post(api::data_tables::query_table),
        )
        .route(
            "/api/v1/data-tables/:id/rows/:row_index",
            put(api::data_tables::update_row).delete(api::data_tables::delete_row),
        )
        .route(
            "/api/v1/data-tables/:id/page-build",
            post(api::data_page_publish::build_page),
        )
        .route(
            "/api/v1/data-tables/:id/page-builds",
            get(api::data_page_publish::list_page_builds),
        )
        .route(
            "/api/v1/data-tables/:id/page-builds/:build_id",
            get(api::data_page_publish::get_page_build),
        )
        .route(
            "/api/v1/data-tables/:id/page-chat",
            post(api::data_page_publish::page_chat),
        )
        // Published data pages (ComposerX-style URLs)
        .route(
            "/api/v1/publishes/resolve-path",
            get(api::data_page_publish::resolve_path),
        )
        .route(
            "/api/v1/publishes",
            get(api::data_page_publish::list_publishes),
        )
        .route(
            "/api/v1/publishes/catalog",
            get(api::data_page_publish::list_publish_catalog),
        )
        .route(
            "/api/v1/publishes",
            post(api::data_page_publish::create_publish),
        )
        .route(
            "/public/p/:slug",
            get(api::data_page_publish::serve_published_page),
        )
        .route(
            "/public/published/:slug",
            get(api::data_page_publish::get_published_page_json),
        )
        // Docs (Academi bridge + publish)
        .route(
            "/api/v1/docs-bridge/*path",
            any(api::docs_bridge::proxy_with_query),
        )
        .route("/api/v1/docs-publish", post(api::docs_publish::publish))
        .route(
            "/api/v1/docs-publish",
            get(api::docs_publish::list_publishes),
        )
        .route(
            "/public/docs/:slug",
            get(api::docs_publish::serve_published),
        )
        // Report builder
        .route(
            "/api/v1/reports/builder/execute",
            post(api::report_builder::execute),
        )
        .route(
            "/api/v1/reports/assistant/chat",
            post(api::report_assistant::chat),
        )
        .route("/api/v1/assistant/chat", post(api::report_assistant::chat))
        .route(
            "/api/v1/data-ai/chat",
            post(api::report_assistant::data_ai_chat),
        )
        // Dashboard routes
        .route("/api/v1/dashboards", get(api::dashboards::list))
        .route("/api/v1/dashboards", post(api::dashboards::create))
        .route("/api/v1/dashboards/:id", get(api::dashboards::get))
        .route(
            "/api/v1/dashboards/:id/embed",
            get(api::dashboards::get_embed_url),
        )
        // Query routes
        .route("/api/v1/queries", get(api::queries::list))
        .route("/api/v1/queries", post(api::queries::create))
        .route("/api/v1/queries/execute", post(api::queries::execute))
        // Embed routes
        .route("/api/v1/embed/dashboard/:id", post(api::embed::dashboard))
        .route("/api/v1/embed/card/:id", post(api::embed::card))
        // Metabase proxy (Axum 0.7 wildcards use `/*name`, not `{*name}`)
        .route("/metabase/*path", any(api::metabase::proxy))
        // Root: redirect to Vite in debug builds (see `api::frontend::root`); do not use fallback
        .route("/", get(api::frontend::root))
        // Fallback for frontend (will serve static files)
        .fallback(api::frontend::serve)
        // Shared state
        .with_state(app_state.clone())
        // Middleware (outermost layer runs first on request)
        .layer(axum::middleware::from_fn_with_state(
            app_state.clone(),
            middleware::auth::auth_middleware,
        ))
        .layer(axum::middleware::from_fn(
            middleware::logging::logging_middleware,
        ))
        .layer(axum::middleware::from_fn_with_state(
            app_state.clone(),
            middleware::cors::cors_middleware,
        ));

    // Run the application
    let addr = SocketAddr::from(([0, 0, 0, 0], server_port));
    info!("Listening on {}", addr);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap_or_else(|e| {
        if e.kind() == std::io::ErrorKind::AddrInUse {
            panic!(
                "Port {server_port} is already in use. Stop the other DataPulse process (e.g. `lsof -i :{server_port}` then `kill <pid>`) or set APP_SERVER_PORT to another port."
            );
        }
        panic!("Failed to bind to {addr}: {e}");
    });
    axum::serve(listener, app).await.unwrap();
}
