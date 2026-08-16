package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Question struct {
	ID        int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	FormID    int64          `json:"form_id" gorm:"index;not null"`
	PageID    int64          `json:"page_id" gorm:"index"`
	Title     string         `json:"title" gorm:"size:512;not null"`
	Type      string         `json:"type" gorm:"size:32;not null"` // text, select, multiselect, boolean, image, document (+ legacy types still readable)
	Required  bool           `json:"required" gorm:"type:tinyint(1);default:0"`
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	Config    QuestionConfig `json:"config" gorm:"type:json"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Question) TableName() string { return "questions" }

// QuestionConfig holds type-specific options (e.g. options for select, max_size for file).
type QuestionConfig struct {
	Multiline         bool         `json:"multiline,omitempty"`
	Options           []OptionItem `json:"options,omitempty"`
	MaxSelections     int          `json:"max_selections,omitempty"`
	Min               *float64     `json:"min,omitempty"`
	Max               *float64     `json:"max,omitempty"`
	MaxSize           int64        `json:"max_size,omitempty"`
	AllowedMime       string       `json:"allowed_mime,omitempty"`
	AllowedExt        []string     `json:"allowed_extensions,omitempty"`
	ValidationPattern string       `json:"validation_pattern,omitempty"`
	// DefaultValue is the preset shown on the public form (respondent may keep or change it). JSON shape matches the answer type (string, number, bool, []number for multiselect, etc.).
	DefaultValue interface{} `json:"default_value,omitempty"`
	// QuestionPromptMedia optional image or video for the prompt (never both).
	QuestionPromptMedia *QuestionPromptMedia `json:"question_prompt_media,omitempty"`
}

type OptionItem struct {
	Value int64  `json:"value"`
	Label string `json:"label"`
}

func (c QuestionConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *QuestionConfig) Scan(value interface{}) error {
	if value == nil {
		*c = QuestionConfig{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for QuestionConfig")
	}
	return json.Unmarshal(b, c)
}

type CreateQuestionRequest struct {
	Title     string         `json:"title" binding:"required"`
	Type      string         `json:"type" binding:"required"`
	Required  bool           `json:"required"`
	PageID    int64          `json:"page_id"`
	SortOrder int            `json:"sort_order"`
	Config    QuestionConfig `json:"config"`
}

type UpdateQuestionRequest struct {
	Title     *string         `json:"title"`
	Type      *string         `json:"type"`
	Required  *bool           `json:"required"`
	PageID    *int64          `json:"page_id"`
	SortOrder *int            `json:"sort_order"`
	Config    *QuestionConfig `json:"config"`
}

// Question types supported for create/edit.
const (
	QTypeText        = "text"
	QTypeInteger     = "integer" // legacy — not creatable
	QTypeSelect      = "select"
	QTypeMultiselect = "multiselect"
	QTypeFloat       = "float" // legacy — not creatable
	QTypeBoolean     = "boolean"
	QTypeDate        = "date"     // legacy — not creatable
	QTypeDatetime    = "datetime" // legacy — not creatable
	QTypeImage       = "image"
	QTypeDocument    = "document"
	QTypeQRCode      = "qrcode" // legacy — not creatable
)

// ValidQuestionTypes are allowed when creating or updating questions.
var ValidQuestionTypes = map[string]bool{
	QTypeText: true, QTypeSelect: true, QTypeMultiselect: true,
	QTypeBoolean: true, QTypeImage: true, QTypeDocument: true,
}

// LegacyQuestionTypes remain valid for existing forms and answer validation.
var LegacyQuestionTypes = map[string]bool{
	QTypeInteger: true, QTypeFloat: true, QTypeDate: true, QTypeDatetime: true, QTypeQRCode: true,
}

// IsKnownQuestionType reports whether a type is creatable or legacy.
func IsKnownQuestionType(t string) bool {
	return ValidQuestionTypes[t] || LegacyQuestionTypes[t]
}
