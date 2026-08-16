package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/robo/webresearch"
)

func gatherWebContext(query string) (notes string, sources []webresearch.Source) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", nil
	}
	return webresearch.Gather(query)
}

func resolveWebSearchQuery(messages []composerAIMessage, explicit, fallback string) string {
	if q := strings.TrimSpace(explicit); q != "" {
		return q
	}
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			if q := strings.TrimSpace(messages[i].Content); q != "" {
				return q
			}
		}
	}
	return strings.TrimSpace(fallback)
}

func webSearchEnabledComposer(useFlag *bool) bool {
	if useFlag == nil {
		return true
	}
	return *useFlag
}

// WebSearch handles public web research for Morph AI tool loops.
func (a *App) WebSearch(c *gin.Context) {
	var body struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	query := strings.TrimSpace(body.Query)
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}
	notes, sources := gatherWebContext(query)
	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"notes":   notes,
		"sources": sources,
	})
}

// ComposerXAppAbilitiesMCP documents Morph AI generation abilities for email and publish pages.
func (a *App) ComposerXAppAbilitiesMCP(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":     "composerx-app-abilities",
		"description": "Morph AI app abilities for ComposerX: templates, saved markdown, reference-doc search, web-grounded email content, and public publish pages.",
		"tools": []gin.H{
			{
				"name":        "composerx_web_search",
				"description": "Search public sources for email/page content context.",
				"method":      "POST",
				"path":        "/ai/web-search",
				"body_example": gin.H{
					"query": "nonprofit fundraising email best practices",
				},
			},
			{
				"name":        "composerx_composer_chat",
				"description": "Generate or refine markdown documents with optional web search and reference-doc RAG.",
				"method":      "POST",
				"path":        "/ai/composer-chat",
				"body_example": gin.H{
					"messages":         []gin.H{{"role": "user", "content": "Draft a workshop agenda as a markdown document"}},
					"use_web_search":   true,
					"current_markdown": "",
				},
				"response_keys": []string{"assistant_message", "proposed_markdown", "research_notes", "sources"},
			},
			{
				"name":        "composerx_publish_chat",
				"description": "Generate or refine a full public HTML page with optional web search.",
				"method":      "POST",
				"path":        "/ai/publish-chat",
				"body_example": gin.H{
					"messages":       []gin.H{{"role": "user", "content": "Create a product launch landing page"}},
					"use_web_search": true,
					"theme":          "modern dark",
					"current_html":   "",
				},
				"response_keys": []string{"assistant_message", "proposed_page_html", "research_notes", "sources"},
			},
			{
				"name":        "composerx_assistant_chat",
				"description": "Platform assistant for templates, saved emails, reference docs, UsersPanel threads, and safe creates.",
				"method":      "POST",
				"path":        "/ai/assistant/chat",
			},
			{
				"name":        "composerx_list_templates",
				"description": "List email templates before fetching one by id.",
				"method":      "POST",
				"path":        "/ai/assistant/chat",
				"tool_call":   gin.H{"tool": "list_templates", "args": gin.H{"limit": 10, "offset": 0}},
			},
			{
				"name":        "composerx_get_template",
				"description": "Fetch one email template by id.",
				"method":      "POST",
				"path":        "/ai/assistant/chat",
				"tool_call":   gin.H{"tool": "get_template", "args": gin.H{"id": 1}},
			},
			{
				"name":        "composerx_search_reference_docs",
				"description": "Search reference files uploaded while composing, before using generic writing advice.",
				"method":      "POST",
				"path":        "/ai/assistant/chat",
				"tool_call":   gin.H{"tool": "search_reference_docs", "args": gin.H{"query": "brand voice", "limit": 10}},
			},
		},
	})
}
