package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

// Config mirrors core environment + config.json for Phase 1
type Config struct {
	SQLitePath        string
	BadgerPath        string
	FileStorePath     string
	OpenAIAPIKey      string
	OpenAIChatModel   string
	QwenAPIKey        string
	QwenBaseURL       string
	QwenDefaultModel  string
	UsersPanelBaseURL string
}

type App struct {
	cfg            Config
	sqlDB          *sql.DB
	badgerDB       *badger.DB
	referenceDocs  *ReferenceDocsStore
	httpRouter     *gin.Engine
	templates      *TemplateRepository
	mergeData      *MergeDataRepository
	savedEmails    *SavedEmailRepository
	publishedPages *PublishedPageRepository
	publishDrafts  *PublishDraftRepository
	ai             *morphai.Client
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// corsMiddleware handles browser CORS preflight. Using "Allow-Headers: *" is unreliable for
// preflight; echo Access-Control-Request-Headers and reflect Origin for localhost dev.
func corsMiddleware() gin.HandlerFunc {
	const defaultAllowedHeaders = "Content-Type, Accept, Authorization, X-Requested-With, X-User-Role, X-User-Roles, X-User-Permissions"
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin != "" && isAllowedDevOrigin(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
		reqHdr := c.GetHeader("Access-Control-Request-Headers")
		if reqHdr != "" {
			c.Header("Access-Control-Allow-Headers", reqHdr)
		} else {
			c.Header("Access-Control-Allow-Headers", defaultAllowedHeaders)
		}
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func isAllowedDevOrigin(origin string) bool {
	u := strings.ToLower(origin)
	return strings.HasPrefix(u, "http://localhost:") ||
		strings.HasPrefix(u, "http://127.0.0.1:") ||
		strings.HasPrefix(u, "https://localhost:") ||
		strings.HasPrefix(u, "https://127.0.0.1:")
}

func requireTranmailAccess(usersPanelBaseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || strings.HasPrefix(c.Request.URL.Path, "/auth/") || strings.HasPrefix(c.Request.URL.Path, "/public/") {
			c.Next()
			return
		}
		role := resolveScopedRole(c.GetHeader("X-User-Role"), c.GetHeader("X-User-Roles"))
		rawPermissions := c.GetHeader("X-User-Permissions")
		token := bearerToken(c.GetHeader("Authorization"))
		var panelPerms []string
		if token != "" {
			r, perms := resolveRoleAndPermissionsFromUsersPanel(usersPanelBaseURL, token)
			panelPerms = perms
			if r != "" {
				role = r
			}
			if len(perms) > 0 {
				rawPermissions = strings.Join(perms, ",")
				c.Request.Header.Set("X-User-Permissions", rawPermissions)
			}
		}
		// Morph-hosted auth: any authenticated user may use ComposerX.
		if role == "admin" || len(panelPerms) > 0 {
			c.Next()
			return
		}
		if role == "employee" && hasCSVPermission(rawPermissions, "compose_email") {
			c.Next()
			return
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "tranmail access is restricted by admin policy"})
	}
}

func resolveScopedRole(primary, fallback string) string {
	v := strings.ToLower(strings.TrimSpace(primary))
	if v != "" {
		return v
	}
	role := ""
	for _, raw := range strings.Split(fallback, ",") {
		r := strings.ToLower(strings.TrimSpace(raw))
		switch r {
		case "admin":
			return "admin"
		case "employee":
			role = "employee"
		case "member":
			if role == "" {
				role = "member"
			}
		case "forms", "email composer", "composerx", "main panel", "sharp reports":
			if role != "admin" {
				role = "employee"
			}
		}
	}
	return role
}

func hasCSVPermission(rawCSV, needed string) bool {
	needed = strings.ToLower(strings.TrimSpace(needed))
	for _, raw := range strings.Split(rawCSV, ",") {
		if strings.ToLower(strings.TrimSpace(raw)) == needed {
			return true
		}
	}
	return false
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

func resolveRoleAndPermissionsFromUsersPanel(baseURL, token string) (string, []string) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" || token == "" {
		return "", nil
	}
	reqUser, _ := http.NewRequest(http.MethodGet, baseURL+"/api/auth/user", nil)
	reqUser.Header.Set("Authorization", "Bearer "+token)
	userResp, err := http.DefaultClient.Do(reqUser)
	if err != nil {
		return "", nil
	}
	defer userResp.Body.Close()
	if userResp.StatusCode != http.StatusOK {
		return "", nil
	}
	userBody, _ := io.ReadAll(userResp.Body)
	var userPayload struct {
		User struct {
			Roles []string `json:"roles"`
		} `json:"user"`
	}
	if json.Unmarshal(userBody, &userPayload) != nil {
		return "", nil
	}
	role := resolveScopedRole("", strings.Join(userPayload.User.Roles, ","))

	reqPerms, _ := http.NewRequest(http.MethodGet, baseURL+"/api/auth/permissions", nil)
	reqPerms.Header.Set("Authorization", "Bearer "+token)
	permsResp, err := http.DefaultClient.Do(reqPerms)
	if err != nil {
		return role, nil
	}
	defer permsResp.Body.Close()
	if permsResp.StatusCode != http.StatusOK {
		return role, nil
	}
	permsBody, _ := io.ReadAll(permsResp.Body)
	var permsPayload struct {
		Permissions []string `json:"permissions"`
	}
	if json.Unmarshal(permsBody, &permsPayload) != nil {
		return role, nil
	}
	return role, permsPayload.Permissions
}

// --- bootstrap ---

func main() {
	cfg := Config{
		SQLitePath:        getEnv("COMPOSERX_SQLITE_PATH", "./data/composerx.sqlite"),
		BadgerPath:        getEnv("COMPOSERX_BADGER_PATH", "./data/composerx_badger"),
		FileStorePath:     getEnv("TRAN_FILE_STORAGE_PATH", "./storage"),
		OpenAIAPIKey:      getEnv("TRAN_OPENAI_API_KEY", ""),
		OpenAIChatModel:   getEnv("TRAN_OPENAI_MODEL", "gpt-4o-mini"),
		UsersPanelBaseURL: getEnv("USERS_PANEL_BASE_URL", "http://127.0.0.1:9090"),
	}
	mergeAIConfigFromFile(&cfg, getEnv("TRAN_AI_CONFIG_PATH", "ai.config.json"))
	mergeMorphAIEnv(&cfg)
	normalizeQwenDefaults(&cfg)

	sqlDB, err := openComposerXSQLite(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("failed to open sqlite: %v", err)
	}

	badgerDB, err := openComposerXBadger(cfg.BadgerPath)
	if err != nil {
		log.Fatalf("failed to open badger: %v", err)
	}

	router := gin.New()
	router.Use(corsMiddleware())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(requireTranmailAccess(cfg.UsersPanelBaseURL))

	emailBodies := NewEmailContentStore(badgerDB)
	refDocs := NewReferenceDocsStore(badgerDB)

	app := &App{
		cfg:            cfg,
		sqlDB:          sqlDB,
		badgerDB:       badgerDB,
		referenceDocs:  refDocs,
		httpRouter:     router,
		templates:      NewTemplateRepository(sqlDB),
		mergeData:      NewMergeDataRepository(sqlDB),
		savedEmails:    NewSavedEmailRepository(sqlDB, emailBodies),
		publishedPages: NewPublishedPageRepository(sqlDB, emailBodies),
		publishDrafts:  NewPublishDraftRepository(sqlDB),
		ai:             morphai.NewClient(morphai.LoadFromEnv()),
	}

	app.registerRoutes()

	port := getEnv("PORT", "8043")
	log.Printf("ComposerX listening on :%s (sqlite=%s badger=%s)", port, cfg.SQLitePath, cfg.BadgerPath)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start HTTP server on %s: %v", port, err)
	}
}

func (a *App) registerRoutes() {
	r := a.httpRouter

	// health
	r.GET("/health", a.handleHealth)
	r.POST("/auth/login", a.handleLogin)
	r.GET("/auth/me", a.handleMe)

	// templates
	r.GET("/templates", a.listTemplates)
	r.GET("/templates/:id", a.getTemplate)
	r.POST("/templates", a.createTemplate)
	r.PUT("/templates/:id", a.updateTemplate)
	r.DELETE("/templates/:id", a.deleteTemplate)

	// merge data
	r.GET("/merge-data", a.listMergeDataSources)
	r.POST("/merge-data/upload", a.uploadMergeDataSource)
	r.GET("/merge-data/:id/download", a.downloadMergeDataSource)
	r.DELETE("/merge-data/:id", a.deleteMergeDataSource)

	// saved emails (SQLite index + Badger body)
	r.GET("/emails", a.listSavedEmails)
	r.POST("/emails", a.createSavedEmail)
	r.GET("/emails/:id", a.getSavedEmail)
	r.PUT("/emails/:id", a.updateSavedEmail)
	r.DELETE("/emails/:id", a.deleteSavedEmail)
	r.GET("/publishes/resolve-path", a.resolvePublishSlug)
	r.GET("/publishes/history", a.listPublishedPages)
	r.POST("/publishes", a.createPublishedPage)
	r.GET("/publish-drafts", a.listPublishDrafts)
	r.POST("/publish-drafts", a.createPublishDraft)
	r.GET("/publish-drafts/:id", a.getPublishDraft)
	r.DELETE("/publish-drafts/:id", a.deletePublishDraft)

	// reports
	r.GET("/reports/available", a.listReports)
	r.POST("/reports/order", a.orderReport)
	r.GET("/reports/status/:guid", a.getReportStatus)

	// AI composer + reference document RAG
	r.GET("/ai/reference-docs", a.listReferenceDocuments)
	r.POST("/ai/reference-docs/upload", a.uploadReferenceDocument)
	r.GET("/ai/reference-docs/:id/download", a.downloadReferenceDocument)
	r.DELETE("/ai/reference-docs/:id", a.deleteReferenceDocument)
	r.POST("/ai/composer-chat", a.composerAIChat)
	r.POST("/ai/publish-chat", a.publishAIChat)
	r.POST("/ai/web-search", a.WebSearch)
	r.GET("/ai/app-abilities", a.ComposerXAppAbilitiesMCP)
	r.GET("/ai/mcp-tools", a.ComposerXAppAbilitiesMCP)
	r.POST("/ai/publish-sources/process", a.processPublishSources)
	r.POST("/ai/assistant/chat", a.platformAssistantChat)

	// public published pages (no auth)
	r.GET("/public/published/:slug", a.getPublishedPageJSON)
	r.GET("/public/p/:slug", a.servePublishedPage)
}

// --- handlers (Phase 1: minimal stubs, wired to DB later) ---

func (a *App) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	})
}

func (a *App) handleLogin(c *gin.Context) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	payload := map[string]string{
		"email":    strings.TrimSpace(in.Email),
		"password": in.Password,
	}
	raw, _ := json.Marshal(payload)
	endpoint := strings.TrimRight(a.cfg.UsersPanelBaseURL, "/") + "/api/auth/login"
	req, _ := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "userspanel unavailable"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		msg := "invalid credentials"
		var wrap struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &wrap) == nil && strings.TrimSpace(wrap.Error) != "" {
			msg = wrap.Error
		}
		if resp.StatusCode == http.StatusUnauthorized && strings.EqualFold(strings.TrimSpace(msg), "unauthorized") {
			msg = "Invalid email or password"
		}
		switch resp.StatusCode {
		case http.StatusBadRequest:
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		case http.StatusUnauthorized:
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
		case http.StatusForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": msg})
		default:
			c.JSON(http.StatusBadGateway, gin.H{"error": "userspanel login failed"})
		}
		return
	}
	var loginResp map[string]any
	if json.Unmarshal(body, &loginResp) != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid auth response"})
		return
	}
	token, _ := loginResp["token"].(string)
	role, perms := resolveRoleAndPermissionsFromUsersPanel(a.cfg.UsersPanelBaseURL, token)
	loginResp["resolved_role"] = role
	loginResp["permissions"] = perms
	c.JSON(http.StatusOK, loginResp)
}

func (a *App) handleMe(c *gin.Context) {
	token := bearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
		return
	}
	endpoint := strings.TrimRight(a.cfg.UsersPanelBaseURL, "/") + "/api/auth/user"
	req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "userspanel unavailable"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}
	var userPayload map[string]any
	if json.Unmarshal(body, &userPayload) != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid user payload"})
		return
	}
	role, perms := resolveRoleAndPermissionsFromUsersPanel(a.cfg.UsersPanelBaseURL, token)
	userPayload["resolved_role"] = role
	userPayload["permissions"] = perms
	c.JSON(http.StatusOK, userPayload)
}

// email templates
func (a *App) listTemplates(c *gin.Context) {
	ctx := c.Request.Context()

	limit := 50
	offset := 0

	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	items, total, err := a.templates.List(ctx, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (a *App) getTemplate(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	t, err := a.templates.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load template"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (a *App) createTemplate(c *gin.Context) {
	var payload struct {
		Name        string `json:"name"`
		Tag         string `json:"tag"`
		Description string `json:"description"`
		HTMLContent string `json:"html_content"`
		CreatedBy   int64  `json:"created_by"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if payload.Name == "" || payload.HTMLContent == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and html_content are required"})
		return
	}

	ctx := c.Request.Context()

	id, err := a.templates.Create(ctx, payload.Name, payload.Tag, payload.Description, payload.HTMLContent, payload.CreatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (a *App) updateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var payload struct {
		Name        string `json:"name"`
		Tag         string `json:"tag"`
		Description string `json:"description"`
		HTMLContent string `json:"html_content"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if payload.Name == "" || payload.HTMLContent == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and html_content are required"})
		return
	}

	ctx := c.Request.Context()

	if err := a.templates.Update(ctx, id, payload.Name, payload.Tag, payload.Description, payload.HTMLContent); err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update template"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *App) deleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()

	if err := a.templates.Delete(ctx, id); err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete template"})
		return
	}

	c.Status(http.StatusNoContent)
}

// --- saved emails (SQL metadata + Badger body) ---

func (a *App) listSavedEmails(c *gin.Context) {
	ctx := c.Request.Context()
	limit := 50
	offset := 0
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	rows, total, err := a.savedEmails.List(ctx, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list emails"})
		return
	}

	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		updated := ""
		if row.UpdatedAt.Valid {
			updated = row.UpdatedAt.Time.UTC().Format(time.RFC3339)
		}
		items = append(items, gin.H{
			"id":         row.ID,
			"name":       row.Name,
			"updated_at": updated,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func normalizeComposerAISessionJSON(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "null" {
		return ""
	}
	if !json.Valid([]byte(s)) {
		return ""
	}
	return s
}

func resolveSavedContentBody(markdown, html string) string {
	if body := strings.TrimSpace(markdown); body != "" {
		return markdown
	}
	return html
}

func (a *App) createSavedEmail(c *gin.Context) {
	var payload struct {
		Name              string          `json:"name"`
		MarkdownContent   string          `json:"markdown_content"`
		HTMLContent       string          `json:"html_content"` // legacy alias
		CreatedBy         int64           `json:"created_by"`
		ComposerAISession json.RawMessage `json:"composer_ai_session"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if payload.CreatedBy <= 0 {
		payload.CreatedBy = 1
	}
	body := resolveSavedContentBody(payload.MarkdownContent, payload.HTMLContent)

	ctx := c.Request.Context()
	id, err := a.savedEmails.Create(ctx, payload.Name, body, normalizeComposerAISessionJSON(payload.ComposerAISession), payload.CreatedBy)
	if err != nil {
		if err.Error() == "name and markdown required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save content"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (a *App) getSavedEmail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ctx := c.Request.Context()
	detail, err := a.savedEmails.GetDetail(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "email not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load email"})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (a *App) updateSavedEmail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var payload struct {
		Name              string          `json:"name"`
		MarkdownContent   string          `json:"markdown_content"`
		HTMLContent       string          `json:"html_content"` // legacy alias
		ComposerAISession json.RawMessage `json:"composer_ai_session"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	body := resolveSavedContentBody(payload.MarkdownContent, payload.HTMLContent)
	ctx := c.Request.Context()
	if err := a.savedEmails.Update(ctx, id, payload.Name, body, normalizeComposerAISessionJSON(payload.ComposerAISession)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "content not found"})
			return
		}
		if err.Error() == "name and markdown required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update content"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *App) deleteSavedEmail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ctx := c.Request.Context()
	if err := a.savedEmails.Delete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "email not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete email"})
		return
	}
	c.Status(http.StatusNoContent)
}

// merge data sources
func (a *App) listMergeDataSources(c *gin.Context) {
	ctx := c.Request.Context()

	limit := 50
	offset := 0

	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	items, total, err := a.mergeData.List(ctx, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list merge data sources"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (a *App) uploadMergeDataSource(c *gin.Context) {
	ctx := c.Request.Context()

	name := c.PostForm("name")
	fileType := c.PostForm("file_type")
	uploadedByStr := c.PostForm("uploaded_by")

	if name == "" || fileType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and file_type are required"})
		return
	}

	var uploadedBy int64
	if uploadedByStr != "" {
		if v, err := strconv.ParseInt(uploadedByStr, 10, 64); err == nil {
			uploadedBy = v
		}
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	dir := a.cfg.FileStorePath + "/merge-data"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare storage"})
		return
	}

	filename := time.Now().UTC().Format("20060102T150405") + "_" + header.Filename
	targetPath := dir + "/" + filename

	dst, err := os.Create(targetPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write file"})
		return
	}

	id, err := a.mergeData.Create(ctx, name, fileType, targetPath, uploadedBy)
	if err != nil {
		_ = os.Remove(targetPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist merge data source"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (a *App) downloadMergeDataSource(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	src, err := a.mergeData.GetByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "merge data source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load merge data source"})
		return
	}
	if strings.TrimSpace(src.FilePath) == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not available"})
		return
	}
	name := strings.TrimSpace(src.Name)
	if name == "" {
		name = filepath.Base(src.FilePath)
	}
	c.FileAttachment(src.FilePath, name)
}

func (a *App) deleteMergeDataSource(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	src, err := a.mergeData.GetByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "merge data source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load merge data source"})
		return
	}

	if err := a.mergeData.Delete(ctx, id); err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "merge data source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete merge data source"})
		return
	}

	if src.FilePath != "" {
		_ = os.Remove(src.FilePath)
	}

	c.Status(http.StatusNoContent)
}

// reports
func (a *App) listReports(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "total": 0})
}

func (a *App) orderReport(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{"order_guid": "placeholder-order-guid"})
}

func (a *App) getReportStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"order_guid": c.Param("guid"),
		"status":     "pending",
	})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
