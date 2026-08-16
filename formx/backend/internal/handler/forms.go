package handler

import (
	"net/http"
	"strconv"

	"github.com/formsx/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// CreateForm creates a new form.
// @Summary Create form
// @Tags forms
// @Accept json
// @Produce json
// @Param body body models.CreateFormRequest true "Form data"
// @Success 201 {object} models.Form
// @Failure 400,409 {object} map[string]string
// @Router /api/v1/forms [post]
func (h *Handler) CreateForm(c *gin.Context) {
	var req models.CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	exists, err := h.FormRepo.ExistsSlug(req.Slug, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "slug already in use"})
		return
	}
	form := &models.Form{
		Name:               req.Name,
		Description:        req.Description,
		Slug:               req.Slug,
		SingleResponseOnly: req.SingleResponseOnly,
		ExamMode:           req.ExamMode,
		LandingHTML:        req.LandingHTML,
	}
	if err := h.FormRepo.Create(form); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.FormRepo.EnqueueFormUpsert(c.Request.Context(), form.ID)
	if _, err := h.PageRepo.EnsureDefaultPage(form.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, form)
}

// ListForms returns paginated forms.
// @Summary List forms
// @Tags forms
// @Produce json
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(20)
// @Param search query string false "Search"
// @Success 200 {object} models.FormListResponse
// @Router /api/v1/forms [get]
func (h *Handler) ListForms(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	list, total, err := h.FormRepo.List(page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.FormListResponse{Forms: list, Total: total, Page: page, Limit: limit})
}

// GetForm returns a form by ID.
// @Summary Get form by ID
// @Tags forms
// @Produce json
// @Param id path int true "Form ID"
// @Success 200 {object} models.Form
// @Failure 404 {object} map[string]string
// @Router /api/v1/forms/{id} [get]
func (h *Handler) GetForm(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	form, err := h.FormRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	c.JSON(http.StatusOK, form)
}

// GetFormBySlug returns a form by slug.
// @Summary Get form by slug
// @Tags forms
// @Produce json
// @Param slug path string true "Slug"
// @Success 200 {object} models.Form
// @Failure 404 {object} map[string]string
// @Router /api/v1/forms/by-slug/{slug} [get]
func (h *Handler) GetFormBySlug(c *gin.Context) {
	slug := c.Param("slug")
	form, err := h.FormRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	c.JSON(http.StatusOK, form)
}

// UpdateForm updates a form.
// @Summary Update form
// @Tags forms
// @Accept json
// @Produce json
// @Param id path int true "Form ID"
// @Param body body models.UpdateFormRequest true "Updates"
// @Success 200 {object} models.Form
// @Failure 400,404,409 {object} map[string]string
// @Router /api/v1/forms/{id} [put]
func (h *Handler) UpdateForm(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	form, err := h.FormRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	var req models.UpdateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != nil {
		form.Name = *req.Name
	}
	if req.Description != nil {
		form.Description = *req.Description
	}
	if req.SingleResponseOnly != nil {
		form.SingleResponseOnly = *req.SingleResponseOnly
	}
	if req.ExamMode != nil {
		form.ExamMode = *req.ExamMode
	}
	if req.LandingHTML != nil {
		form.LandingHTML = *req.LandingHTML
	}
	if req.Slug != nil && *req.Slug != form.Slug {
		exists, _ := h.FormRepo.ExistsSlug(*req.Slug, id)
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "slug already in use"})
			return
		}
		form.Slug = *req.Slug
	}
	if err := h.FormRepo.Update(form); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.FormRepo.EnqueueFormUpsert(c.Request.Context(), form.ID)
	c.JSON(http.StatusOK, form)
}

// DeleteForm deletes a form.
// @Summary Delete form
// @Tags forms
// @Param id path int true "Form ID"
// @Success 204
// @Failure 400,404 {object} map[string]string
// @Router /api/v1/forms/{id} [delete]
func (h *Handler) DeleteForm(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.FormRepo.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	h.FormRepo.EnqueueFormDelete(c.Request.Context(), id)
	c.Status(http.StatusNoContent)
}
