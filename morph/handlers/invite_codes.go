package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type redeemInviteBody struct {
	Code string `json:"code"`
}

// ListAdminInviteCodes GET /api/admin/invite-codes
func (h *Handlers) ListAdminInviteCodes(c *gin.Context) {
	if !h.requireAuthDB(c) || !h.requireAdmin(c) {
		return
	}
	list, err := h.TranMySQL.ListInviteCodes(c.Request.Context(), 500)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	codes := make([]map[string]any, 0, len(list))
	for i := range list {
		codes = append(codes, list[i].Public())
	}
	c.JSON(http.StatusOK, gin.H{"codes": codes, "total": len(codes)})
}

// CreateAdminInviteCode POST /api/admin/invite-codes
func (h *Handlers) CreateAdminInviteCode(c *gin.Context) {
	if !h.requireAuthDB(c) || !h.requireAdmin(c) {
		return
	}
	u, ok := h.userFromBearer(c)
	if !ok {
		return
	}
	inv, err := h.TranMySQL.CreateInviteCode(c.Request.Context(), u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": inv.Public()})
}

// RedeemInviteCode POST /api/invite/redeem
func (h *Handlers) RedeemInviteCode(c *gin.Context) {
	if !h.requireAuthDB(c) {
		return
	}
	var body redeemInviteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	username, password, err := h.TranMySQL.RedeemInviteCode(c.Request.Context(), body.Code)
	if err != nil {
		msg := err.Error()
		st := http.StatusBadRequest
		if strings.Contains(msg, "invalid") {
			st = http.StatusNotFound
		} else if strings.Contains(msg, "already used") {
			st = http.StatusGone
		}
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": username, "password": password})
}
