-- Persisted imported data tables (flat file uploads)
CREATE TABLE IF NOT EXISTS data_tables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    source_filename TEXT,
    source_format VARCHAR(32) NOT NULL,
    column_schema JSONB NOT NULL,
    row_count INTEGER NOT NULL DEFAULT 0,
    created_by TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS data_table_rows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id UUID NOT NULL REFERENCES data_tables(id) ON DELETE CASCADE,
    row_index INTEGER NOT NULL,
    data JSONB NOT NULL,
    UNIQUE(table_id, row_index)
);

CREATE INDEX IF NOT EXISTS idx_data_table_rows_table_id ON data_table_rows(table_id);
