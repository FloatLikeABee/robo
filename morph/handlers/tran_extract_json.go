package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"

	"idongivaflyinfa/ai"

	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

const extractJSONMaxRunes = 16000

// ExtractJSONFromText POST /api/tran/extract-json
// { "text": "...", "purpose": "generic_data" | "asset_detail" } → { "json": { ... } }
func (h *Handlers) ExtractJSONFromText(c *gin.Context) {
	var in struct {
		Text    string `json:"text"`
		Purpose string `json:"purpose"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	purpose := strings.ToLower(strings.TrimSpace(in.Purpose))
	if purpose != "generic_data" && purpose != "asset_detail" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "purpose must be generic_data or asset_detail"})
		return
	}
	text := strings.TrimSpace(in.Text)
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
		return
	}
	if utf8.RuneCountInString(text) > extractJSONMaxRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text too long"})
		return
	}
	if h.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}

	out, err := h.aiService.ChatCompletionLong(c.Request.Context(), []ai.DashScopeMessage{
		{Role: "user", Content: extractJSONPrompt(purpose, text)},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	obj, ok := morphai.ExtractJSONObject(out)
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI did not return valid JSON"})
		return
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(obj), &parsed); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI did not return valid JSON"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"json": parsed})
}

func extractJSONPrompt(purpose, text string) string {
	role := "structured operational data"
	if purpose == "asset_detail" {
		role = "an asset / equipment / vehicle record's extended detail fields"
	}
	return `Extract a single JSON object from the user's description of ` + role + `.
Rules:
- Reply with ONLY a JSON object. No markdown fences, no commentary.
- Use clear snake_case keys.
- Infer reasonable field names from the text. Do not invent facts that are not implied.
- Nested objects and arrays are allowed when they match the description.
- If the text lists multiple records, use a "records" array.

User description:
` + text
}
