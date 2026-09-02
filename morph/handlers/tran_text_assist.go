package handlers

import (
	"context"
	"net/http"
	"strings"

	"idongivaflyinfa/ai"

	"github.com/gin-gonic/gin"
)

const (
	textAssistMaxSeed = 2000
	textAssistMaxText = 16000
)

// TranTextAssist returns short AI-generated or improved text for notes, todos, task chains, and similar helpers.
// POST /api/tran/text-assist  { "mode": "...", "kind": "note|todo", "seed": "...", "text": "..." }
func (h *Handlers) TranTextAssist(c *gin.Context) {
	if h.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}
	var in struct {
		Mode string `json:"mode"`
		Kind string `json:"kind"`
		Seed string `json:"seed"`
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	mode := strings.ToLower(strings.TrimSpace(in.Mode))
	kind := strings.ToLower(strings.TrimSpace(in.Kind))
	seed := strings.TrimSpace(in.Seed)
	text := strings.TrimSpace(in.Text)
	maxSeedRunes := textAssistMaxSeed
	if mode == "task_chain_step" {
		maxSeedRunes = 8000
	}
	if len([]rune(seed)) > maxSeedRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "seed too long"})
		return
	}
	if len([]rune(text)) > textAssistMaxText {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text too long"})
		return
	}

	ctx := context.Background()
	var prompt string
	switch mode {
	case "generate_note":
		prompt = textAssistGenerateNotePrompt(seed)
	case "generate_todo":
		prompt = textAssistGenerateTodoPrompt(seed)
	case "improve":
		if text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "text required for improve"})
			return
		}
		prompt = textAssistImprovePrompt(kind, seed, text)
	case "task_chain_step":
		if seed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "task description required for task_chain_step"})
			return
		}
		prompt = textAssistTaskChainStepPrompt(seed, text)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown mode"})
		return
	}

	out, err := h.aiService.ChatCompletion(ctx, []ai.DashScopeMessage{{Role: "user", Content: prompt}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out = strings.TrimSpace(out)
	out = strings.TrimPrefix(out, "```")
	out = strings.TrimSuffix(out, "```")
	out = strings.TrimSpace(out)
	c.JSON(http.StatusOK, gin.H{"text": out})
}

const textAssistBodyOnlyRules = `Follow only that title as the topic. Do not invent a different subject. Do not output a replacement title (the app keeps the user's title). Plain text, no preamble, no markdown fences. Output only the body.`

func textAssistGenerateNotePrompt(seed string) string {
	if seed == "" {
		return `Write a concise personal note (2–6 short paragraphs) on a generic personal or work topic. Plain text, no preamble, no markdown fences. Output only the body.`
	}
	return `Title: "` + seed + `"

Write a concise personal note (2–6 short paragraphs) about that title.
` + textAssistBodyOnlyRules
}

func textAssistGenerateTodoPrompt(seed string) string {
	if seed == "" {
		return `Write an actionable TODO body: up to 5 bullet subtasks for a generic personal or work task. Plain text, no preamble, no markdown fences. Output only the body.`
	}
	return `Title: "` + seed + `"

Write an actionable TODO body for that title: up to 5 bullet subtasks.
` + textAssistBodyOnlyRules
}

func textAssistImprovePrompt(kind, seed, text string) string {
	kindHint := "note or todo"
	switch kind {
	case "note":
		kindHint = "note"
	case "todo":
		kindHint = "todo"
	}
	extra := ""
	if strings.TrimSpace(seed) != "" {
		extra = "\nItem title (keep this topic):\n" + seed + "\n"
	}
	return `Improve the following ` + kindHint + ` text for clarity, tone, and usefulness. Keep the same meaning; fix grammar; be concise. Plain text only, no markdown fences, no preamble or closing remarks.` + extra + `

Text to improve:
` + text + `

Improved text:`
}

func textAssistTaskChainStepPrompt(taskDescription, priorOutputs string) string {
	prior := strings.TrimSpace(priorOutputs)
	priorBlock := "(none yet — this is the first task node.)\n"
	if prior != "" {
		priorBlock = prior + "\n"
	}
	return `You are Morph AI executing one step in a user's task chain. Follow the current task text as the only domain.

Current task:
` + taskDescription + `

Prior outputs from earlier task nodes (plain text, may be empty):
` + priorBlock + `
Fulfill the current task in a single response.

Rules:
- Produce concrete, useful output: drafts, summaries, plans, email text, checklists, tables in plain text, or step-by-step instructions — whatever fits the task.
- If the task implies live web search, sending email, or fetching private app data you cannot access, say so briefly and still deliver the best possible draft, outline, or template from context and general knowledge.
- Build on prior outputs when relevant; note corrections instead of silently contradicting earlier steps.
- Plain text only (no markdown code fences unless the user explicitly asked for code).

Response:`
}
