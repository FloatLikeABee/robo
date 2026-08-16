package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/formsx/backend/internal/models"
	"github.com/gin-gonic/gin"
)

type mongoMCPCallRequest struct {
	Tool string                 `json:"tool"`
	Args map[string]interface{} `json:"args"`
}

// FormsXMongoMCPTools describes Badger-backed document tools for MorphAI and other MCP-style clients.
func (h *Handler) FormsXMongoMCPTools(c *gin.Context) {
	base := "/api/v1/ai/mongodb-mcp"
	c.JSON(http.StatusOK, gin.H{
		"service":     "formsx-mongodb-mcp",
		"description": "Document-store tool gateway for FormsX system details and Events & Info (Badger-backed).",
		"collections": []string{"ai_system_documents", "workspace_events", "form_responses"},
		"call_endpoint": gin.H{
			"method": "POST",
			"path":   base + "/call",
		},
		"tools": []gin.H{
			{
				"name":        "sync_system_documents",
				"description": "Sync FormsX system details (forms) into ai_system_documents.",
				"args":        gin.H{},
			},
			{
				"name":        "list_system_documents",
				"description": "List synced system documents. Use doc_type/search for fast form/system lookups.",
				"args":        gin.H{"doc_type": "form (optional)", "search": "optional text", "page": 1, "limit": 50},
			},
			{
				"name":        "search_system_documents",
				"description": "Search synced system documents by keyword. Prefer this before broad form or system questions.",
				"args":        gin.H{"query": "required text", "doc_type": "form (optional)", "page": 1, "limit": 25},
			},
			{
				"name":        "get_system_document",
				"description": "Get a synced system document by id.",
				"args":        gin.H{"id": "document id"},
			},
			{
				"name":        "list_events_info",
				"description": "List Events & Info from workspace_events.",
				"args":        gin.H{"page": 1, "limit": 50},
			},
			{
				"name":        "get_event_info",
				"description": "Get one Events & Info record by id.",
				"args":        gin.H{"event_id": "event id"},
			},
			{
				"name":        "list_form_responses",
				"description": "List form responses for a given form id.",
				"args":        gin.H{"form_id": 1, "page": 1, "limit": 50},
			},
			{
				"name":        "get_form_response",
				"description": "Get one form response by form id + response id.",
				"args":        gin.H{"form_id": 1, "response_id": "response id"},
			},
		},
	})
}

// FormsXMongoMCPCall executes MCP-style tool calls against Mongo-backed data.
func (h *Handler) FormsXMongoMCPCall(c *gin.Context) {
	var req mongoMCPCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Tool = strings.TrimSpace(req.Tool)
	if req.Args == nil {
		req.Args = map[string]interface{}{}
	}
	payload, err := h.execMongoMCPTool(c.Request.Context(), req.Tool, req.Args)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "tool": req.Tool})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tool": req.Tool, "result": payload})
}

func (h *Handler) execMongoMCPTool(ctx context.Context, tool string, args map[string]interface{}) (interface{}, error) {
	if h.AIDocRepo == nil {
		return nil, fmt.Errorf("ai document repo is not configured")
	}
	switch tool {
	case "sync_system_documents":
		return h.syncSystemDocuments(ctx)
	case "list_system_documents":
		page := intArg(args, "page", 1)
		limit := intArg(args, "limit", 50)
		docType := stringArg(args, "doc_type")
		search := stringArg(args, "search")
		list, total, err := h.AIDocRepo.List(ctx, docType, search, page, limit)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, 0, len(list))
		for i := range list {
			out = append(out, models.AIDocumentToMap(&list[i]))
		}
		return gin.H{"documents": out, "total": total, "page": page, "limit": limit}, nil
	case "search_system_documents":
		query := stringArg(args, "query")
		if query == "" {
			query = stringArg(args, "search")
		}
		if query == "" {
			return nil, fmt.Errorf("search_system_documents requires query")
		}
		page := intArg(args, "page", 1)
		limit := intArg(args, "limit", 25)
		if limit > 50 {
			limit = 50
		}
		docType := stringArg(args, "doc_type")
		list, total, err := h.AIDocRepo.List(ctx, docType, query, page, limit)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, 0, len(list))
		for i := range list {
			doc := &list[i]
			out = append(out, gin.H{
				"id":       doc.ID,
				"doc_type": doc.DocType,
				"title":    doc.Title,
				"summary":  doc.Summary,
				"tags":     doc.Tags,
			})
		}
		return gin.H{"documents": out, "total": total, "page": page, "limit": limit, "query": query}, nil
	case "get_system_document":
		id := stringArg(args, "id")
		if id == "" {
			return nil, fmt.Errorf("get_system_document requires id")
		}
		doc, err := h.AIDocRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		return models.AIDocumentToMap(doc), nil
	case "list_events_info":
		page := intArg(args, "page", 1)
		limit := intArg(args, "limit", 50)
		list, total, err := h.EventInfoRepo.List(ctx, page, limit)
		if err != nil {
			return nil, err
		}
		out := make([]*models.EventInfoResponse, 0, len(list))
		for i := range list {
			out = append(out, models.EventInfoToResponse(&list[i]))
		}
		return gin.H{"events": out, "total": total, "page": page, "limit": limit}, nil
	case "get_event_info":
		idHex := stringArg(args, "event_id")
		if idHex == "" {
			return nil, fmt.Errorf("get_event_info requires event_id")
		}
		ev, err := h.EventInfoRepo.GetByID(ctx, idHex)
		if err != nil {
			return nil, err
		}
		return models.EventInfoToResponse(ev), nil
	case "list_form_responses":
		formID := int64Arg(args, "form_id")
		if formID <= 0 {
			return nil, fmt.Errorf("list_form_responses requires form_id")
		}
		page := intArg(args, "page", 1)
		limit := intArg(args, "limit", 50)
		list, total, err := h.ResponseRepo.List(ctx, formID, page, limit, nil, nil)
		if err != nil {
			return nil, err
		}
		return gin.H{"responses": list, "total": total, "page": page, "limit": limit, "form_id": formID}, nil
	case "get_form_response":
		formID := int64Arg(args, "form_id")
		responseID := stringArg(args, "response_id")
		if formID <= 0 || responseID == "" {
			return nil, fmt.Errorf("get_form_response requires form_id and response_id")
		}
		resp, err := h.ResponseRepo.GetByID(ctx, formID, responseID)
		if err != nil {
			return nil, err
		}
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown tool %q", tool)
	}
}

func (h *Handler) syncSystemDocuments(ctx context.Context) (gin.H, error) {
	if h.AIDocRepo == nil {
		return nil, fmt.Errorf("ai document repo is not configured")
	}
	start := time.Now()

	forms, err := h.fetchAllForms()
	if err != nil {
		return nil, err
	}

	var upsertedForms int
	for i := range forms {
		form := forms[i]
		qs, err := h.QuestionRepo.ListByFormID(form.ID)
		if err != nil {
			return nil, err
		}
		doc := &models.AIDocument{
			DocType:  "form",
			SourceID: strconv.FormatInt(form.ID, 10),
			Title:    form.Name,
			Summary:  form.Description,
			Tags:     []string{"formsx", "form", form.Slug},
			Data: map[string]interface{}{
				"form":      form,
				"questions": qs,
			},
		}
		if err := h.AIDocRepo.Upsert(ctx, doc); err != nil {
			return nil, err
		}
		upsertedForms++
	}

	return gin.H{
		"upserted": gin.H{
			"forms": upsertedForms,
			"total": upsertedForms,
		},
		"duration_ms": time.Since(start).Milliseconds(),
		"collection":  "ai_system_documents",
	}, nil
}

func (h *Handler) fetchAllForms() ([]models.Form, error) {
	page := 1
	limit := 100
	all := make([]models.Form, 0)
	for {
		list, total, err := h.FormRepo.List(page, limit, "")
		if err != nil {
			return nil, err
		}
		all = append(all, list...)
		if int64(len(all)) >= total || len(list) == 0 {
			break
		}
		page++
	}
	return all, nil
}
