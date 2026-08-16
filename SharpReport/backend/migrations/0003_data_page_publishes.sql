-- AI-built public data pages (ComposerX-style publish URLs)
CREATE TABLE IF NOT EXISTS data_page_publishes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_table_id UUID NOT NULL REFERENCES data_tables(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    theme VARCHAR(64) NOT NULL DEFAULT 'light',
    html_content TEXT NOT NULL,
    created_by TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_data_page_publishes_table_id ON data_page_publishes(data_table_id);
CREATE INDEX IF NOT EXISTS idx_data_page_publishes_slug ON data_page_publishes(slug);
