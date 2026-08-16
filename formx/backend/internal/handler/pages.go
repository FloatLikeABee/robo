package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/formsx/backend/internal/models"
	"github.com/gin-gonic/gin"
)

func normalizePageName(name string) string {
	return strings.TrimSpace(name)
}

// ListPages returns pages for a form, ensuring at least one default page exists.
func (h *Handler) ListPages(c *gin.Context) {
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
	list, err := h.PageRepo.ListByFormID(formID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// CreatePage adds a page to a form.
func (h *Handler) CreatePage(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	if _, err := h.FormRepo.GetByID(formID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	var req models.CreatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing, _ := h.PageRepo.ListByFormID(formID)
	sortOrder := req.SortOrder
	if sortOrder == 0 && len(existing) > 0 {
		sortOrder = len(existing)
	}
	page := &models.FormPage{
		FormID:    formID,
		Name:      normalizePageName(req.Name),
		SortOrder: sortOrder,
	}
	if err := h.PageRepo.Create(page); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, page)
}

// UpdatePage updates a page.
func (h *Handler) UpdatePage(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	pageID, err := strconv.ParseInt(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page id"})
		return
	}
	page, err := h.PageRepo.GetByFormIDAndID(formID, pageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}
	var req models.UpdatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != nil {
		page.Name = normalizePageName(*req.Name)
	}
	if req.SortOrder != nil {
		page.SortOrder = *req.SortOrder
	}
	if err := h.PageRepo.Update(page); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, page)
}

// DeletePage deletes a page. The last page cannot be deleted; questions move to another page.
func (h *Handler) DeletePage(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	pageID, err := strconv.ParseInt(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page id"})
		return
	}
	if _, err := h.PageRepo.GetByFormIDAndID(formID, pageID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}
	count, err := h.PageRepo.CountByFormID(formID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if count <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete the only page"})
		return
	}
	pages, err := h.PageRepo.ListByFormID(formID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var targetPageID int64
	for _, p := range pages {
		if p.ID != pageID {
			targetPageID = p.ID
			break
		}
	}
	if targetPageID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete the only page"})
		return
	}
	if err := h.QuestionRepo.ReassignPage(formID, pageID, targetPageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.PageRepo.Delete(pageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
