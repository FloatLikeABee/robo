-- Ensures a row for created_by=1 (FK target for templates / saved emails).
INSERT INTO users (id, email, name) VALUES (1, '__builtin_seed@local', 'System')
ON DUPLICATE KEY UPDATE id = id;

-- Adds tag + description (for UI / future AI) and builtin_key for seeded defaults.
ALTER TABLE email_templates
  ADD COLUMN tag VARCHAR(64) NOT NULL DEFAULT '' AFTER name,
  ADD COLUMN description VARCHAR(1024) NOT NULL DEFAULT '' AFTER tag,
  ADD COLUMN builtin_key VARCHAR(64) NULL AFTER description,
  ADD UNIQUE KEY uk_email_templates_builtin_key (builtin_key);
