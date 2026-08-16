package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// FormsXAppAbilitiesMCP documents Morph AI generation abilities for FormsX (web-grounded form templates).
func (h *Handler) FormsXAppAbilitiesMCP(c *gin.Context) {
	base := "/api/v1"
	c.JSON(http.StatusOK, gin.H{
		"service":     "formsx-app-abilities",
		"description": "Morph AI app abilities for FormsX: web-grounded form HTML templates, question suggestions, and public web research.",
		"tools": []gin.H{
			{
				"name":        "formsx_web_search",
				"description": "Search public sources (Wikipedia, DuckDuckGo, arXiv) for form design context.",
				"method":      "POST",
				"path":        base + "/ai/web-search",
				"body_example": gin.H{
					"query": "customer satisfaction survey best practices",
				},
			},
			{
				"name":        "formsx_form_template_chat",
				"description": "Generate a web-grounded HTML form template and suggested questions. User edits and saves via PUT /forms/:id landing_html.",
				"method":      "POST",
				"path":        base + "/ai/form-template-chat",
				"body_example": gin.H{
					"messages":         []gin.H{{"role": "user", "content": "Create a volunteer signup form for a community garden"}},
					"form_name":        "Community garden volunteers",
					"form_description": "Collect volunteer interest and availability",
					"use_web_search":   true,
					"current_html":     "",
				},
				"response_keys": []string{"assistant_message", "proposed_form_html", "proposed_questions", "research_notes", "sources"},
			},
			{
				"name":        "formsx_assistant_chat",
				"description": "Platform assistant with CRUD tools (forms, questions, events) and web_search tool.",
				"method":      "POST",
				"path":        base + "/assistant/chat",
			},
			{
				"name":        "formsx_save_landing_html",
				"description": "Persist edited HTML template on a form.",
				"method":      "PUT",
				"path":        base + "/forms/{id}",
				"body_example": gin.H{
					"landing_html": "<!doctype html>…",
				},
			},
		},
	})
}
