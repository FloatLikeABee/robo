package models

import (
	"time"

	"gorm.io/gorm"
)

type Form struct {
	ID                 int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Name               string         `json:"name" gorm:"size:255;not null"`
	Description        string         `json:"description" gorm:"type:text"`
	Slug               string         `json:"slug" gorm:"size:255;uniqueIndex;not null"`
	SingleResponseOnly bool           `json:"single_response_only" gorm:"type:tinyint(1);default:0"`
	ExamMode           bool           `json:"exam_mode" gorm:"type:tinyint(1);default:0"`
	LandingHTML        string         `json:"landing_html" gorm:"type:longtext"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`

	Questions []Question `json:"questions,omitempty" gorm:"foreignKey:FormID"`
}

func (Form) TableName() string { return "forms" }

type CreateFormRequest struct {
	Name               string `json:"name" binding:"required"`
	Description        string `json:"description"`
	Slug               string `json:"slug" binding:"required"`
	SingleResponseOnly bool   `json:"single_response_only"`
	ExamMode           bool   `json:"exam_mode"`
	LandingHTML        string `json:"landing_html"`
}

type UpdateFormRequest struct {
	Name               *string `json:"name"`
	Description        *string `json:"description"`
	Slug               *string `json:"slug"`
	SingleResponseOnly *bool   `json:"single_response_only"`
	ExamMode           *bool   `json:"exam_mode"`
	LandingHTML        *string `json:"landing_html"`
}

type FormListResponse struct {
	Forms []Form `json:"forms"`
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}
