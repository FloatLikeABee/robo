-- Timelines: AI timeline documents from file/URL/paste sources (replaces Stories UX)
CREATE TABLE IF NOT EXISTS timeline (
  id INT NOT NULL AUTO_INCREMENT,
  user_id INT NOT NULL,
  owner_key VARCHAR(191) NOT NULL DEFAULT '',
  title VARCHAR(255) NOT NULL,
  source_summary TEXT NOT NULL,
  source_file_name VARCHAR(512) NULL,
  source_url VARCHAR(2048) NULL,
  has_paste TINYINT(1) NOT NULL DEFAULT 0,
  markdown_content LONGTEXT NOT NULL,
  html_content LONGTEXT NOT NULL,
  published_slug VARCHAR(255) NULL,
  published_path VARCHAR(512) NULL,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY ix_timeline_user (user_id),
  KEY ix_timeline_owner (owner_key),
  KEY ix_timeline_created (created_on),
  UNIQUE KEY uq_timeline_published_slug (published_slug)
);
