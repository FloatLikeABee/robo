-- MergeEmailX core MySQL schema
-- This mirrors the design.md database section.

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  email VARCHAR(320) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS email_templates (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  tag VARCHAR(64) NOT NULL DEFAULT '',
  description VARCHAR(1024) NOT NULL DEFAULT '',
  builtin_key VARCHAR(64) NULL,
  html_content MEDIUMTEXT NOT NULL,
  created_by BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_email_templates_builtin_key (builtin_key),
  KEY idx_email_templates_created_by (created_by),
  CONSTRAINT fk_email_templates_user FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS merge_data_sources (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  file_type ENUM('csv', 'json', 'excel') NOT NULL,
  file_path VARCHAR(1024) NOT NULL,
  uploaded_by BIGINT UNSIGNED NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_merge_data_sources_name (name),
  KEY idx_merge_data_sources_uploaded_by (uploaded_by),
  CONSTRAINT fk_merge_data_sources_user FOREIGN KEY (uploaded_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS saved_emails (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  content_mongo_id CHAR(24) NOT NULL,
  created_by BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_saved_emails_mongo_id (content_mongo_id),
  KEY idx_saved_emails_updated_at (updated_at),
  CONSTRAINT fk_saved_emails_user FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS published_pages (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL,
  theme VARCHAR(128) NOT NULL DEFAULT 'default',
  content_mongo_id CHAR(24) NOT NULL,
  created_by BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_published_pages_slug (slug),
  UNIQUE KEY uk_published_pages_mongo_id (content_mongo_id),
  KEY idx_published_pages_updated_at (updated_at),
  CONSTRAINT fk_published_pages_user FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS publish_drafts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  theme VARCHAR(128) NOT NULL DEFAULT 'default',
  html_content MEDIUMTEXT NOT NULL,
  created_by BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_publish_drafts_updated_at (updated_at),
  CONSTRAINT fk_publish_drafts_user FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS report_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  report_name VARCHAR(255) NOT NULL,
  report_service_url VARCHAR(1024) NOT NULL,
  report_type VARCHAR(64) NOT NULL,
  max_files_allowed INT NOT NULL DEFAULT 1,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Optional cleanup on existing databases (run manually):
-- DROP TABLE IF EXISTS email_report_orders;
-- DROP TABLE IF EXISTS email_recipients;
-- DROP TABLE IF EXISTS email_jobs;
-- DROP TABLE IF EXISTS saved_email_recipients;
-- DROP TABLE IF EXISTS email_contacts;
-- ALTER TABLE saved_emails DROP COLUMN recipient_count, DROP COLUMN recipients_preview;
