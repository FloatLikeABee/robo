use axum::Json;
use serde_json::json;

pub async fn tools() -> Json<serde_json::Value> {
    Json(json!({
        "service": "datax-ai-tools",
        "description": "Read-only MCP-style catalog for DataX AI database, schema, query, and imported table tools.",
        "safety": {
            "sql": "execute_query is read-only and accepts SELECT/WITH only",
            "mutations": "Use normal DataX HTTP APIs for explicit user-requested mutations; AI data analysis tools are read-first"
        },
        "tools": [
            {
                "name": "list_databases",
                "description": "List connected database ids, names, and engines.",
                "assistant_endpoint": "/api/v1/data-ai/chat"
            },
            {
                "name": "database_schema",
                "description": "Fetch tables and columns for one connected database.",
                "args": {"database_id": "uuid string"},
                "http_equivalent": "GET /api/v1/databases/{id}/schema"
            },
            {
                "name": "execute_query",
                "description": "Run a read-only SQL SELECT/WITH query against a connected database.",
                "args": {"database_id": "uuid string", "sql": "SELECT ..."},
                "http_equivalent": "POST /api/v1/queries/execute"
            },
            {
                "name": "list_data_tables",
                "description": "List imported DataX data tables with ids, columns, and row counts.",
                "http_equivalent": "GET /api/v1/data-tables"
            },
            {
                "name": "data_table_schema",
                "description": "Fetch imported table columns, row count, and sample rows.",
                "args": {"table_id": "uuid string"},
                "http_equivalent": "GET /api/v1/data-tables/{id}"
            },
            {
                "name": "query_data_table",
                "description": "Search, sort, page, and aggregate imported data table rows.",
                "args": {
                    "table_id": "uuid string",
                    "search": "optional text",
                    "sort_by": "optional column",
                    "sort_dir": "asc or desc",
                    "group_by": "optional column",
                    "aggregate_op": "count|sum|avg|min|max",
                    "aggregate_column": "optional numeric column"
                },
                "http_equivalent": "POST /api/v1/data-tables/{id}/query"
            }
        ]
    }))
}
