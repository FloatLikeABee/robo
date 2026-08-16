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
