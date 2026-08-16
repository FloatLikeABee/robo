package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/formsx/backend/internal/models"
	"github.com/formsx/backend/internal/mongo"
	"github.com/gin-gonic/gin"
)

const (
	eventInfoMaxTitleLen    = 500
	eventInfoMaxDetailLen   = 20_000
	eventInfoMaxReporterLen = 200
)

func insertEventFromRequest(req models.CreateEventInfoRequest) (*models.EventInfo, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, errors.New("title is required")
	}
	if len(title) > eventInfoMaxTitleLen {
		return nil, errors.New("title is too long")
	}
	detail := strings.TrimSpace(req.Detail)
	if len(detail) > eventInfoMaxDetailLen {
		return nil, errors.New("detail is too long")
	}
	reporter := strings.TrimSpace(req.Reporter)
	if len(reporter) > eventInfoMaxReporterLen {
		return nil, errors.New("reporter is too long")
	}
	t, err := time.Parse(time.RFC3339, req.Time)
	if err != nil {
		return nil, errors.New("time must be RFC3339 / ISO-8601")
	}
	return &models.EventInfo{
		Title:     title,
		Detail:    detail,
		Reporter:  reporter,
		EventTime: t.UTC(),
	}, nil
}

// ListEventInfo returns paginated Events & Info (MongoDB only).
func (h *Handler) ListEventInfo(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	ctx := context.Background()
	list, total, err := h.EventInfoRepo.List(ctx, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]*models.EventInfoResponse, 0, len(list))
	for i := range list {
		out = append(out, models.EventInfoToResponse(&list[i]))
	}
	c.JSON(http.StatusOK, gin.H{"events": out, "total": total, "page": page, "limit": limit})
}

// CreateEventInfo stores a new event in MongoDB.
func (h *Handler) CreateEventInfo(c *gin.Context) {
	var req models.CreateEventInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ev, err := insertEventFromRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.EventInfoRepo.Insert(context.Background(), ev); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, models.EventInfoToResponse(ev))
}

// PublicCreateEventInfo stores a new event without authentication (public collection).
func (h *Handler) PublicCreateEventInfo(c *gin.Context) {
	var req models.CreateEventInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ev, err := insertEventFromRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.EventInfoRepo.Insert(context.Background(), ev); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, models.EventInfoToResponse(ev))
}

// GetEventInfo returns one event by id.
func (h *Handler) GetEventInfo(c *gin.Context) {
	idHex := c.Param("id")
	if idHex == "collection-info" {
		h.GetEventInfoCollectionInfo(c)
		return
	}
	if strings.TrimSpace(idHex) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ev, err := h.EventInfoRepo.GetByID(context.Background(), idHex)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.EventInfoToResponse(ev))
}

// GetEventInfoAIContext returns the event for Morph / assistants.
func (h *Handler) GetEventInfoAIContext(c *gin.Context) {
	idHex := c.Param("id")
	if strings.TrimSpace(idHex) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ev, err := h.EventInfoRepo.GetByID(context.Background(), idHex)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.EventInfoAIContextResponse{
		Event: models.EventInfoToResponse(ev),
	})
}

// DeleteEventInfo removes an event from Badger storage.
func (h *Handler) DeleteEventInfo(c *gin.Context) {
	idHex := c.Param("id")
	if strings.TrimSpace(idHex) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.EventInfoRepo.Delete(context.Background(), idHex); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// FormsXMorphMiniMCP returns a small tool catalog so Morph AI and other clients can call FormsX Events APIs.
func (h *Handler) FormsXMorphMiniMCP(c *gin.Context) {
	base := "/api/v1"
	c.JSON(http.StatusOK, gin.H{
		"service":     "formsx-events-info",
		"description": "FormsX Events & Info (embedded Badger store). Use these HTTP tools (same auth as FormsX) to fetch operational notes.",
		"tools": []gin.H{
			{
				"name":          "formsx_list_events",
				"description":   "List recent events & info (title, reporter, time, id).",
				"method":        "GET",
				"path":          base + "/events-info",
				"query_example": "page=1&limit=50",
			},
			{
				"name":        "formsx_get_event",
				"description": "Get one event with full detail.",
				"method":      "GET",
				"path":        base + "/events-info/{id}",
			},
			{
				"name":        "formsx_get_event_ai_context",
				"description": "Get one event with full detail for AI assistants.",
				"method":      "GET",
				"path":        base + "/events-info/{id}/ai-context",
			},
			{
				"name":        "formsx_create_event",
				"description": "Create an event. Body JSON: title, detail, reporter, time (RFC3339).",
				"method":      "POST",
				"path":        base + "/events-info",
				"body_example": gin.H{
					"title":    "Quarterly review",
					"detail":   "Discussed roadmap with team.",
					"reporter": "Jane Smith",
					"time":     time.Now().UTC().Format(time.RFC3339),
				},
			},
			{
				"name":        "formsx_public_create_event",
				"description": "Create an event without auth (public collection). Same JSON body as formsx_create_event.",
				"method":      "POST",
				"path":        base + "/public/events-info",
				"body_example": gin.H{
					"title":    "Quarterly review",
					"detail":   "Discussed roadmap with team.",
					"reporter": "Jane Smith",
					"time":     time.Now().UTC().Format(time.RFC3339),
				},
			},
			{
				"name":        "formsx_delete_event",
				"description": "Delete an event by id.",
				"method":      "DELETE",
				"path":        base + "/events-info/{id}",
			},
			{
				"name":        "formsx_mongodb_mcp",
				"description": "Document store MCP tool gateway for full system documents + events + form responses.",
				"method":      "GET",
				"path":        base + "/ai/mongodb-mcp",
			},
			{
				"name":        "formsx_app_abilities",
				"description": "Web-grounded form template generation and related Morph AI app abilities.",
				"method":      "GET",
				"path":        base + "/ai/app-abilities",
			},
		},
	})
}
