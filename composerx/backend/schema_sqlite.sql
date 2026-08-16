-- ComposerX embedded SQLite schema (adapted from schema.sql).
-- No ENGINE/ENUM; INTEGER PRIMARY KEY AUTOINCREMENT; content_mongo_id kept as TEXT for UUID Badger keys.

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS email_templates (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  tag TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  builtin_key TEXT NULL,
  html_content TEXT NOT NULL,
  created_by INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (builtin_key),
  FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_email_templates_created_by ON email_templates(created_by);

CREATE TABLE IF NOT EXISTS merge_data_sources (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  file_type TEXT NOT NULL CHECK (file_type IN ('csv', 'json', 'excel')),
  file_path TEXT NOT NULL,
  uploaded_by INTEGER NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (uploaded_by) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_merge_data_sources_name ON merge_data_sources(name);
CREATE INDEX IF NOT EXISTS idx_merge_data_sources_uploaded_by ON merge_data_sources(uploaded_by);

CREATE TABLE IF NOT EXISTS saved_emails (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  content_mongo_id TEXT NOT NULL,
  created_by INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (content_mongo_id),
  FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_saved_emails_updated_at ON saved_emails(updated_at);

CREATE TABLE IF NOT EXISTS published_pages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  slug TEXT NOT NULL,
  theme TEXT NOT NULL DEFAULT 'default',
  content_mongo_id TEXT NOT NULL,
  created_by INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (slug),
  UNIQUE (content_mongo_id),
  FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_published_pages_updated_at ON published_pages(updated_at);

CREATE TABLE IF NOT EXISTS publish_drafts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  theme TEXT NOT NULL DEFAULT 'default',
  html_content TEXT NOT NULL,
  created_by INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_publish_drafts_updated_at ON publish_drafts(updated_at);

CREATE TABLE IF NOT EXISTS report_configs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  report_name TEXT NOT NULL,
  report_service_url TEXT NOT NULL,
  report_type TEXT NOT NULL,
  max_files_allowed INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS email_contacts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL DEFAULT '',
  email TEXT NOT NULL,
  phone TEXT NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (email)
);

CREATE INDEX IF NOT EXISTS idx_email_contacts_name ON email_contacts(name);

-- Local Neo4j ingest outbox (ComposerX enqueueGraphSync); worker may live elsewhere.
CREATE TABLE IF NOT EXISTS graph_sync_outbox (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  source TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  op TEXT NOT NULL,
  payload_json TEXT NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  available_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  attempts INTEGER NOT NULL DEFAULT 0,
  locked_by TEXT NULL,
  locked_at TEXT NULL,
  processed_at TEXT NULL,
  last_error TEXT NULL
);
