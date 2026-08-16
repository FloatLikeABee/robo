package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

const eventInfoAIDraftSystem = `You draft operational Events & Info records for a field / construction / ops team.

Given the user's prompt, return ONLY one JSON object (no markdown fences):
{
  "title": "short concrete title",
  "detail": "markdown body with what happened, context, and next steps if relevant",
  "reporter": "name or role if stated, otherwise empty string",
  "time": "RFC3339 timestamp if a time is implied, otherwise empty string"
}

Rules:
- title is required and must be specific (not "Event" or "Note").
- detail may use Markdown (headings, bullets, bold). Do not invent people, sites, or numbers that were not implied.
- If the prompt gives no clock time, leave "time" empty so the app can use now.
- Prefer the user's wording when they already named the event.`

type eventInfoAIDraftBody struct {
	Prompt string `json:"prompt"`
	Query  string `json:"query"`
}

type eventInfoAIDraftResult struct {
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Reporter string `json:"reporter"`
	Time     string `json:"time"` // RFC3339 when known
}

// DraftEventInfoAI POST /events-info/ai-draft
//
// Turns a free-text prompt into a draft Events & Info entry. Nothing is stored —
// the client reviews and saves through the normal create route.
func (h *Handler) DraftEventInfoAI(c *gin.Context) {
	var body eventInfoAIDraftBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prompt := strings.TrimSpace(body.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(body.Query)
	}
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}
	if h.AI == nil || !h.AI.Configured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI is not configured — set MORPH_AI_API_KEY to generate Events & Info from a prompt",
		})
		return
	}

	userPrompt := fmt.Sprintf(
		"Draft an Events & Info record from this prompt:\n\n%s\n\nCurrent UTC time for reference: %s",
		prompt,
		time.Now().UTC().Format(time.RFC3339),
	)

	reply, err := h.AI.ChatCompletion(c.Request.Context(), []morphai.Message{
		{Role: "system", Content: eventInfoAIDraftSystem},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI request failed: " + err.Error()})
		return
	}

	raw, ok := morphai.ExtractJSONObject(reply)
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI reply was not valid JSON", "raw": truncateStr(reply, 800)})
		return
	}

	var draft eventInfoAIDraftResult
	if err := json.Unmarshal([]byte(raw), &draft); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not parse AI draft: " + err.Error()})
		return
	}

	draft.Title = strings.TrimSpace(draft.Title)
	draft.Detail = strings.TrimSpace(draft.Detail)
	draft.Reporter = strings.TrimSpace(draft.Reporter)
	draft.Time = strings.TrimSpace(draft.Time)

	if draft.Title == "" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI draft was missing a title", "raw": raw})
		return
	}
	if len(draft.Title) > eventInfoMaxTitleLen {
		draft.Title = draft.Title[:eventInfoMaxTitleLen]
	}
	if len(draft.Detail) > eventInfoMaxDetailLen {
		draft.Detail = draft.Detail[:eventInfoMaxDetailLen]
	}
	if len(draft.Reporter) > eventInfoMaxReporterLen {
		draft.Reporter = draft.Reporter[:eventInfoMaxReporterLen]
	}

	// Normalize time: accept RFC3339 or leave empty for the client to use now.
	if draft.Time != "" {
		if t, err := time.Parse(time.RFC3339, draft.Time); err == nil {
			draft.Time = t.UTC().Format(time.RFC3339)
		} else if t, err := time.Parse(time.RFC3339Nano, draft.Time); err == nil {
			draft.Time = t.UTC().Format(time.RFC3339)
		} else {
			draft.Time = ""
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"title":             draft.Title,
		"detail":            draft.Detail,
		"reporter":          draft.Reporter,
		"time":              draft.Time,
		"assistant_message": fmt.Sprintf("Drafted “%s”. Review and save, or edit first.", draft.Title),
	})
}
