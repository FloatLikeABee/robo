use sqlx::SqlitePool;

const SCHEMA: &str = r#"
CREATE TABLE IF NOT EXISTS organizations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  country TEXT DEFAULT '',
  currency TEXT DEFAULT 'USD',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL REFERENCES organizations(id),
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  role TEXT NOT NULL DEFAULT 'owner',
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS projects (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL REFERENCES organizations(id),
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  client TEXT DEFAULT '',
  location TEXT DEFAULT '',
  status TEXT NOT NULL DEFAULT 'planning',
  start_date TEXT,
  end_date TEXT,
  budget_total REAL NOT NULL DEFAULT 0,
  progress_pct REAL NOT NULL DEFAULT 0,
  description TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS project_tasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER NOT NULL REFERENCES projects(id),
  title TEXT NOT NULL,
  assignee TEXT DEFAULT '',
  status TEXT NOT NULL DEFAULT 'open',
  priority TEXT NOT NULL DEFAULT 'normal',
  due_date TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS site_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER NOT NULL REFERENCES projects(id),
  log_date TEXT NOT NULL,
  weather TEXT DEFAULT '',
  crew_count INTEGER DEFAULT 0,
  summary TEXT NOT NULL,
  issues TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS materials (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  category TEXT DEFAULT 'general',
  unit TEXT NOT NULL DEFAULT 'ea',
  unit_cost REAL NOT NULL DEFAULT 0,
  supplier TEXT DEFAULT '',
  stock_qty REAL NOT NULL DEFAULT 0,
  reorder_level REAL NOT NULL DEFAULT 0,
  description TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS project_materials (
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  material_id INTEGER NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
  PRIMARY KEY (project_id, material_id)
);

CREATE TABLE IF NOT EXISTS material_usages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER NOT NULL REFERENCES projects(id),
  material_id INTEGER NOT NULL REFERENCES materials(id),
  qty REAL NOT NULL,
  used_at TEXT NOT NULL,
  notes TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS resource_files (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  source_type TEXT NOT NULL DEFAULT 'url',
  file_url TEXT DEFAULT '',
  file_name TEXT DEFAULT '',
  description TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS project_resource_files (
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  resource_file_id INTEGER NOT NULL REFERENCES resource_files(id) ON DELETE CASCADE,
  PRIMARY KEY (project_id, resource_file_id)
);

CREATE TABLE IF NOT EXISTS project_finances (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
  summary TEXT DEFAULT '',
  total_planned REAL NOT NULL DEFAULT 0,
  total_actual REAL NOT NULL DEFAULT 0,
  notes TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS budget_lines (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER NOT NULL REFERENCES projects(id),
  cost_code TEXT NOT NULL,
  description TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT 'material',
  planned_amount REAL NOT NULL DEFAULT 0,
  actual_amount REAL NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS resources (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  resource_type TEXT NOT NULL DEFAULT 'equipment',
  cost_per_day REAL NOT NULL DEFAULT 0,
  availability TEXT NOT NULL DEFAULT 'available',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS resource_allocations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER NOT NULL REFERENCES projects(id),
  resource_id INTEGER NOT NULL REFERENCES resources(id),
  start_date TEXT NOT NULL,
  end_date TEXT,
  qty REAL NOT NULL DEFAULT 1,
  notes TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS contractors (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  trade TEXT DEFAULT '',
  contact_name TEXT DEFAULT '',
  phone TEXT DEFAULT '',
  email TEXT DEFAULT '',
  rating REAL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'active',
  description TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS project_contractors (
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  contractor_id INTEGER NOT NULL REFERENCES contractors(id) ON DELETE CASCADE,
  PRIMARY KEY (project_id, contractor_id)
);

CREATE TABLE IF NOT EXISTS contractor_contracts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER NOT NULL REFERENCES projects(id),
  contractor_id INTEGER NOT NULL REFERENCES contractors(id),
  scope TEXT NOT NULL,
  contract_value REAL NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'draft',
  start_date TEXT,
  end_date TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS public_relations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  role TEXT DEFAULT '',
  contact TEXT DEFAULT '',
  influence TEXT NOT NULL DEFAULT 'medium',
  sentiment TEXT NOT NULL DEFAULT 'neutral',
  description TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS communications (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER REFERENCES projects(id),
  relation_id INTEGER REFERENCES public_relations(id),
  channel TEXT NOT NULL DEFAULT 'meeting',
  subject TEXT NOT NULL,
  body TEXT NOT NULL DEFAULT '',
  occurred_at TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_projects_org ON projects(organization_id);
CREATE INDEX IF NOT EXISTS idx_materials_org ON materials(organization_id);
CREATE INDEX IF NOT EXISTS idx_resource_files_org ON resource_files(organization_id);
CREATE INDEX IF NOT EXISTS idx_public_relations_project ON public_relations(project_id);

CREATE TABLE IF NOT EXISTS verification_sessions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER REFERENCES projects(id),
  title TEXT NOT NULL DEFAULT '',
  verify_prompt TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'draft',
  verdict TEXT NOT NULL DEFAULT '',
  summary TEXT NOT NULL DEFAULT '',
  result_html TEXT NOT NULL DEFAULT '',
  result_markdown TEXT NOT NULL DEFAULT '',
  findings_json TEXT NOT NULL DEFAULT '[]',
  file_bundle_json TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS verification_files (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id INTEGER NOT NULL REFERENCES verification_sessions(id) ON DELETE CASCADE,
  organization_id INTEGER NOT NULL,
  original_name TEXT NOT NULL,
  stored_name TEXT NOT NULL,
  mime_type TEXT DEFAULT '',
  size_bytes INTEGER NOT NULL DEFAULT 0,
  excerpt TEXT NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_verification_sessions_org ON verification_sessions(organization_id);
"#;

const MIGRATIONS: &[&str] = &[
    "ALTER TABLE materials ADD COLUMN description TEXT DEFAULT ''",
    "ALTER TABLE contractors ADD COLUMN description TEXT DEFAULT ''",
    r#"
CREATE TABLE IF NOT EXISTS project_materials (
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  material_id INTEGER NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
  PRIMARY KEY (project_id, material_id)
)"#,
    r#"
CREATE TABLE IF NOT EXISTS resource_files (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  source_type TEXT NOT NULL DEFAULT 'url',
  file_url TEXT DEFAULT '',
  file_name TEXT DEFAULT '',
  description TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
)"#,
    r#"
CREATE TABLE IF NOT EXISTS project_resource_files (
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  resource_file_id INTEGER NOT NULL REFERENCES resource_files(id) ON DELETE CASCADE,
  PRIMARY KEY (project_id, resource_file_id)
)"#,
    r#"
CREATE TABLE IF NOT EXISTS project_finances (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
  summary TEXT DEFAULT '',
  total_planned REAL NOT NULL DEFAULT 0,
  total_actual REAL NOT NULL DEFAULT 0,
  notes TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
)"#,
    r#"
CREATE TABLE IF NOT EXISTS project_contractors (
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  contractor_id INTEGER NOT NULL REFERENCES contractors(id) ON DELETE CASCADE,
  PRIMARY KEY (project_id, contractor_id)
)"#,
    r#"
CREATE TABLE IF NOT EXISTS public_relations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  role TEXT DEFAULT '',
  contact TEXT DEFAULT '',
  influence TEXT NOT NULL DEFAULT 'medium',
  sentiment TEXT NOT NULL DEFAULT 'neutral',
  description TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
)"#,
    r#"
CREATE TABLE IF NOT EXISTS verification_sessions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  project_id INTEGER REFERENCES projects(id),
  title TEXT NOT NULL DEFAULT '',
  verify_prompt TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'draft',
  verdict TEXT NOT NULL DEFAULT '',
  summary TEXT NOT NULL DEFAULT '',
  result_html TEXT NOT NULL DEFAULT '',
  result_markdown TEXT NOT NULL DEFAULT '',
  findings_json TEXT NOT NULL DEFAULT '[]',
  file_bundle_json TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
)"#,
    r#"
CREATE TABLE IF NOT EXISTS verification_files (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id INTEGER NOT NULL REFERENCES verification_sessions(id) ON DELETE CASCADE,
  organization_id INTEGER NOT NULL,
  original_name TEXT NOT NULL,
  stored_name TEXT NOT NULL,
  mime_type TEXT DEFAULT '',
  size_bytes INTEGER NOT NULL DEFAULT 0,
  excerpt TEXT NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
)"#,
    r#"
CREATE TABLE IF NOT EXISTS project_import_sessions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'analyzing',
  instruction TEXT NOT NULL DEFAULT '',
  draft_json TEXT NOT NULL DEFAULT '{}',
  error TEXT NOT NULL DEFAULT '',
  created_project_ids TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
)"#,
    r#"
CREATE TABLE IF NOT EXISTS project_import_files (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id INTEGER NOT NULL REFERENCES project_import_sessions(id) ON DELETE CASCADE,
  organization_id INTEGER NOT NULL,
  original_name TEXT NOT NULL,
  stored_name TEXT NOT NULL DEFAULT '',
  mime_type TEXT DEFAULT '',
  size_bytes INTEGER NOT NULL DEFAULT 0,
  excerpt TEXT NOT NULL DEFAULT '',
  error TEXT NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
)"#,
    "CREATE INDEX IF NOT EXISTS idx_project_import_sessions_org ON project_import_sessions(organization_id)",
    "CREATE INDEX IF NOT EXISTS idx_project_import_files_session ON project_import_files(session_id)",
    r#"CREATE TABLE IF NOT EXISTS flow_log_entries (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  organization_id INTEGER NOT NULL REFERENCES organizations(id),
  entry_date TEXT NOT NULL,
  direction TEXT NOT NULL,
  amount REAL NOT NULL,
  currency TEXT NOT NULL DEFAULT 'USD',
  category TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'logged',
  title TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  tags TEXT NOT NULL DEFAULT '[]',
  created_by INTEGER,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
)"#,
    "ALTER TABLE projects ADD COLUMN markdown_content TEXT NOT NULL DEFAULT ''",
    "ALTER TABLE projects ADD COLUMN html_content TEXT NOT NULL DEFAULT ''",
    "ALTER TABLE projects ADD COLUMN source_summary TEXT NOT NULL DEFAULT ''",
    "ALTER TABLE projects ADD COLUMN published_slug TEXT",
    "ALTER TABLE projects ADD COLUMN published_path TEXT",
];

pub async fn run(pool: &SqlitePool) -> Result<(), sqlx::Error> {
    for stmt in SCHEMA.split(';').map(str::trim).filter(|s| !s.is_empty()) {
        sqlx::query(stmt).execute(pool).await?;
    }
    for stmt in MIGRATIONS {
        let _ = sqlx::query(stmt).execute(pool).await;
    }
    // Migrate legacy documents → resource_files when old table exists
    let has_docs = sqlx::query_scalar::<_, i32>(
        "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='documents'",
    )
    .fetch_one(pool)
    .await
    .unwrap_or(0);
    if has_docs > 0 {
        let _ = sqlx::query(
            r#"
            INSERT INTO resource_files (organization_id, name, source_type, file_url, file_name, description, created_at)
            SELECT d.organization_id, d.name, 'url', d.file_url, '', COALESCE(d.notes, ''), d.created_at
            FROM documents d
            WHERE NOT EXISTS (
              SELECT 1 FROM resource_files rf
              WHERE rf.organization_id = d.organization_id AND rf.name = d.name AND rf.file_url = d.file_url
            )
            "#,
        )
        .execute(pool)
        .await;
        let _ = sqlx::query(
            r#"
            INSERT OR IGNORE INTO project_resource_files (project_id, resource_file_id)
            SELECT d.project_id, rf.id
            FROM documents d
            JOIN resource_files rf ON rf.organization_id = d.organization_id AND rf.name = d.name AND rf.file_url = d.file_url
            "#,
        )
        .execute(pool)
        .await;
    }
    Ok(())
}
