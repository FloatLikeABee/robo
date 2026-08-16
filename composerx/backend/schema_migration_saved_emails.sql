-- Run against existing DBs that already applied schema.sql before saved_emails.
-- mysql ... tran < schema_migration_saved_emails.sql

CREATE TABLE IF NOT EXISTS saved_emails (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  content_mongo_id CHAR(24) NOT NULL,
  recipient_count INT NOT NULL DEFAULT 0,
  recipients_preview VARCHAR(512) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_saved_emails_mongo_id (content_mongo_id),
  KEY idx_saved_emails_updated_at (updated_at),
  CONSTRAINT fk_saved_emails_user FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS saved_email_recipients (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  saved_email_id BIGINT UNSIGNED NOT NULL,
  email VARCHAR(320) NOT NULL,
  merge_key VARCHAR(255) DEFAULT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  KEY idx_saved_email_recipients_saved_email_id (saved_email_id),
  CONSTRAINT fk_saved_email_recipients_saved_email FOREIGN KEY (saved_email_id) REFERENCES saved_emails(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE email_jobs
  MODIFY template_id BIGINT UNSIGNED NULL,
  ADD COLUMN saved_email_id BIGINT UNSIGNED NULL AFTER template_id;

ALTER TABLE email_jobs
  ADD KEY idx_email_jobs_saved_email_id (saved_email_id);

ALTER TABLE email_jobs
  ADD CONSTRAINT fk_email_jobs_saved_email FOREIGN KEY (saved_email_id) REFERENCES saved_emails(id);
