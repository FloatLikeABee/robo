package handler

import (
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/formsx/backend/internal/mail"
	"github.com/gin-gonic/gin"
)

type eventInfoShareEmailRequest struct {
	To      []string `json:"to" binding:"required,min=1,dive,email"`
	Kind    string   `json:"kind" binding:"required,oneof=page api"`
	Message string   `json:"message"`
}

func (h *Handler) eventInfoSubmitPageURL() string {
	return strings.TrimRight(h.Cfg.PublicFormBaseURL, "/") + "/events-info/submit"
}

func (h *Handler) eventInfoPublicAPIURL() string {
	return strings.TrimRight(h.Cfg.PublicFormBaseURL, "/") + "/api/v1/public/events-info"
}

func (h *Handler) eventInfoSampleCurl() string {
	sampleTime := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf(`curl -X POST '%s' \
  -H 'Content-Type: application/json' \
  -d '{"title":"Quarterly review","detail":"Discussed roadmap.","reporter":"Jane Smith","time":"%s"}'`,
		h.eventInfoPublicAPIURL(), sampleTime)
}

// GetEventInfoCollectionInfo returns public submit page and API URLs for admins.
func (h *Handler) GetEventInfoCollectionInfo(c *gin.Context) {
	sampleTime := time.Now().UTC().Format(time.RFC3339)
	c.JSON(http.StatusOK, gin.H{
		"submit_page_url": h.eventInfoSubmitPageURL(),
		"public_api_url":  h.eventInfoPublicAPIURL(),
		"sample_curl":     h.eventInfoSampleCurl(),
		"sample_json": gin.H{
			"title":    "Quarterly review",
			"detail":   "Discussed roadmap with team.",
			"reporter": "Jane Smith",
			"time":     sampleTime,
		},
	})
}

// SendEventInfoCollectionEmail emails the public submit link or API instructions.
func (h *Handler) SendEventInfoCollectionEmail(c *gin.Context) {
	var req eventInfoShareEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	to := make([]string, 0, len(req.To))
	for _, addr := range req.To {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			to = append(to, addr)
		}
	}
	if len(to) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to is required"})
		return
	}

	var subject, body string
	switch req.Kind {
	case "page":
		subject = "Submit an Events & Info entry"
		body = h.eventInfoPageInviteHTML(req.Message)
	case "api":
		subject = "Events & Info — public API instructions"
		body = h.eventInfoAPIInstructionsHTML(req.Message)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be page or api"})
		return
	}

	if err := mail.SendHTML(h.Cfg, to, subject, body); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "sent_to": to})
}

func (h *Handler) eventInfoPageInviteHTML(message string) string {
	link := h.eventInfoSubmitPageURL()
	msgBlock := ""
	if strings.TrimSpace(message) != "" {
		msgBlock = fmt.Sprintf(`<p style="margin:0 0 16px;line-height:1.5">%s</p>`, html.EscapeString(message))
	}
	return fmt.Sprintf(`<!DOCTYPE html><html><body style="font-family:system-ui,sans-serif;color:#1e293b;line-height:1.5">
%s
<p style="margin:0 0 16px">Use the link below to submit an operational note for Events &amp; Info. No account is required.</p>
<p style="margin:0 0 20px"><a href="%s" style="display:inline-block;padding:10px 18px;background:#7c3aed;color:#fff;text-decoration:none;border-radius:8px;font-weight:600">Open submit form</a></p>
<p style="margin:0;font-size:13px;color:#64748b">Or copy this URL:<br/><a href="%s">%s</a></p>
</body></html>`, msgBlock, html.EscapeString(link), html.EscapeString(link), html.EscapeString(link))
}

func (h *Handler) eventInfoAPIInstructionsHTML(message string) string {
	apiURL := h.eventInfoPublicAPIURL()
	curl := h.eventInfoSampleCurl()
	sampleTime := time.Now().UTC().Format(time.RFC3339)
	msgBlock := ""
	if strings.TrimSpace(message) != "" {
		msgBlock = fmt.Sprintf(`<p style="margin:0 0 16px;line-height:1.5">%s</p>`, html.EscapeString(message))
	}
	return fmt.Sprintf(`<!DOCTYPE html><html><body style="font-family:system-ui,sans-serif;color:#1e293b;line-height:1.5">
%s
<p style="margin:0 0 12px">Submit Events &amp; Info entries programmatically with a <strong>public POST</strong> (no auth token).</p>
<p style="margin:0 0 8px"><strong>Endpoint</strong></p>
<p style="margin:0 0 16px;font-family:monospace;font-size:13px;background:#f1f5f9;padding:10px;border-radius:6px">POST %s</p>
<p style="margin:0 0 8px"><strong>JSON body</strong></p>
<pre style="margin:0 0 16px;font-size:12px;background:#f1f5f9;padding:10px;border-radius:6px;overflow-x:auto">{
  "title": "Quarterly review",
  "detail": "Discussed roadmap with team.",
  "reporter": "Jane Smith",
  "time": "%s"
}</pre>
<p style="margin:0 0 8px"><strong>curl example</strong></p>
<pre style="margin:0 0 16px;font-size:11px;background:#f1f5f9;padding:10px;border-radius:6px;overflow-x:auto;white-space:pre-wrap">%s</pre>
<p style="margin:0;font-size:13px;color:#64748b"><code>title</code> and <code>time</code> (RFC3339 / ISO-8601) are required.</p>
</body></html>`, msgBlock, html.EscapeString(apiURL), html.EscapeString(sampleTime), html.EscapeString(curl))
}
