-- FormsX: create database and tables (optional; GORM AutoMigrate also creates tables)
CREATE DATABASE IF NOT EXISTS tran
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE tran;

-- Forms table (GORM will create if not exists; this is for reference / manual setup)
CREATE TABLE IF NOT EXISTS forms (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  slug VARCHAR(255) NOT NULL,
  single_response_only TINYINT(1) DEFAULT 0,
  exam_mode TINYINT(1) DEFAULT 0,
  created_at DATETIME(3),
  updated_at DATETIME(3),
  deleted_at DATETIME(3),
  PRIMARY KEY (id),
  UNIQUE KEY idx_forms_slug (slug),
  KEY idx_forms_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS form_pages (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  form_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(255) NOT NULL DEFAULT '',
  sort_order INT DEFAULT 0,
  created_at DATETIME(3),
  updated_at DATETIME(3),
  deleted_at DATETIME(3),
  PRIMARY KEY (id),
  KEY idx_form_pages_form_id (form_id),
  KEY idx_form_pages_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS questions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  form_id BIGINT UNSIGNED NOT NULL,
  page_id BIGINT UNSIGNED DEFAULT NULL,
  title VARCHAR(512) NOT NULL,
  type VARCHAR(32) NOT NULL,
  required TINYINT(1) DEFAULT 0,
  sort_order INT DEFAULT 0,
  config JSON,
  created_at DATETIME(3),
  updated_at DATETIME(3),
  deleted_at DATETIME(3),
  PRIMARY KEY (id),
  KEY idx_questions_form_id (form_id),
  KEY idx_questions_page_id (page_id),
  KEY idx_questions_sort (form_id, sort_order),
  KEY idx_questions_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS question_rules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  question_id BIGINT UNSIGNED NOT NULL,
  depends_on_question_id BIGINT UNSIGNED NOT NULL,
  `condition` VARCHAR(32) NOT NULL,
  created_at DATETIME(3),
  updated_at DATETIME(3),
  deleted_at DATETIME(3),
  PRIMARY KEY (id),
  UNIQUE KEY idx_rule_question_depends (question_id, depends_on_question_id),
  KEY idx_question_rules_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- workspace_events collection is created by FormsX on first write (MongoDB / athena).

-- Dropped in app migration: form_templates (replaced by Events & Info in MongoDB).
-- Run on existing databases if you want to remove the old table:
-- DROP TABLE IF EXISTS form_templates;
