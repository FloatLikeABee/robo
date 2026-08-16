package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MessagesStub answers legacy messaging routes with Gone — inbox UI has been removed.
func (h *Handlers) MessagesStub(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{
		"error":   "messaging_removed",
		"message": "In-app messaging has been removed.",
		"threads": []any{},
		"messages": []any{},
		"users":   []any{},
		"count":   0,
		"unread":  0,
	})
}
