package models

import (
	"time"
)

// SurveyBotTemplate is a markdown/txt-driven AI Sheets script stored in Badger.
type SurveyBotTemplate struct {
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Tags        []string   `json:"tags,omitempty"`
	Markdown    string     `json:"markdown"`
	Summary     string     `json:"summary,omitempty"`
	Published   bool       `json:"published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedBy   string     `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// SurveyBotResult is a completed survey run saved as themed HTML + answers.
type SurveyBotResult struct {
	ID           string            `json:"id"`
	TemplateID   string            `json:"template_id"`
	TemplateSlug string            `json:"template_slug,omitempty"`
	Title        string            `json:"title"`
	Answers      map[string]string `json:"answers"`
	HTML         string            `json:"html,omitempty"`
	SessionID    string            `json:"session_id,omitempty"`
	CreatedBy    string            `json:"created_by,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

func SurveyBotTemplateToMap(t *SurveyBotTemplate) map[string]interface{} {
	if t == nil {
		return nil
	}
	m := map[string]interface{}{
		"id":         t.ID,
		"slug":       t.Slug,
		"title":      t.Title,
		"tags":       t.Tags,
		"markdown":   t.Markdown,
		"summary":    t.Summary,
		"published":  t.Published,
		"created_by": t.CreatedBy,
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	}
	if t.PublishedAt != nil {
		m["published_at"] = t.PublishedAt
	}
	return m
}

func SurveyBotResultToMap(r *SurveyBotResult, includeHTML bool) map[string]interface{} {
	if r == nil {
		return nil
	}
	m := map[string]interface{}{
		"id":            r.ID,
		"template_id":   r.TemplateID,
		"template_slug": r.TemplateSlug,
		"title":         r.Title,
		"answers":       r.Answers,
		"session_id":    r.SessionID,
		"created_by":    r.CreatedBy,
		"created_at":    r.CreatedAt,
	}
	if includeHTML {
		m["html"] = r.HTML
	}
	return m
}
