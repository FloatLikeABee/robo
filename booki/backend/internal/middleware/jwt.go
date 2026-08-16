package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"uid"`
	OrgID  int64  `json:"org"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		raw := strings.TrimPrefix(h, "Bearer ")
		tok, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		cl, ok := tok.Claims.(*Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}
		c.Set("userID", cl.UserID)
		c.Set("orgID", cl.OrgID)
		c.Set("role", cl.Role)
		c.Next()
	}
}

func GetOrgID(c *gin.Context) int64 {
	v, _ := c.Get("orgID")
	id, _ := v.(int64)
	return id
}

func GetUserID(c *gin.Context) int64 {
	v, _ := c.Get("userID")
	id, _ := v.(int64)
	return id
}

func RequireBookiLedgerAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := strings.ToLower(strings.TrimSpace(roleFromContext(c)))
		if role == "admin" || role == "owner" {
			c.Next()
			return
		}
		if role == "employee" && hasPermissionCSV(c.GetHeader("X-User-Permissions"), "morph_booki") {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "booki access restricted by admin policy"})
	}
}

func roleFromContext(c *gin.Context) string {
	if v, ok := c.Get("role"); ok {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	if h := strings.TrimSpace(c.GetHeader("X-User-Role")); h != "" {
		return h
	}
	out := ""
	for _, raw := range strings.Split(c.GetHeader("X-User-Roles"), ",") {
		r := strings.ToLower(strings.TrimSpace(raw))
		switch r {
		case "admin", "owner":
			return "admin"
		case "employee":
			out = "employee"
		case "member":
			if out == "" {
				out = "member"
			}
		case "forms", "email composer", "main panel", "sharp reports", "booki ledger":
			if out != "admin" {
				out = "employee"
			}
		}
	}
	return out
}

func hasPermissionCSV(rawCSV, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, raw := range strings.Split(rawCSV, ",") {
		if strings.ToLower(strings.TrimSpace(raw)) == target {
			return true
		}
	}
	return false
}
