-- Big Notes: AI HTML/MD notes + questionnaires with responses/analysis
CREATE TABLE IF NOT EXISTS big_note (
  id INT NOT NULL AUTO_INCREMENT,
  user_id INT NOT NULL,
  owner_key VARCHAR(191) NOT NULL DEFAULT '',
  title VARCHAR(255) NOT NULL,
  idea TEXT NOT NULL,
  note_kind VARCHAR(32) NOT NULL DEFAULT 'note',
  markdown_content LONGTEXT NOT NULL,
  html_content LONGTEXT NOT NULL,
  questions_json LONGTEXT NULL,
  theme VARCHAR(64) NOT NULL DEFAULT 'default',
  published_slug VARCHAR(255) NULL,
  published_path VARCHAR(512) NULL,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY ix_big_note_user (user_id),
  KEY ix_big_note_owner (owner_key),
  KEY ix_big_note_created (created_on)
);

CREATE TABLE IF NOT EXISTS big_note_response (
  id INT NOT NULL AUTO_INCREMENT,
  big_note_id INT NOT NULL,
  answers_json LONGTEXT NOT NULL,
  analysis_markdown LONGTEXT NULL,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY ix_big_note_response_note (big_note_id),
  KEY ix_big_note_response_created (created_on)
);
