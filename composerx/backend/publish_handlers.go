package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
)

type publishPageCreateRequest struct {
	Name        string `json:"name"`
	Theme       string `json:"theme"`
	HTMLContent string `json:"html_content"`
	CreatedBy   int64  `json:"created_by"`
}

type publishAIRequest struct {
	Messages         []composerAIMessage     `json:"messages"`
	CurrentHTML      string                  `json:"current_html"`
	SourceText       string                  `json:"source_text"`
	SourceMaterials  []publishSourceMaterial `json:"source_materials"`
	Theme            string                  `json:"theme"`
	UseWebSearch     *bool                   `json:"use_web_search"`
	WebSearchQuery   string                  `json:"web_search_query"`
}

type publishAIResponse struct {
	AssistantMessage string                 `json:"assistant_message"`
	ProposedPageHTML *string                `json:"proposed_page_html,omitempty"`
	ResearchNotes    string                 `json:"research_notes,omitempty"`
	Sources          []webresearchSource    `json:"sources,omitempty"`
}

type webresearchSource struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	URL   string `json:"url,omitempty"`
}

type publishAIResponseFlexible struct {
	AssistantSnake string  `json:"assistant_message"`
	AssistantCamel string  `json:"assistantMessage"`
	ProposedSnake  *string `json:"proposed_page_html"`
	ProposedCamel  *string `json:"proposedPageHtml"`
}

func mergePublishAIFields(f publishAIResponseFlexible) (string, *string) {
	msg := strings.TrimSpace(f.AssistantSnake)
	if msg == "" {
		msg = strings.TrimSpace(f.AssistantCamel)
	}
	if f.ProposedSnake != nil && strings.TrimSpace(*f.ProposedSnake) != "" {
		return msg, f.ProposedSnake
	}
	if f.ProposedCamel != nil && strings.TrimSpace(*f.ProposedCamel) != "" {
		return msg, f.ProposedCamel
	}
	return msg, nil
}

func (a *App) resolvePublishSlug(c *gin.Context) {
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	slug, err := a.publishedPages.ResolveUniqueSlug(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve publish slug"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"slug": slug, "public_path": "/public/p/" + slug})
}

func (a *App) createPublishedPage(c *gin.Context) {
	var payload publishPageCreateRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	row, err := a.publishedPages.Create(
		c.Request.Context(),
		payload.Name,
		payload.Theme,
		payload.HTMLContent,
		payload.CreatedBy,
	)
	if err != nil {
		if err.Error() == "name and html required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish page: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":          row.ID,
		"name":        row.Name,
		"slug":        row.Slug,
		"theme":       row.Theme,
		"public_path": "/public/p/" + row.Slug,
	})
}

func parseListParams(c *gin.Context, defaultLimit int) (limit int, offset int) {
	limit = defaultLimit
	offset = 0
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return
}

func (a *App) listPublishedPages(c *gin.Context) {
	limit, offset := parseListParams(c, 50)
	items, total, err := a.publishedPages.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list published pages"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (a *App) createPublishDraft(c *gin.Context) {
	var payload publishPageCreateRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	id, err := a.publishDrafts.Create(c.Request.Context(), payload.Name, payload.Theme, payload.HTMLContent, payload.CreatedBy)
	if err != nil {
		if err.Error() == "name and html required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save publish draft"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (a *App) listPublishDrafts(c *gin.Context) {
	limit, offset := parseListParams(c, 50)
	items, total, err := a.publishDrafts.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list publish drafts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (a *App) getPublishDraft(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	row, err := a.publishDrafts.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "publish draft not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load publish draft"})
		return
	}
	c.JSON(http.StatusOK, row)
}

func (a *App) deletePublishDraft(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := a.publishDrafts.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "publish draft not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete publish draft"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *App) getPublishedPageJSON(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slug"})
		return
	}
	row, err := a.publishedPages.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "published page not found"})
		return
	}
	c.JSON(http.StatusOK, row)
}

func (a *App) servePublishedPage(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.String(http.StatusBadRequest, "invalid slug")
		return
	}
	row, err := a.publishedPages.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.String(http.StatusNotFound, "published page not found")
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, ensureFullHTMLDocument(row.HTMLContent, row.Name))
}

func ensureFullHTMLDocument(html, title string) string {
	trimmed := strings.TrimSpace(html)
	if trimmed == "" {
		return "<!doctype html><html><head><meta charset=\"utf-8\"><title>Published page</title></head><body></body></html>"
	}
	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "<html") && strings.Contains(lower, "<body") {
		return trimmed
	}
	safeTitle := strings.TrimSpace(title)
	if safeTitle == "" {
		safeTitle = "Published page"
	}
	return fmt.Sprintf("<!doctype html><html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>%s</title></head><body>%s</body></html>", safeTitle, trimmed)
}

func (a *App) publishAIChat(c *gin.Context) {
	chatClient, chatModel, ok := a.chatCompletionClient()
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI assistant is not configured"})
		return
	}

	var payload publishAIRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if len(payload.Messages) == 0 && strings.TrimSpace(payload.SourceText) == "" && buildCombinedPublishSourceText(payload.SourceMaterials) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message, source_text, or source_materials is required"})
		return
	}

	system := `You are an expert HTML page composer.
CRITICAL: Output one JSON object only with exactly these keys:
- "assistant_message": concise plain-text guidance for the user.
- "proposed_page_html": complete HTML for a public webpage (include <style> in the markup when useful), OR null if no page update is needed.

Goal:
- Transform plain text ideas into a polished, readable HTML page.
- Respect the requested theme.
- Keep semantic structure (header/main/section/footer where useful).
- Avoid scripts unless explicitly requested.`

	if strings.TrimSpace(payload.Theme) != "" {
		system += "\n\nRequested page theme: " + strings.TrimSpace(payload.Theme)
	}
	if strings.TrimSpace(payload.CurrentHTML) != "" {
		system += "\n\nCurrent draft HTML that can be improved or replaced:\n" + payload.CurrentHTML
	}
	if strings.TrimSpace(payload.SourceText) != "" {
		system += "\n\nSource text from the user:\n" + payload.SourceText
	}
	if combined := buildCombinedPublishSourceText(payload.SourceMaterials); combined != "" {
		system += "\n\nSummaries and descriptions extracted from user files (treat as ground truth for page content):\n" + combined
	}

	var researchNotes string
	var sources []webresearchSource
	if webSearchEnabledComposer(payload.UseWebSearch) {
		if q := resolveWebSearchQuery(payload.Messages, payload.WebSearchQuery, payload.SourceText); q != "" {
			notes, srcs := gatherWebContext(q)
			researchNotes = notes
			for _, s := range srcs {
				sources = append(sources, webresearchSource{Title: s.Title, Type: s.Type, URL: s.URL})
			}
			if researchNotes != "" {
				system += "\n\nWeb research (ground truth for topic and structure):\n" + researchNotes
			}
		}
	}

	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: system},
	}
	for _, m := range payload.Messages {
		switch m.Role {
		case "user":
			msgs = append(msgs, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: m.Content})
		case "assistant":
			msgs = append(msgs, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: m.Content})
		}
	}
	if len(msgs) == 1 && strings.TrimSpace(payload.SourceText) != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: "Turn the source text into a polished public webpage.",
		})
	}

	resp, err := chatClient.CreateChatCompletion(c.Request.Context(), openai.ChatCompletionRequest{
		Model:    chatModel,
		Messages: msgs,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("AI chat: %v", err)})
		return
	}
	if len(resp.Choices) == 0 || strings.TrimSpace(resp.Choices[0].Message.Content) == "" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "empty model response"})
		return
	}

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	var flex publishAIResponseFlexible
	if err := json.Unmarshal([]byte(raw), &flex); err != nil {
		c.JSON(http.StatusOK, publishAIResponse{AssistantMessage: raw})
		return
	}
	msg, html := mergePublishAIFields(flex)
	if msg == "" && html != nil {
		msg = "Page draft is ready. Apply it when you are happy."
	}
	if msg == "" {
		msg = raw
	}
	c.JSON(http.StatusOK, publishAIResponse{
		AssistantMessage: msg,
		ProposedPageHTML: html,
		ResearchNotes:    researchNotes,
		Sources:          sources,
	})
}
