package models

import (
	"time"
)

// AIDocument stores normalized system records for document-based AI retrieval.
type AIDocument struct {
	ID        string                 `json:"id"`
	DocType   string                 `json:"doc_type"`
	SourceID  string                 `json:"source_id"`
	Title     string                 `json:"title"`
	Summary   string                 `json:"summary,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func AIDocumentToMap(doc *AIDocument) map[string]interface{} {
	if doc == nil {
		return nil
	}
	return map[string]interface{}{
		"id":         doc.ID,
		"doc_type":   doc.DocType,
		"source_id":  doc.SourceID,
		"title":      doc.Title,
		"summary":    doc.Summary,
		"tags":       doc.Tags,
		"data":       doc.Data,
		"created_at": doc.CreatedAt,
		"updated_at": doc.UpdatedAt,
	}
}
