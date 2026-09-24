package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health reports process liveness. It does not call Morph, MorphUtils, or a sibling embed.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
