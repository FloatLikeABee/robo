package handler

import (
	"net/http"
	"strconv"

	"github.com/formsx/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// ListRulesByFormID returns all rules for questions in the form (for admin UI).
// @Summary List all rules for a form
// @Tags rules
// @Produce json
// @Param id path int true "Form ID"
// @Success 200 {array} models.QuestionRule
// @Failure 400,404 {object} map[string]string
// @Router /api/v1/forms/{id}/rules [get]
func (h *Handler) ListRulesByFormID(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	if _, err := h.FormRepo.GetByID(formID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	list, err := h.RuleRepo.ListByFormID(formID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListRules returns rules for a question.
// @Summary List rules for a question
// @Tags rules
// @Produce json
// @Param id path int true "Form ID"
// @Param questionId path int true "Question ID"
// @Success 200 {array} models.QuestionRule
// @Failure 400,404 {object} map[string]string
// @Router /api/v1/forms/{id}/questions/{questionId}/rules [get]
func (h *Handler) ListRules(c *gin.Context) {
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
	if _, err := h.QuestionRepo.GetByFormIDAndID(formID, qID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}
	list, err := h.RuleRepo.ListByQuestionID(qID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// CreateRule adds a visibility rule to a question.
// @Summary Create rule for a question
// @Tags rules
// @Accept json
// @Produce json
// @Param id path int true "Form ID"
// @Param questionId path int true "Question ID"
// @Param body body models.CreateRuleRequest true "Rule data"
// @Success 201 {object} models.QuestionRule
// @Failure 400,404,409 {object} map[string]string
// @Router /api/v1/forms/{id}/questions/{questionId}/rules [post]
func (h *Handler) CreateRule(c *gin.Context) {
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
	if _, err := h.QuestionRepo.GetByFormIDAndID(formID, qID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}
	var req models.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !models.ValidRuleConditions[req.Condition] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid condition: use 'answered' or 'not_answered'"})
		return
	}
	if req.DependsOnQuestionID == qID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question cannot depend on itself"})
		return
	}
	if _, err := h.QuestionRepo.GetByFormIDAndID(formID, req.DependsOnQuestionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "depends_on_question_id must be another question in the same form"})
		return
	}
	exists, err := h.RuleRepo.ExistsForQuestionAndDepends(qID, req.DependsOnQuestionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "a rule for this question and dependency already exists; rules must not negate each other"})
		return
	}
	rule := &models.QuestionRule{
		QuestionID:          qID,
		DependsOnQuestionID: req.DependsOnQuestionID,
		Condition:           req.Condition,
	}
	if err := h.RuleRepo.Create(rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

// DeleteRule removes a rule.
// @Summary Delete rule
// @Tags rules
// @Param id path int true "Form ID"
// @Param questionId path int true "Question ID"
// @Param ruleId path int true "Rule ID"
// @Success 204
// @Failure 400,404 {object} map[string]string
// @Router /api/v1/forms/{id}/questions/{questionId}/rules/{ruleId} [delete]
func (h *Handler) DeleteRule(c *gin.Context) {
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
	ruleID, err := strconv.ParseInt(c.Param("ruleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}
	if _, err := h.QuestionRepo.GetByFormIDAndID(formID, qID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}
	rule, err := h.RuleRepo.GetByID(ruleID)
	if err != nil || rule.QuestionID != qID {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	if err := h.RuleRepo.Delete(ruleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
