package models

import (
	"time"

	"gorm.io/gorm"
)

type FormPage struct {
	ID        int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	FormID    int64          `json:"form_id" gorm:"index;not null"`
	Name      string         `json:"name" gorm:"size:255;not null;default:''"`
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (FormPage) TableName() string { return "form_pages" }

type CreatePageRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type UpdatePageRequest struct {
	Name      *string `json:"name"`
	SortOrder *int    `json:"sort_order"`
}
