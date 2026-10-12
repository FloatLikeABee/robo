package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"idongivaflyinfa/ai"
	"idongivaflyinfa/grill"

	"github.com/gin-gonic/gin"
)

// TranGrill runs one step of Notes & TODOs AI assist for the signed-in user.
// POST /api/tran/grill
func (h *Handlers) TranGrill(c *gin.Context) {
	owner, ok := grillOwner(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	var in struct {
		Action      string `json:"action"`
		ID          string `json:"id"`
		Kind        string `json:"kind"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Answer      string `json:"answer"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	view, err := h.grillStore().Act(c.Request.Context(), owner, grill.Input{
		Action:      in.Action,
		ID:          in.ID,
		Kind:        in.Kind,
		Title:       in.Title,
		Description: in.Description,
		Answer:      in.Answer,
	}, h.grillCompleter())
	if err != nil {
		c.JSON(grillHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, view)
}

// TranGrillUI returns the grill HTML document the modal iframes.
// GET /api/tran/grill/ui
func (h *Handlers) TranGrillUI(c *gin.Context) {
	if _, ok := grillOwner(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(grill.AppHTML))
}

func grillOwner(c *gin.Context) (string, bool) {
	raw, ok := c.Get("auth_user_id")
	owner, _ := raw.(string)
	owner = strings.TrimSpace(owner)
	if !ok || owner == "" {
		return "", false
	}
	return owner, true
}

func (h *Handlers) grillStore() *grill.Store {
	h.grillMu.Lock()
	defer h.grillMu.Unlock()
	if h.grillSessions == nil {
		h.grillSessions = grill.NewStore()
	}
	return h.grillSessions
}

func (h *Handlers) grillCompleter() grill.Completer {
	if h != nil && h.grillComplete != nil {
		return h.grillComplete
	}
	return func(ctx context.Context, prompt string) (string, error) {
		if h == nil || h.aiService == nil {
			return "", errors.New("AI service not configured")
		}
		return h.aiService.ChatCompletion(ctx, []ai.DashScopeMessage{{Role: "user", Content: prompt}})
	}
}

func grillHTTPStatus(err error) int {
	switch {
	case errors.Is(err, grill.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, grill.ErrTitleRequired),
		errors.Is(err, grill.ErrAnswerRequired),
		errors.Is(err, grill.ErrKind),
		errors.Is(err, grill.ErrAction),
		errors.Is(err, grill.ErrTooLong):
		return http.StatusBadRequest
	case err != nil && strings.Contains(err.Error(), "not configured"):
		return http.StatusServiceUnavailable
	default:
		return http.StatusBadGateway
	}
}
