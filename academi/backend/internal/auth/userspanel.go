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

// PlatformSession is a validated UsersPanel login/session.
type PlatformSession struct {
	Token    string
	Email    string
	Username string
	Roles    []string
}

func UsersPanelLogin(ctx context.Context, baseURL, email, password string) (*PlatformSession, int, string) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, http.StatusBadGateway, "userspanel unavailable"
	}
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/auth/login", bytes.NewReader(body))
	if err != nil {
		return nil, http.StatusBadGateway, "userspanel unavailable"
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, "userspanel unavailable"
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, parseUPErr(raw, "invalid credentials")
	}
	var out struct {
		Token string `json:"token"`
		User  struct {
			Email    string   `json:"email"`
			Username string   `json:"username"`
			Roles    []string `json:"roles"`
		} `json:"user"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.Token == "" {
		return nil, http.StatusBadGateway, "invalid auth response"
	}
	return &PlatformSession{
		Token: out.Token, Email: out.User.Email, Username: out.User.Username, Roles: out.User.Roles,
	}, 0, ""
}

func UsersPanelSession(ctx context.Context, baseURL, token string) (*PlatformSession, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" || token == "" {
		return nil, fmt.Errorf("userspanel unavailable")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/auth/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
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
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return nil, err
	}
	email := strings.TrimSpace(wrap.User.Email)
	if email == "" {
		return nil, fmt.Errorf("invalid session")
	}
	return &PlatformSession{
		Token: token, Email: email, Username: wrap.User.Username, Roles: wrap.User.Roles,
	}, nil
}

func parseUPErr(body []byte, fallback string) string {
	var w struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &w) == nil && w.Error != "" {
		return w.Error
	}
	return fallback
}
