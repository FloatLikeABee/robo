package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/formsx/backend/internal/models"
	"github.com/formsx/backend/internal/surveybot"
	"github.com/gin-gonic/gin"
)

type surveyBotPublishBody struct {
	Published *bool `json:"published"`
}

// PublishSurveyBotTemplate POST /survey-bot/templates/:id/publish
func (h *Handler) PublishSurveyBotTemplate(c *gin.Context) {
	h.setSurveyBotPublished(c, true)
}

// UnpublishSurveyBotTemplate POST /survey-bot/templates/:id/unpublish
func (h *Handler) UnpublishSurveyBotTemplate(c *gin.Context) {
	h.setSurveyBotPublished(c, false)
}

// SetSurveyBotTemplatePublished PUT /survey-bot/templates/:id/published
func (h *Handler) SetSurveyBotTemplatePublished(c *gin.Context) {
	var body surveyBotPublishBody
	if err := c.ShouldBindJSON(&body); err != nil || body.Published == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "published boolean required"})
		return
	}
	h.setSurveyBotPublished(c, *body.Published)
}

func (h *Handler) setSurveyBotPublished(c *gin.Context, published bool) {
	if h.SurveyBotTemplateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	t, err := h.SurveyBotTemplateRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	if published {
		parsed, err := surveybot.ParseMarkdown(t.Markdown)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := surveybot.RequireSteps(parsed); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		now := time.Now().UTC()
		t.Published = true
		t.PublishedAt = &now
	} else {
		t.Published = false
		t.PublishedAt = nil
	}
	if err := h.SurveyBotTemplateRepo.Update(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := models.SurveyBotTemplateToMap(t)
	base := strings.TrimRight(h.Cfg.PublicFormBaseURL, "/")
	if base == "" {
		base = ""
	}
	if t.Published {
		out["public_url"] = base + "/s/" + t.Slug
	}
	c.JSON(http.StatusOK, out)
}
