package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"idongivaflyinfa/ai"
	"idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
	"github.com/robo/dorphai"
)

type chatProviderSave struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
}

func providerSettingsJSON(row db.UserAIProvider) gin.H {
	return gin.H{
		"provider": row.Provider,
		"base_url": row.BaseURL,
		"model":    row.Model,
		"key_set":  strings.TrimSpace(row.APIKey) != "",
	}
}

// resolveChatConfig uses the saved row only when both provider and model are set.
// Otherwise the environment client stays.
func resolveChatConfig(saved *db.UserAIProvider, env dorphai.Config) dorphai.Config {
	if saved == nil || strings.TrimSpace(saved.Provider) == "" || strings.TrimSpace(saved.Model) == "" {
		return env
	}
	return dorphai.Config{
		Provider: dorphai.ProviderID(strings.TrimSpace(saved.Provider)),
		APIKey:   saved.APIKey,
		BaseURL:  strings.TrimSpace(saved.BaseURL),
		Model:    strings.TrimSpace(saved.Model),
	}
}

func (h *Handlers) savedChatConfig(c *gin.Context) (dorphai.Config, bool) {
	if h == nil || h.TranMySQL == nil || c == nil || c.Request == nil {
		return dorphai.Config{}, false
	}
	userID, ok := h.trustedLessonUserID(c)
	if !ok {
		return dorphai.Config{}, false
	}
	row, found, err := h.TranMySQL.GetUserAIProvider(c.Request.Context(), userID)
	if err != nil || !found {
		return dorphai.Config{}, false
	}
	cfg := resolveChatConfig(&row, dorphai.Config{})
	if cfg.Provider == "" || strings.TrimSpace(cfg.Model) == "" {
		return dorphai.Config{}, false
	}
	return cfg, true
}

// aiForChat returns the process-wide client when this user has no complete row.
func (h *Handlers) aiForChat(c *gin.Context) *ai.AIService {
	if h == nil {
		return nil
	}
	cfg, ok := h.savedChatConfig(c)
	if !ok || h.aiService == nil {
		return h.aiService
	}
	return h.aiService.ForChatConfig(cfg)
}

func (h *Handlers) ListChatProviders(c *gin.Context) {
	if _, ok := h.requireLessonUser(c); !ok {
		return
	}
	out := make([]gin.H, 0)
	for _, item := range dorphai.ListProviders() {
		out = append(out, gin.H{
			"id":               item.ID,
			"display_name":     item.DisplayName,
			"requires_key":     item.RequiresKey,
			"default_base_url": item.DefaultBaseURL,
		})
	}
	c.JSON(http.StatusOK, gin.H{"providers": out})
}

func (h *Handlers) GetChatProvider(c *gin.Context) {
	userID, ok := h.requireLessonUser(c)
	if !ok {
		return
	}
	row, found, err := h.TranMySQL.GetUserAIProvider(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load provider"})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"provider": "", "base_url": "", "model": "", "key_set": false})
		return
	}
	c.JSON(http.StatusOK, providerSettingsJSON(row))
}

func (h *Handlers) SaveChatProvider(c *gin.Context) {
	userID, ok := h.requireLessonUser(c)
	if !ok {
		return
	}
	var body chatProviderSave
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	info, known := dorphai.LookupProvider(dorphai.ProviderID(strings.TrimSpace(body.Provider)))
	if !known {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider"})
		return
	}
	if info.RequiresKey && strings.TrimSpace(body.APIKey) == "" {
		prev, found, err := h.TranMySQL.GetUserAIProvider(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load provider"})
			return
		}
		if !found || strings.TrimSpace(prev.APIKey) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "api key is required"})
			return
		}
	}
	row := db.UserAIProvider{
		UserID:   userID,
		Provider: string(info.ID),
		BaseURL:  strings.TrimSpace(body.BaseURL),
		APIKey:   body.APIKey,
		Model:    strings.TrimSpace(body.Model),
	}
	if row.BaseURL == "" {
		row.BaseURL = info.DefaultBaseURL
	}
	if err := h.TranMySQL.SaveUserAIProvider(c.Request.Context(), row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save provider"})
		return
	}
	saved, _, err := h.TranMySQL.GetUserAIProvider(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load provider"})
		return
	}
	c.JSON(http.StatusOK, providerSettingsJSON(saved))
}

func (h *Handlers) ScanChatProvider(c *gin.Context) {
	userID, ok := h.requireLessonUser(c)
	if !ok {
		return
	}
	var body chatProviderSave
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	key := body.APIKey
	if strings.TrimSpace(key) == "" {
		if row, found, err := h.TranMySQL.GetUserAIProvider(c.Request.Context(), userID); err == nil && found {
			key = row.APIKey
		}
	}
	ids, err := scanProviderModels(c.Request.Context(), body.Provider, body.BaseURL, key)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not scan models"})
		return
	}
	models := make([]gin.H, 0, len(ids))
	for _, id := range ids {
		models = append(models, gin.H{"id": id})
	}
	c.JSON(http.StatusOK, gin.H{"models": models})
}

func scanProviderModels(ctx context.Context, provider, baseURL, apiKey string) ([]string, error) {
	info, ok := dorphai.LookupProvider(dorphai.ProviderID(strings.TrimSpace(provider)))
	if !ok {
		return nil, errScan("unknown provider")
	}
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = info.DefaultBaseURL
	}
	endpoint := strings.TrimRight(base, "/") + "/models"
	if info.ID == dorphai.ProviderOllama {
		endpoint = ollamaTagsURL(base)
	}
	if err := validateScanURL(endpoint); err != nil {
		return nil, err
	}
	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if key := strings.TrimSpace(apiKey); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	client := &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errScan("request failed")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errScan("provider rejected the scan")
	}
	if info.ID == dorphai.ProviderOllama {
		return parseOllamaTags(body)
	}
	return parseOpenAIModels(body)
}

type scanError string

func (e scanError) Error() string { return string(e) }

func errScan(msg string) error { return scanError(msg) }

func validateScanURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil {
		return errScan("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errScan("invalid url")
	}
	return nil
}

// ollamaTagsURL calls the daemon's /api/tags. The catalog base ends in /v1, and that prefix is not part of the tags route.
func ollamaTagsURL(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	base = strings.TrimSuffix(base, "/v1")
	return strings.TrimRight(base, "/") + "/api/tags"
}

func parseOpenAIModels(body []byte) ([]string, error) {
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(payload.Data))
	for _, row := range payload.Data {
		if id := strings.TrimSpace(row.ID); id != "" {
			out = append(out, id)
		}
	}
	return out, nil
}

func parseOllamaTags(body []byte) ([]string, error) {
	var payload struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(payload.Models))
	for _, row := range payload.Models {
		if name := strings.TrimSpace(row.Name); name != "" {
			out = append(out, name)
		}
	}
	return out, nil
}
