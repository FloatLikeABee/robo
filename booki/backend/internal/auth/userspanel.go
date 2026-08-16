package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

type platformSession struct {
	Token       string
	Email       string
	Username    string
	Roles       []string
	Permissions []string
}

func usersPanelLogin(ctx context.Context, baseURL, email, password string) (*platformSession, int, string) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, http.StatusBadGateway, "userspanel unavailable"
	}
	reqBody := usersPanelLoginRequest{Email: strings.TrimSpace(email), Password: password}
	raw, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/auth/login", bytes.NewReader(raw))
	if err != nil {
		return nil, http.StatusBadGateway, "userspanel unavailable"
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, "userspanel unavailable"
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		msg := parseUsersPanelError(body, "invalid credentials")
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
	var out usersPanelLoginResponse
	if err := json.Unmarshal(body, &out); err != nil || strings.TrimSpace(out.Token) == "" {
		return nil, http.StatusBadGateway, "invalid auth response"
	}
	perms, _ := usersPanelPermissions(ctx, baseURL, out.Token)
	return &platformSession{
		Token:       out.Token,
		Email:       out.User.Email,
		Username:    out.User.Username,
		Roles:       out.User.Roles,
		Permissions: perms,
	}, 0, ""
}

func usersPanelSession(ctx context.Context, baseURL, token string) (*platformSession, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" || strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("userspanel unavailable")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/auth/user", nil)
	if err != nil {
		return nil, fmt.Errorf("userspanel unavailable")
	}
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
	var wrap struct {
		User struct {
			Email    string   `json:"email"`
			Username string   `json:"username"`
			Roles    []string `json:"roles"`
		} `json:"user"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		return nil, fmt.Errorf("invalid user response")
	}
	perms, _ := usersPanelPermissions(ctx, baseURL, token)
	return &platformSession{
		Token:       token,
		Email:       wrap.User.Email,
		Username:    wrap.User.Username,
		Roles:       wrap.User.Roles,
		Permissions: perms,
	}, nil
}

func usersPanelPermissions(ctx context.Context, baseURL, token string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/auth/permissions", nil)
	if err != nil {
		return nil, err
	}
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

func parseUsersPanelError(body []byte, fallback string) string {
	var wrap struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &wrap) == nil && strings.TrimSpace(wrap.Error) != "" {
		return wrap.Error
	}
	return fallback
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

const (
	platformRoleAdmin = "Admin"
	permMorphBooki    = "morph_booki"
)

func hasBookiAccess(roles, permissions []string) bool {
	for _, r := range roles {
		if r == platformRoleAdmin {
			return true
		}
	}
	for _, p := range permissions {
		if strings.EqualFold(p, permMorphBooki) {
			return true
		}
	}
	return false
}

func mapBookiRole(platformRoles []string) string {
	for _, r := range platformRoles {
		if r == platformRoleAdmin {
			return "owner"
		}
	}
	return "employee"
}
