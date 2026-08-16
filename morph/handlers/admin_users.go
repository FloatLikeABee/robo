package handlers

import (
	"database/sql"
	"net/http"
	"net/mail"
	"strings"

	"github.com/gin-gonic/gin"
)

type createUserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

type updateUserBody struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
	IsAdmin  *bool   `json:"is_admin"`
}

// ListAdminUsers GET /api/admin/users
func (h *Handlers) ListAdminUsers(c *gin.Context) {
	if !h.requireAuthDB(c) || !h.requireAdmin(c) {
		return
	}
	list, err := h.TranMySQL.ListPlatUsers(c.Request.Context(), 1000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	users := make([]map[string]any, 0, len(list))
	for i := range list {
		users = append(users, list[i].Public())
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "total": len(users)})
}

// CreateAdminUser POST /api/admin/users
func (h *Handlers) CreateAdminUser(c *gin.Context) {
	if !h.requireAuthDB(c) || !h.requireAdmin(c) {
		return
	}
	var body createUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	email := strings.ToLower(strings.TrimSpace(body.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email must be a valid email address"})
		return
	}
	u, err := h.TranMySQL.CreatePlatUser(c.Request.Context(), email, body.Password, body.IsAdmin)
	if err != nil {
		msg := err.Error()
		st := http.StatusBadRequest
		if strings.Contains(msg, "already") {
			st = http.StatusConflict
		}
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "user created", "user": u.Public(), "userId": u.ID})
}

// UpdateAdminUser PATCH /api/admin/users/:id
func (h *Handlers) UpdateAdminUser(c *gin.Context) {
	if !h.requireAuthDB(c) || !h.requireAdmin(c) {
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	var body updateUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	cur, err := h.TranMySQL.GetPlatUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if body.IsAdmin != nil && !*body.IsAdmin && cur.IsAdmin() {
		n, _ := h.TranMySQL.CountPlatAdmins(c.Request.Context())
		if n <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot demote the last admin"})
			return
		}
	}
	email := ""
	if body.Email != nil {
		email = strings.ToLower(strings.TrimSpace(*body.Email))
		if email != "" {
			if _, err := mail.ParseAddress(email); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "email must be a valid email address"})
				return
			}
		}
	}
	password := ""
	if body.Password != nil {
		password = *body.Password
	}
	u, err := h.TranMySQL.UpdatePlatUser(c.Request.Context(), id, email, password, body.IsAdmin)
	if err != nil {
		msg := err.Error()
		st := http.StatusBadRequest
		if strings.Contains(msg, "already") {
			st = http.StatusConflict
		}
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user updated", "user": u.Public()})
}

// DeleteAdminUser DELETE /api/admin/users/:id
func (h *Handlers) DeleteAdminUser(c *gin.Context) {
	if !h.requireAuthDB(c) || !h.requireAdmin(c) {
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	cur, err := h.TranMySQL.GetPlatUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	selfID, _ := c.Get("auth_user_id")
	if sid, ok := selfID.(string); ok && sid == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete your own account"})
		return
	}
	if cur.IsAdmin() {
		n, _ := h.TranMySQL.CountPlatAdmins(c.Request.Context())
		if n <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete the last admin"})
			return
		}
	}
	if err := h.TranMySQL.DeletePlatUser(c.Request.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
