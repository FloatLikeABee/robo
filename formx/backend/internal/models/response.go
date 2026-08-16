package models

import (
	"time"
)

// FormResponse is stored in Badger (former Mongo form_responses collection).
type FormResponse struct {
	ID             string        `json:"id"`
	FormID         int64         `json:"form_id"`
	Slug           string        `json:"slug"`
	RespondentID   string        `json:"respondent_id,omitempty"`
	SubmittedAt    time.Time     `json:"submitted_at"`
	Answers        []AnswerEntry `json:"answers"`
	ExamDurationMs int64         `json:"exam_duration_ms,omitempty"`
}

type AnswerEntry struct {
	QuestionID int64       `json:"question_id"`
	Type       string      `json:"type"`
	Value      interface{} `json:"value"`
	Filename   string      `json:"filename,omitempty"`
	Size       int64       `json:"size,omitempty"`
}

type SubmitRequest struct {
	RespondentID  string        `json:"respondent_id"`
	ExamStartedAt *time.Time    `json:"exam_started_at,omitempty"`
	Answers       []AnswerInput `json:"answers" binding:"required"`
}

type AnswerInput struct {
	QuestionID int64       `json:"question_id" binding:"required"`
	Value      interface{} `json:"value"`
}

type ResponseListResult struct {
	Responses []FormResponse `json:"responses"`
	Total     int64          `json:"total"`
	Page      int            `json:"page"`
	Limit     int            `json:"limit"`
}
