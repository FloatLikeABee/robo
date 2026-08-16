-- One-off migration for existing databases.
-- mysql ... tran < schema_migration_contacts.sql

CREATE TABLE IF NOT EXISTS email_contacts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL DEFAULT '',
  email VARCHAR(320) NOT NULL,
  phone VARCHAR(64) NOT NULL DEFAULT '',
  note VARCHAR(1000) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_email_contacts_email (email),
  KEY idx_email_contacts_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
