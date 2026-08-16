package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type usersPanelLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type usersPanelLoginResponse struct {
	Token string `json:"token"`
	User  struct {
		Email    string   `json:"email"`
		Username string   `json:"username"`
		Roles    []string `json:"roles"`
	} `json:"user"`
}

type usersPanelPermissionsResponse struct {
	Permissions []string `json:"permissions"`
}

func (h *Handler) Login(c *gin.Context) {
	var in usersPanelLoginRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	login, status, msg := h.usersPanelLogin(c.Request.Context(), in.Email, in.Password)
	if login == nil {
		if status == http.StatusUnauthorized && strings.EqualFold(strings.TrimSpace(msg), "unauthorized") {
			msg = "Invalid email or password"
		}
		c.JSON(status, gin.H{"error": msg})
		return
	}
	perms, _ := h.usersPanelPermissions(c.Request.Context(), login.Token)
	c.JSON(http.StatusOK, gin.H{
		"token":       login.Token,
		"user":        login.User,
		"permissions": perms,
	})
}

func (h *Handler) Me(c *gin.Context) {
	token := bearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
		return
	}
	user, err := h.usersPanelUser(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	perms, _ := h.usersPanelPermissions(c.Request.Context(), token)
	c.JSON(http.StatusOK, gin.H{
		"user":        user,
		"permissions": perms,
	})
}

// usersPanelLogin returns the session on success. On failure, status and msg are for JSON {error}; status 0 means success.
func (h *Handler) usersPanelLogin(ctx context.Context, email, password string) (*usersPanelLoginResponse, int, string) {
	reqBody := usersPanelLoginRequest{Email: strings.TrimSpace(email), Password: password}
	raw, _ := json.Marshal(reqBody)
	endpoint := strings.TrimRight(h.Cfg.UsersPanelBaseURL, "/") + "/api/auth/login"
	req, _ := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, "userspanel unavailable"
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		var out usersPanelLoginResponse
		if err := json.Unmarshal(body, &out); err != nil {
			return nil, http.StatusBadGateway, "invalid auth response"
		}
		if strings.TrimSpace(out.Token) == "" {
			return nil, http.StatusBadGateway, "missing token"
		}
		return &out, 0, ""
	}
	msg := "invalid credentials"
	var wrap struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &wrap) == nil && strings.TrimSpace(wrap.Error) != "" {
		msg = wrap.Error
	}
	switch resp.StatusCode {
	case http.StatusBadRequest:
		return nil, http.StatusBadRequest, msg
	case http.StatusUnauthorized:
		return nil, http.StatusUnauthorized, msg
	case http.StatusForbidden:
		return nil, http.StatusForbidden, msg
	default:
		return nil, http.StatusBadGateway, "userspanel login failed"
	}
}

func (h *Handler) usersPanelUser(ctx context.Context, token string) (map[string]any, error) {
	endpoint := strings.TrimRight(h.Cfg.UsersPanelBaseURL, "/") + "/api/auth/user"
	req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userspanel unavailable")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid session")
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("invalid user response")
	}
	return out, nil
}

func (h *Handler) usersPanelPermissions(ctx context.Context, token string) ([]string, error) {
	endpoint := strings.TrimRight(h.Cfg.UsersPanelBaseURL, "/") + "/api/auth/permissions"
	req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("permissions lookup failed")
	}
	var out usersPanelPermissionsResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out.Permissions, nil
}

func bearerToken(header string) string {
	h := strings.TrimSpace(header)
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

