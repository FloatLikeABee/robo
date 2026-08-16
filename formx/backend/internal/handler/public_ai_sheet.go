package handler

import (
	"net/http"
	"strings"

	"github.com/formsx/backend/internal/surveybot"
	"github.com/gin-gonic/gin"
)

type publicAISheetChatBody struct {
	SessionID string                 `json:"session_id"`
	Message   string                 `json:"message"`
	State     *assistantConversation `json:"state"`
}

// PublicGetAISheet GET /public/ai-sheets/:slug — published AI Sheet metadata.
func (h *Handler) PublicGetAISheet(c *gin.Context) {
	if h.SurveyBotTemplateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "not configured"})
		return
	}
	slug := strings.TrimSpace(c.Param("slug"))
	t, err := h.SurveyBotTemplateRepo.GetPublishedBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sheet not found or not published"})
		return
	}
	parsed, _ := surveybot.ParseMarkdown(t.Markdown)
	instructions := t.Summary
	stepCount := 0
	if parsed != nil {
		instructions = parsed.Instructions
		stepCount = len(parsed.Steps)
	}
	c.JSON(http.StatusOK, gin.H{
		"slug":         t.Slug,
		"title":        t.Title,
		"summary":      t.Summary,
		"instructions": instructions,
		"step_count":   stepCount,
	})
}

// PublicAISheetChat POST /public/ai-sheets/:slug/chat — scoped survey bot for one published sheet.
func (h *Handler) PublicAISheetChat(c *gin.Context) {
	if h.SurveyBotTemplateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "not configured"})
		return
	}
	slug := strings.TrimSpace(c.Param("slug"))
	t, err := h.SurveyBotTemplateRepo.GetPublishedBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sheet not found or not published"})
		return
	}
	parsed, err := surveybot.ParseMarkdown(t.Markdown)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := surveybot.RequireSteps(parsed); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body publicAISheetChatBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	msg := strings.TrimSpace(body.Message)
	st := assistantConversation{}
	if body.State != nil {
		st = *body.State
	}
	ensureSurveyBotState(&st)
	st.Intent = "survey_bot"
	sb := st.SurveyBot

	active := sb.TemplateID == t.ID &&
		(sb.Status == "running" || sb.Status == "awaiting_ui" || sb.Status == "completed")

	var turn surveyBotTurn
	if !active || strings.EqualFold(msg, "start") || strings.EqualFold(msg, "restart") {
		turn = h.startSurveyFromTemplate(c.Request.Context(), st, t)
	} else if field, value, ok := surveybot.ParseSurveyAnswerMessage(msg); ok {
		turn = h.applySurveyAnswer(c.Request.Context(), st, field, value)
	} else if sb.Status == "completed" {
		turn = surveyBotTurn{
			Message: "This survey is already complete. Refresh the page or say **restart** to begin again.",
			State:   st,
			Done:    true,
		}
	} else {
		turn = h.continueSurveyWithUserText(c.Request.Context(), st, msg)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   turn.Message,
		"state":     turn.State,
		"ui_blocks": turn.UIBlocks,
		"done":      turn.Done,
		"record":    turn.Record,
		"title":     t.Title,
		"slug":      t.Slug,
	})
}
