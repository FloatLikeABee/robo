package handler

import (
	"net/http"
	"time"

	"github.com/formsx/backend/internal/models"
	"github.com/formsx/backend/pkg/validator"
	"github.com/gin-gonic/gin"
)

const (
	examClockSkew         = 5 * time.Minute
	examMaxSessionHorizon = 24 * time.Hour
)

// PublicFormResponse is form + pages + questions + rules for public display.
type PublicFormResponse struct {
	Form      models.Form           `json:"form"`
	Pages     []models.FormPage     `json:"pages"`
	Questions []models.Question     `json:"questions"`
	Rules     []models.QuestionRule `json:"rules"`
}

// PublicGetForm returns form and questions by slug (for rendering published form).
// @Summary Get public form by slug
// @Tags public
// @Produce json
// @Param slug path string true "Slug"
// @Success 200 {object} PublicFormResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/public/forms/{slug} [get]
func (h *Handler) PublicGetForm(c *gin.Context) {
	slug := c.Param("slug")
	form, err := h.FormRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	if _, err := h.PageRepo.EnsureDefaultPage(form.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	pages, err := h.PageRepo.ListByFormID(form.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	questions, err := h.QuestionRepo.ListByFormID(form.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rules, err := h.RuleRepo.ListByFormID(form.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, PublicFormResponse{Form: *form, Pages: pages, Questions: questions, Rules: rules})
}

// SubmitResponse submits a response for a form.
// @Summary Submit form response
// @Tags public
// @Accept json
// @Produce json
// @Param slug path string true "Slug"
// @Param body body models.SubmitRequest true "Response data"
// @Success 201 {object} models.FormResponse
// @Failure 400,404,409 {object} map[string]string
// @Router /api/v1/public/forms/{slug}/submit [post]
func (h *Handler) SubmitResponse(c *gin.Context) {
	slug := c.Param("slug")
	form, err := h.FormRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	questions, err := h.QuestionRepo.ListByFormID(form.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var req models.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if form.SingleResponseOnly {
		if req.RespondentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "respondent_id required for single-response forms"})
			return
		}
		exists, err := h.ResponseRepo.ExistsByFormAndRespondent(c.Request.Context(), form.ID, req.RespondentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Already responded."})
			return
		}
	}
	if err := validator.ValidateSubmission(questions, req.Answers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	submittedAt := time.Now().UTC()
	var examDurationMs int64
	if form.ExamMode {
		if req.ExamStartedAt == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "exam_started_at is required when the form uses Exam Mode"})
			return
		}
		started := req.ExamStartedAt.UTC()
		if started.After(submittedAt.Add(examClockSkew)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam_started_at"})
			return
		}
		if started.Before(submittedAt.Add(-examMaxSessionHorizon)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "exam session expired; please refresh the form"})
			return
		}
		examDurationMs = submittedAt.Sub(started).Milliseconds()
		if examDurationMs < 0 {
			examDurationMs = 0
		}
	}

	answers := make([]models.AnswerEntry, len(req.Answers))
	byID := make(map[int64]models.Question)
	for _, q := range questions {
		byID[q.ID] = q
	}
	for i, a := range req.Answers {
		q := byID[a.QuestionID]
		answers[i] = models.AnswerEntry{
			QuestionID: a.QuestionID,
			Type:       q.Type,
			Value:      a.Value,
		}
		if m, ok := a.Value.(map[string]interface{}); ok {
			if fn, ok := m["filename"].(string); ok {
				answers[i].Filename = fn
			}
			if s, ok := m["size"].(float64); ok {
				answers[i].Size = int64(s)
			}
		}
	}
	resp := &models.FormResponse{
		FormID:         form.ID,
		Slug:           form.Slug,
		RespondentID:   req.RespondentID,
		SubmittedAt:    submittedAt,
		Answers:        answers,
		ExamDurationMs: examDurationMs,
	}
	if err := h.ResponseRepo.Insert(c.Request.Context(), resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}
