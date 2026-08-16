-- Generic data: user-imported CSV/JSON/PDF material (detail JSON in Mongo entity_details)

CREATE TABLE IF NOT EXISTS generic_data (
  id INT NOT NULL AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  source_type ENUM('csv', 'json', 'pdf') NOT NULL,
  source_filename VARCHAR(500) NULL,
  record_count INT NOT NULL DEFAULT 0,
  description TEXT NULL,
  ai_analysis TEXT NULL,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY ix_generic_data_source_type (source_type),
  KEY ix_generic_data_created_on (created_on)
);
