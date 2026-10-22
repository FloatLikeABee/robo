package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Neo4jIngestStatus GET /api/admin/neo4j-ingest/status
func (h *Handlers) Neo4jIngestStatus(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sqlite not available"})
		return
	}
	st, err := h.TranMySQL.Neo4jIngestQueueStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"pending":     st.Pending,
		"processing":  st.Processing,
		"failed":      st.Failed,
		"done":        st.Done,
		"skipped":     st.Skipped,
		"queue_depth": st.Pending + st.Processing + st.Failed,
	})
}
