package handlers

import (
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

// ComposerXProxy forwards /api/composerx/* to ComposerX (TRAN_MAIL_BASE_URL).
func (h *Handlers) ComposerXProxy(c *gin.Context) {
	suffix := c.Param("filepath")
	path := "/api/composerx" + suffix
	if suffix != "" && !strings.HasPrefix(suffix, "/") {
		path = "/api/composerx/" + suffix
	}

	var body []byte
	if c.Request.Body != nil {
		var err error
		body, err = io.ReadAll(io.LimitReader(c.Request.Body, 8<<20))
		if err != nil {
			c.JSON(502, gin.H{"error": "ComposerX request read failed"})
			return
		}
	}

	status, respBody := h.execComposerXProxy(c, c.Request.Method, path, c.Request.URL.RawQuery, body)
	ct := "application/json"
	if len(respBody) > 0 && respBody[0] != '{' && respBody[0] != '[' {
		ct = "text/plain; charset=utf-8"
	}
	c.Data(status, ct, respBody)
}
