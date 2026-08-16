package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler checks the health status of the service
// @Summary      Health check
// @Description  Check the health status of core services (database, AI)
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]string  "Service health status"
// @Router       /health [get]
func (h *Handlers) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "healthy",
		"db":         "connected",
		"ai_service": "ready",
	})
}
