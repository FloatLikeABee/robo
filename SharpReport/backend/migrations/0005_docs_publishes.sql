-- Docs published HTML pages (Board replacement)
CREATE TABLE IF NOT EXISTS docs_publishes (
  id TEXT PRIMARY KEY NOT NULL,
  doc_id TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  html_content TEXT NOT NULL,
  analysis_prompt TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_docs_publishes_slug ON docs_publishes(slug);
