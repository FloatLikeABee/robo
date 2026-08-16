package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
)

const maxManagementAPIBody = 512 * 1024

// SetGinEngine must be called after all API routes are registered on the engine so the assistant can dispatch internal requests.
func (h *Handlers) SetGinEngine(e *gin.Engine) {
	h.ginEngine = e
}

func allowedManagementPath(path string) bool {
	return strings.HasPrefix(path, "/api/tran/") ||
		strings.HasPrefix(path, "/api/forms/") ||
		strings.HasPrefix(path, "/api/formsx/") ||
		strings.HasPrefix(path, "/api/sheetx/") ||
		strings.HasPrefix(path, "/api/composerx/") ||
		strings.HasPrefix(path, "/api/knowledge/") ||
		strings.HasPrefix(path, "/api/graph/") ||
		strings.HasPrefix(path, "/api/skills")
}

func allowedManagementMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}

// trimManagementPath strips trailing junk (e.g. model pasted "Host:" or a second path) and whitespace.
func trimManagementPath(p string) string {
	p = strings.TrimSpace(p)
	if i := strings.IndexAny(p, " \t\r\n"); i >= 0 {
		p = p[:i]
	}
	return p
}

// sanitizeQueryForURL trims assistant noise and encodes spaces so the request URL parses; raw "&key=value with spaces" would otherwise break net/http.
func sanitizeQueryForURL(q string) string {
	q = strings.TrimSpace(q)
	if i := strings.IndexAny(q, "\r\n"); i >= 0 {
		q = q[:i]
	}
	q = strings.TrimSpace(q)
	if q == "" {
		return ""
	}
	// Prefer proper encoding via ParseQuery when the string looks like a query.
	vals, err := url.ParseQuery(strings.ReplaceAll(q, " ", "+"))
	if err == nil {
		return vals.Encode()
	}
	return strings.ReplaceAll(q, " ", "%20")
}

func (h *Handlers) execManagementAPI(c *gin.Context, method, path, query string, body []byte) (int, []byte) {
	if h.ginEngine == nil {
		return 500, []byte(`{"error":"router not configured for assistant tools"}`)
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = http.MethodGet
	}
	if !allowedManagementMethod(method) {
		return 400, []byte(`{"error":"invalid HTTP method"}`)
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return 400, []byte(`{"error":"missing path"}`)
	}
	// If the model put ?query inside path, fold it into RawQuery.
	if idx := strings.Index(path, "?"); idx >= 0 {
		extra := path[idx+1:]
		path = trimManagementPath(path[:idx])
		q := sanitizeQueryForURL(query)
		if extra != "" {
			if q != "" {
				q = extra + "&" + q
			} else {
				q = extra
			}
		}
		query = q
	} else {
		path = trimManagementPath(path)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !allowedManagementPath(path) {
		return 403, []byte(`{"error":"path not allowed; only /api/tran/*, /api/forms/*, /api/formsx/*, /api/sheetx/*, /api/composerx/*, /api/knowledge/*, /api/graph/*"}`)
	}
	for _, r := range path {
		if unicode.IsControl(r) {
			return 400, []byte(`{"error":"invalid characters in path"}`)
		}
	}
	if len(body) > maxManagementAPIBody {
		return 400, []byte(`{"error":"request body too large"}`)
	}
	if strings.HasPrefix(path, "/api/formsx/") || strings.HasPrefix(path, "/api/sheetx/") {
		return h.execFormsXProxy(c, method, path, query, body)
	}
	if strings.HasPrefix(path, "/api/composerx/") {
		return h.execComposerXProxy(c, method, path, query, body)
	}
	rawQ := sanitizeQueryForURL(query)
	u := url.URL{
		Scheme: "http",
		Host:   "127.0.0.1",
		Path:   path,
	}
	if rawQ != "" {
		u.RawQuery = rawQ
	}
	reqURL := u.String()
	req, err := http.NewRequestWithContext(c.Request.Context(), method, reqURL, bytes.NewReader(body))
	if err != nil {
		return 400, []byte(`{"error":"invalid assistant request URL"}`)
	}

	// AuthzMiddleware needs the same credentials as the outer request. Legacy mode used
	// X-User-ID + role headers; Morph login uses Bearer only — without forwarding
	// Authorization, internal tool calls always get 401 "authentication required".
	if auth := strings.TrimSpace(c.GetHeader("Authorization")); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	uid := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if uid == "" {
		uid = "admin"
	}
	req.Header.Set("X-User-ID", uid)
	for _, key := range []string{"X-User-Role", "X-User-Roles", "X-User-Permissions"} {
		if v := c.GetHeader(key); v != "" {
			req.Header.Set(key, v)
		}
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	h.ginEngine.ServeHTTP(w, req)
	return w.Code, w.Body.Bytes()
}

func (h *Handlers) execFormsXProxy(c *gin.Context, method, path, query string, body []byte) (int, []byte) {
	base := strings.TrimSuffix(strings.TrimSpace(h.tranFormBase), "/")
	if base == "" {
		return 503, []byte(`{"error":"SheetX not configured — set TRANFORM_BASE_URL"}`)
	}
	rest := strings.TrimPrefix(path, "/api/formsx")
	if strings.HasPrefix(path, "/api/sheetx") {
		rest = strings.TrimPrefix(path, "/api/sheetx")
	}
	targetPath := "/api/v1" + rest
	rawQ := sanitizeQueryForURL(query)
	u := base + targetPath
	if rawQ != "" {
		u += "?" + rawQ
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), method, u, bytes.NewReader(body))
	if err != nil {
		return 400, []byte(`{"error":"invalid SheetX proxy request"}`)
	}
	if auth := strings.TrimSpace(c.GetHeader("Authorization")); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 502, []byte(`{"error":"SheetX request failed"}`)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return 502, []byte(`{"error":"SheetX response read failed"}`)
	}
	return resp.StatusCode, respBody
}

func (h *Handlers) execComposerXProxy(c *gin.Context, method, path, query string, body []byte) (int, []byte) {
	base := strings.TrimSuffix(strings.TrimSpace(h.tranMailBase), "/")
	if base == "" {
		return 503, []byte(`{"error":"ComposerX not configured — set TRAN_MAIL_BASE_URL"}`)
	}
	targetPath := strings.TrimPrefix(path, "/api/composerx")
	if targetPath == "" {
		targetPath = "/"
	}
	rawQ := sanitizeQueryForURL(query)
	u := base + targetPath
	if rawQ != "" {
		u += "?" + rawQ
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), method, u, bytes.NewReader(body))
	if err != nil {
		return 400, []byte(`{"error":"invalid ComposerX proxy request"}`)
	}
	if auth := strings.TrimSpace(c.GetHeader("Authorization")); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 502, []byte(`{"error":"ComposerX request failed"}`)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return 502, []byte(`{"error":"ComposerX response read failed"}`)
	}
	return resp.StatusCode, respBody
}
