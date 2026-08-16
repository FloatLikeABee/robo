-- Recent AI-built page drafts per data table (max 5 retained in application code)
CREATE TABLE IF NOT EXISTS data_page_builds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_table_id UUID NOT NULL REFERENCES data_tables(id) ON DELETE CASCADE,
    label VARCHAR(255) NOT NULL DEFAULT 'Build',
    source VARCHAR(32) NOT NULL DEFAULT 'build',
    html_content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_data_page_builds_table_created ON data_page_builds(data_table_id, created_at DESC);
