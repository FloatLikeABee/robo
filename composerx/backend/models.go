package main

import (
	"encoding/json"
	"time"
)

// users
type User struct {
	ID        int64     `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// email_templates (API / list rows; IsBuiltin is derived from builtin_key in DB)
type EmailTemplate struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Tag         string    `db:"tag" json:"tag"`
	Description string    `db:"description" json:"description"`
	HTMLContent string    `db:"html_content" json:"html_content"`
	IsBuiltin   bool      `db:"-" json:"is_builtin"`
	CreatedBy   int64     `db:"created_by" json:"created_by"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// merge_data_sources
type MergeDataSource struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	FileType  string    `db:"file_type" json:"file_type"`
	FilePath  string    `db:"file_path" json:"file_path"`
	UploadedBy int64    `db:"uploaded_by" json:"uploaded_by"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// report_configs
type ReportConfig struct {
	ID              int64  `db:"id" json:"id"`
	ReportName      string `db:"report_name" json:"report_name"`
	ReportServiceURL string `db:"report_service_url" json:"report_service_url"`
	ReportType      string `db:"report_type" json:"report_type"`
	MaxFilesAllowed int    `db:"max_files_allowed" json:"max_files_allowed"`
}

// saved_emails (SQLite index; markdown body lives in Badger)
type SavedEmailDetail struct {
	ID                int64           `json:"id"`
	Name              string          `json:"name"`
	MarkdownContent   string          `json:"markdown_content"`
	ComposerAISession json.RawMessage `json:"composer_ai_session,omitempty"`
	CreatedBy         int64           `json:"created_by"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

