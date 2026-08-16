package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/formsx/backend/internal/models"
	"github.com/formsx/backend/internal/mongo"
	"github.com/formsx/backend/internal/surveybot"
	"github.com/gin-gonic/gin"
	"github.com/robo/webresearch"
)

type surveyBotTemplateBody struct {
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	Tags      []string `json:"tags"`
	Markdown  string   `json:"markdown"`
	Summary   string   `json:"summary"`
	CreatedBy string   `json:"created_by"`
}

type surveyBotAIDraftBody struct {
	Query       string `json:"query"`
	TitleHint   string `json:"title_hint"`
	Description string `json:"description"`
}

func (h *Handler) ensureSurveyBotSeeds(ctx context.Context) {
	if h.SurveyBotTemplateRepo == nil {
		return
	}
	n, err := h.SurveyBotTemplateRepo.Count(ctx)
	if err != nil || n > 0 {
		return
	}
	seeds, err := surveybot.LoadSeeds()
	if err != nil {
		return
	}
	for _, s := range seeds {
		_ = h.SurveyBotTemplateRepo.Insert(ctx, &models.SurveyBotTemplate{
			Slug:     s.Slug,
			Title:    s.Title,
			Tags:     s.Tags,
			Markdown: s.Markdown,
			Summary:  s.Summary,
		})
	}
}

func queryInt(c *gin.Context, key string, def int) int {
	v := strings.TrimSpace(c.Query(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func truncateStr(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= max {
		return string(r)
	}
	return string(r[:max]) + "…"
}

// ListSurveyBotTemplates GET /survey-bot/templates
func (h *Handler) ListSurveyBotTemplates(c *gin.Context) {
	h.ensureSurveyBotSeeds(c.Request.Context())
	if h.SurveyBotTemplateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	page := queryInt(c, "page", 1)
	limit := queryInt(c, "limit", 50)
	search := strings.TrimSpace(c.Query("search"))
	list, total, err := h.SurveyBotTemplateRepo.List(c.Request.Context(), search, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		out = append(out, models.SurveyBotTemplateToMap(&list[i]))
	}
	c.JSON(http.StatusOK, gin.H{"templates": out, "total": total, "page": page, "limit": limit})
}

// CreateSurveyBotTemplate POST /survey-bot/templates
func (h *Handler) CreateSurveyBotTemplate(c *gin.Context) {
	if h.SurveyBotTemplateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	var body surveyBotTemplateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.Markdown = strings.TrimSpace(body.Markdown)
	if body.Markdown == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "markdown is required"})
		return
	}
	parsed, err := surveybot.ParseMarkdown(body.Markdown)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template: " + err.Error()})
		return
	}
	slug := strings.TrimSpace(body.Slug)
	if slug == "" {
		slug = parsed.Slug
	}
	if slug == "" {
		slug = normalizeSlug(parsed.Title)
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		title = parsed.Title
	}
	summary := strings.TrimSpace(body.Summary)
	if summary == "" {
		summary = truncateStr(parsed.Instructions, 160)
	}
	tags := body.Tags
	if len(tags) == 0 {
		tags = parsed.Tags
	}
	t := &models.SurveyBotTemplate{
		Slug:      slug,
		Title:     title,
		Tags:      tags,
		Markdown:  body.Markdown,
		Summary:   summary,
		CreatedBy: body.CreatedBy,
	}
	if err := h.SurveyBotTemplateRepo.Insert(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := models.SurveyBotTemplateToMap(t)
	out["needs_compile"] = parsed.NeedsCompile
	c.JSON(http.StatusCreated, out)
}

// GetSurveyBotTemplate GET /survey-bot/templates/:id
func (h *Handler) GetSurveyBotTemplate(c *gin.Context) {
	if h.SurveyBotTemplateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	id := c.Param("id")
	t, err := h.SurveyBotTemplateRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			t2, err2 := h.SurveyBotTemplateRepo.GetBySlug(c.Request.Context(), id)
			if err2 != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
				return
			}
			t = t2
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	out := models.SurveyBotTemplateToMap(t)
	if parsed, err := surveybot.ParseMarkdown(t.Markdown); err == nil {
		out["needs_compile"] = parsed.NeedsCompile
	}
	c.JSON(http.StatusOK, out)
}

// UpdateSurveyBotTemplate PUT /survey-bot/templates/:id
func (h *Handler) UpdateSurveyBotTemplate(c *gin.Context) {
	if h.SurveyBotTemplateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	t, err := h.SurveyBotTemplateRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	var body surveyBotTemplateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(body.Markdown) != "" {
		if _, err := surveybot.ParseMarkdown(body.Markdown); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template: " + err.Error()})
			return
		}
		t.Markdown = body.Markdown
	}
	if strings.TrimSpace(body.Title) != "" {
		t.Title = body.Title
	}
	if strings.TrimSpace(body.Slug) != "" {
		t.Slug = body.Slug
	}
	if body.Tags != nil {
		t.Tags = body.Tags
	}
	if body.Summary != "" {
		t.Summary = body.Summary
	}
	if err := h.SurveyBotTemplateRepo.Update(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.SurveyBotTemplateToMap(t))
}

// DeleteSurveyBotTemplate DELETE /survey-bot/templates/:id
func (h *Handler) DeleteSurveyBotTemplate(c *gin.Context) {
	if h.SurveyBotTemplateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	if err := h.SurveyBotTemplateRepo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DraftSurveyBotTemplateAI POST /survey-bot/templates/ai-draft
func (h *Handler) DraftSurveyBotTemplateAI(c *gin.Context) {
	var body surveyBotAIDraftBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := strings.TrimSpace(body.Query)
	if query == "" {
		query = strings.TrimSpace(body.TitleHint + " " + body.Description)
	}
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query or title_hint required"})
		return
	}
	notes, sources := webresearch.Gather(query + " survey questionnaire best practices")
	md := draftSurveyMarkdown(query, body.TitleHint, body.Description, notes)
	if _, err := surveybot.ParseMarkdown(md); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "draft failed validation: " + err.Error(), "markdown": md})
		return
	}
	srcOut := make([]gin.H, 0, len(sources))
	for _, s := range sources {
		srcOut = append(srcOut, gin.H{"title": s.Title, "type": s.Type, "url": s.URL})
	}
	c.JSON(http.StatusOK, gin.H{
		"markdown":          md,
		"research_notes":    notes,
		"sources":           srcOut,
		"assistant_message": "Draft Survey Bot markdown ready. Review and save, or ask me to adjust questions.",
	})
}

func draftSurveyMarkdown(query, titleHint, description, notes string) string {
	title := strings.TrimSpace(titleHint)
	if title == "" {
		title = strings.TrimSpace(query)
	}
	if title == "" {
		title = "Custom survey"
	}
	slug := normalizeSlug(title)
	if slug == "" {
		slug = "custom-survey"
	}
	desc := strings.TrimSpace(description)
	if desc == "" {
		desc = "Collect answers one question at a time."
	}
	_ = notes
	return fmt.Sprintf(`---
slug: %s
title: %s
tags: [custom, survey-bot]
---

# Instructions
%s
Use selectors for categorical answers.

## Q1 — Name
- field: name
- collect: text
- required: true
- prompt: What is your name?

## Q2 — Category
- field: category
- collect: mcp_html
- widget: select
- options: [Option A, Option B, Option C, Other]
- required: true
- prompt: Please choose the best matching category.

## Q3 — Details
- field: details
- collect: text
- required: false
- prompt: Any additional details?
`, slug, title, desc)
}

// ListSurveyBotResults GET /survey-bot/results
func (h *Handler) ListSurveyBotResults(c *gin.Context) {
	if h.SurveyBotResultRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	page := queryInt(c, "page", 1)
	limit := queryInt(c, "limit", 50)
	search := strings.TrimSpace(c.Query("search"))
	list, total, err := h.SurveyBotResultRepo.List(c.Request.Context(), search, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		out = append(out, models.SurveyBotResultToMap(&list[i], false))
	}
	c.JSON(http.StatusOK, gin.H{"results": out, "total": total, "page": page, "limit": limit})
}

// GetSurveyBotResult GET /survey-bot/results/:id
func (h *Handler) GetSurveyBotResult(c *gin.Context) {
	if h.SurveyBotResultRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	res, err := h.SurveyBotResultRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
		return
	}
	c.JSON(http.StatusOK, models.SurveyBotResultToMap(res, true))
}

// GetSurveyBotResultHTML GET /survey-bot/results/:id/html
func (h *Handler) GetSurveyBotResultHTML(c *gin.Context) {
	if h.SurveyBotResultRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	res, err := h.SurveyBotResultRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, res.HTML)
}

// DeleteSurveyBotResult DELETE /survey-bot/results/:id
func (h *Handler) DeleteSurveyBotResult(c *gin.Context) {
	if h.SurveyBotResultRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	if err := h.SurveyBotResultRepo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
