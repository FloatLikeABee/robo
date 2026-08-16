package handler

import (
	"net/http"
	"strconv"

	"github.com/formsx/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// CreateQuestion adds a question to a form.
// @Summary Create question
// @Tags questions
// @Accept json
// @Produce json
// @Param id path int true "Form ID"
// @Param body body models.CreateQuestionRequest true "Question data"
// @Success 201 {object} models.Question
// @Failure 400,404 {object} map[string]string
// @Router /api/v1/forms/{formId}/questions [post]
func (h *Handler) CreateQuestion(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	if _, err := h.FormRepo.GetByID(formID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	var req models.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	title, err := normalizeQuestionTitle(req.Title)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !models.ValidQuestionTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question type"})
		return
	}
	pageID := req.PageID
	if pageID == 0 {
		if _, err := h.PageRepo.EnsureDefaultPage(formID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		pages, err := h.PageRepo.ListByFormID(formID)
		if err != nil || len(pages) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no pages available"})
			return
		}
		pageID = pages[0].ID
	} else if _, err := h.PageRepo.GetByFormIDAndID(formID, pageID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_id"})
		return
	}
	q := &models.Question{
		FormID:    formID,
		PageID:    pageID,
		Title:     title,
		Type:      req.Type,
		Required:  req.Required,
		SortOrder: req.SortOrder,
		Config:    req.Config,
	}
	if err := h.QuestionRepo.Create(q); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, q)
}

// ListQuestions returns questions for a form.
// @Summary List questions
// @Tags questions
// @Produce json
// @Param id path int true "Form ID"
// @Success 200 {array} models.Question
// @Failure 400,404 {object} map[string]string
// @Router /api/v1/forms/{formId}/questions [get]
func (h *Handler) ListQuestions(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	if _, err := h.FormRepo.GetByID(formID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	if _, err := h.PageRepo.EnsureDefaultPage(formID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	list, err := h.QuestionRepo.ListByFormID(formID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// UpdateQuestion updates a question.
// @Summary Update question
// @Tags questions
// @Accept json
// @Produce json
// @Param id path int true "Form ID"
// @Param questionId path int true "Question ID"
// @Param body body models.UpdateQuestionRequest true "Updates"
// @Success 200 {object} models.Question
// @Failure 400,404 {object} map[string]string
// @Router /api/v1/forms/{formId}/questions/{id} [put]
func (h *Handler) UpdateQuestion(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	qID, err := strconv.ParseInt(c.Param("questionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}
	q, err := h.QuestionRepo.GetByFormIDAndID(formID, qID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}
	var req models.UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Title != nil {
		title, err := normalizeQuestionTitle(*req.Title)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		q.Title = title
	}
	if req.Type != nil {
		if !models.ValidQuestionTypes[*req.Type] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question type"})
			return
		}
		q.Type = *req.Type
	}
	if req.Required != nil {
		q.Required = *req.Required
	}
	if req.PageID != nil {
		if *req.PageID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_id"})
			return
		}
		if _, err := h.PageRepo.GetByFormIDAndID(formID, *req.PageID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_id"})
			return
		}
		q.PageID = *req.PageID
	}
	if req.SortOrder != nil {
		q.SortOrder = *req.SortOrder
	}
	if req.Config != nil {
		q.Config = *req.Config
	}
	if err := h.QuestionRepo.Update(q); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

// DeleteQuestion deletes a question.
// @Summary Delete question
// @Tags questions
// @Param id path int true "Form ID"
// @Param questionId path int true "Question ID"
// @Success 204
// @Failure 400,404 {object} map[string]string
// @Router /api/v1/forms/{formId}/questions/{id} [delete]
func (h *Handler) DeleteQuestion(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	qID, err := strconv.ParseInt(c.Param("questionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}
	q, err := h.QuestionRepo.GetByFormIDAndID(formID, qID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}
	if q.Config.QuestionPromptMedia != nil {
		deletePromptMediaOnDisk(h.Cfg.UploadDir, q.Config.QuestionPromptMedia.RelativePath)
	}
	if err := h.QuestionRepo.Delete(qID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
