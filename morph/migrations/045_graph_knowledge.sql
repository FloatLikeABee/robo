-- Morph GraphRAG outbox + Morph-only knowledge library
CREATE TABLE IF NOT EXISTS graph_sync_outbox (
  id BIGINT NOT NULL AUTO_INCREMENT,
  source VARCHAR(32) NOT NULL,
  entity_type VARCHAR(64) NOT NULL,
  entity_id VARCHAR(128) NOT NULL,
  op ENUM('upsert','delete') NOT NULL,
  payload_json JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  available_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  attempts INT NOT NULL DEFAULT 0,
  locked_by VARCHAR(64) NULL,
  locked_at DATETIME(3) NULL,
  processed_at DATETIME(3) NULL,
  last_error TEXT NULL,
  PRIMARY KEY (id),
  KEY idx_claim (processed_at, available_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS morph_knowledge_files (
  id BIGINT NOT NULL AUTO_INCREMENT,
  title VARCHAR(512) NOT NULL,
  filename VARCHAR(512) NOT NULL,
  content_type VARCHAR(128) NOT NULL DEFAULT '',
  kind VARCHAR(32) NOT NULL,
  storage_path VARCHAR(1024) NOT NULL,
  byte_size BIGINT NOT NULL DEFAULT 0,
  text_excerpt MEDIUMTEXT NULL,
  created_by VARCHAR(128) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_updated (updated_at),
  KEY idx_kind (kind)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS morph_knowledge_chunks (
  id BIGINT NOT NULL AUTO_INCREMENT,
  file_id BIGINT NOT NULL,
  chunk_index INT NOT NULL,
  text_content MEDIUMTEXT NOT NULL,
  embedding_json LONGTEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uq_file_chunk (file_id, chunk_index),
  KEY idx_file (file_id),
  CONSTRAINT fk_knowledge_chunks_file FOREIGN KEY (file_id) REFERENCES morph_knowledge_files(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
