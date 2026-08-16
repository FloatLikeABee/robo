package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListMorphAIAgentsHandler GET /api/ai-agents — AI tools (BK) assistants only for Morph AI chat.
// @Summary      List Morph AI agents
// @Tags         Chat
// @Produce      json
// @Param        enabled  query     string  false  "Ignored; BK assistants are listed when available"
// @Success      200      {object}  map[string]interface{}
// @Router       /api/ai-agents [get]
func (h *Handlers) ListMorphAIAgentsHandler(c *gin.Context) {
	out := make([]gin.H, 0, 8)

	if bkList, bkErr := h.listBKAssistants(c.Request.Context()); bkErr != nil {
		log.Printf("[AI-AGENTS] BK assistants unavailable: %v", bkErr)
	} else {
		sortBase := 1000
		for i, a := range bkList {
			id := strings.TrimSpace(a.ID)
			if id == "" {
				continue
			}
			desc := strings.TrimSpace(a.Description)
			if desc == "" && len(a.RAGCollections) > 0 {
				desc = "AI tools assistant · RAG: " + strings.Join(a.RAGCollections, ", ")
			} else if desc == "" {
				desc = "AI tools assistant"
			}
			out = append(out, gin.H{
				"id":              morphAgentIDForBK(id),
				"name":            strings.TrimSpace(a.Name),
				"description":     desc,
				"system_defined":  false,
				"sort_order":      sortBase + i,
				"enabled":         true,
				"source":          "bk",
				"rag_collections": a.RAGCollections,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"agents": out, "total": len(out)})
}
