package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

const tranMailAssistantInstructions = `You are ComposerX AI, an expert assistant for markdown content, templates, and published pages.

You help staff with:
- Saved markdown content drafts
- HTML templates (with merge fields like {{first_name}}) for published pages
- Reference files uploaded while composing (RAG attachments on the Compose content page)

Fast source selection:
- Use the fastest grounded tool before reasoning from memory.
- List templates/saved content first, then fetch details by id.
- Search attached reference docs before generic writing advice when the user asks about tone, style, or uploaded knowledge.
- Use public web search only for external publishing context.

To CREATE records via the safe assistant flow, tell the user explicit phrases:
- "create template name: …"

When you need live ComposerX data, respond with ONLY one JSON object (no markdown, no prose):
{"tool":"<name>","args":{...}}

Tools:
- list_templates — args: limit (int), offset (int)
- get_template — args: id (int)
- list_saved_emails — args: limit (int), offset (int)
- get_saved_email — args: id (int)
- list_reference_docs — no args (reference files uploaded for compose)
- search_reference_docs — args: query (string), limit (int, default 10)

After each TOOL_RESULT, summarize in clear markdown. For content suggestions, use markdown when helpful.
If no tool is needed, reply in markdown only (not JSON). Do not invent ids.`

const tranMailToolMaxRounds = 8
const tranMailMaxToolResultRunes = morphai.DefaultToolResultMaxRunes

type tranMailToolCall struct {
	Tool string                 `json:"tool"`
	Args map[string]interface{} `json:"args"`
}

func (a *App) chatWithTranMailLLM(ctx context.Context, c *gin.Context, req platformAssistantRequest, lastUser string) (string, any, error) {
	if a.ai == nil || !a.ai.Configured() {
		return "", nil, fmt.Errorf("MorphAI not configured")
	}

	first := tranMailAssistantInstructions
	if hist := formatPlatformHistory(req.Messages); hist != "" {
		first += "\n\nRecent conversation:\n" + hist
	}
	if st := formatPlatformState(req.State); st != "" {
		first += "\n\nActive assistant state:\n" + st
	}
	first += "\n\nLatest user message:\n" + lastUser

	messages := []morphai.Message{{Role: "user", Content: first}}
	var lastRecord any

	for round := 0; round < tranMailToolMaxRounds; round++ {
		reply, err := a.ai.ChatCompletion(ctx, messages)
		if err != nil {
			return "", nil, err
		}
		reply = strings.TrimSpace(reply)
		if reply == "" {
			return "", nil, fmt.Errorf("empty model response")
		}

		obj, ok := morphai.ExtractJSONObject(reply)
		if !ok {
			return reply, lastRecord, nil
		}

		call, err := parseTranMailToolCall(obj)
		if err != nil {
			return reply, lastRecord, nil
		}

		result, record, execErr := a.execTranMailTool(ctx, call)
		if record != nil {
			lastRecord = record
		}
		toolMsg := "TOOL_RESULT"
		if execErr != nil {
			toolMsg += " error=" + execErr.Error()
		} else {
			toolMsg += "\n" + morphai.TruncateRunes(result, tranMailMaxToolResultRunes)
		}

		followUp := morphai.ToolFollowUpPrompt(toolMsg)
		messages = append(messages, morphai.Message{Role: "user", Content: followUp})
	}

	return "I hit the tool limit while gathering data. Please narrow your question or try again.", lastRecord, nil
}

func (a *App) respondPlatformGeneralAssistant(c *gin.Context, req platformAssistantRequest, userMsg string) {
	if a.ai != nil && a.ai.Configured() {
		reply, record, err := a.chatWithTranMailLLM(c.Request.Context(), c, req, userMsg)
		if err == nil {
			c.JSON(http.StatusOK, platformAssistantResponse{
				AssistantMessage: reply,
				Intent:           "general",
				State:            req.State,
				Completed:        true,
				Record:           record,
			})
			return
		}
	}
	c.JSON(http.StatusOK, platformAssistantResponse{
		AssistantMessage: "I can help with **ComposerX**: templates, saved content, reference docs, and writing support.\n\nTry:\n- `list templates`\n- `create template name: …`\n\nSet `MORPH_AI_API_KEY` to enable the full LLM assistant.",
		Intent:           "general",
		State:            req.State,
		Completed:        true,
	})
}

func formatPlatformHistory(messages []platformAssistantMessage) string {
	start := 0
	if len(messages) > morphai.DefaultHistoryMaxMessages {
		start = len(messages) - morphai.DefaultHistoryMaxMessages
	}
	var b strings.Builder
	for _, m := range messages[start:] {
		role := strings.TrimSpace(m.Role)
		if role != "user" && role != "assistant" {
			continue
		}
		content := morphai.TruncateHistoryContent(m.Content, morphai.DefaultHistoryMaxRunes)
		if content == "" {
			continue
		}
		b.WriteString(role)
		b.WriteString(": ")
		b.WriteString(content)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

func formatPlatformState(st platformAssistantState) string {
	if st.Intent == "" && len(st.Fields) == 0 {
		return ""
	}
	raw, _ := json.Marshal(st)
	return string(raw)
}

func parseTranMailToolCall(raw string) (*tranMailToolCall, error) {
	var call tranMailToolCall
	if err := json.Unmarshal([]byte(raw), &call); err != nil {
		return nil, err
	}
	call.Tool = strings.TrimSpace(call.Tool)
	if call.Tool == "" {
		return nil, fmt.Errorf("missing tool")
	}
	if call.Args == nil {
		call.Args = map[string]interface{}{}
	}
	return &call, nil
}

func (a *App) execTranMailTool(ctx context.Context, call *tranMailToolCall) (string, any, error) {
	switch call.Tool {
	case "list_templates":
		limit := intArg(call.Args, "limit", 10)
		offset := intArg(call.Args, "offset", 0)
		items, total, err := a.templates.List(ctx, limit, offset)
		if err != nil {
			return "", nil, err
		}
		payload := map[string]any{"items": items, "total": total, "limit": limit, "offset": offset}
		return mustJSON(payload), payload, nil

	case "get_template":
		id := int64Arg(call.Args, "id")
		if id <= 0 {
			return "", nil, fmt.Errorf("get_template requires id")
		}
		item, err := a.templates.Get(ctx, id)
		if err != nil {
			return "", nil, err
		}
		return mustJSON(item), item, nil

	case "list_saved_emails":
		limit := intArg(call.Args, "limit", 15)
		offset := intArg(call.Args, "offset", 0)
		items, total, err := a.savedEmails.List(ctx, limit, offset)
		if err != nil {
			return "", nil, err
		}
		payload := map[string]any{"items": items, "total": total, "limit": limit, "offset": offset}
		return mustJSON(payload), payload, nil

	case "get_saved_email":
		id := int64Arg(call.Args, "id")
		if id <= 0 {
			return "", nil, fmt.Errorf("get_saved_email requires id")
		}
		item, err := a.savedEmails.GetDetail(ctx, id)
		if err != nil {
			return "", nil, err
		}
		return mustJSON(item), item, nil

	case "list_reference_docs":
		docs, err := a.referenceDocs.List(ctx, 100)
		if err != nil {
			return "", nil, err
		}
		rows := make([]referenceDocListRow, 0, len(docs))
		for _, d := range docs {
			rows = append(rows, referenceDocListRow{
				ID: d.ID, Name: d.Name, Kind: d.Kind, MimeType: d.MimeType,
				ChunkCount: len(d.Chunks), CreatedAt: d.CreatedAt,
			})
		}
		payload := map[string]any{"items": rows, "total": len(rows)}
		return mustJSON(payload), payload, nil

	case "search_reference_docs":
		query := stringArg(call.Args, "query")
		if query == "" {
			return "", nil, fmt.Errorf("search_reference_docs requires query")
		}
		limit := intArg(call.Args, "limit", 10)
		if limit > 25 {
			limit = 25
		}
		docs, err := a.referenceDocs.Search(ctx, query, limit)
		if err != nil {
			return "", nil, err
		}
		rows := make([]referenceDocListRow, 0, len(docs))
		for _, d := range docs {
			rows = append(rows, referenceDocListRow{
				ID: d.ID, Name: d.Name, Kind: d.Kind, MimeType: d.MimeType,
				ChunkCount: len(d.Chunks), CreatedAt: d.CreatedAt,
			})
		}
		payload := map[string]any{"items": rows, "total": len(rows), "query": query}
		return mustJSON(payload), payload, nil

	default:
		return "", nil, fmt.Errorf("unknown tool %q", call.Tool)
	}
}

func intArg(args map[string]interface{}, key string, def int) int {
	v, ok := args[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(n))
		if err == nil {
			return i
		}
	}
	return def
}

func int64Arg(args map[string]interface{}, key string) int64 {
	v, ok := args[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(n), 10, 64)
		if err == nil {
			return i
		}
	}
	return 0
}

func stringArg(args map[string]interface{}, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	default:
		return strings.TrimSpace(fmt.Sprint(s))
	}
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func extractJSONObject(s string) (string, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") {
		return "", false
	}
	depth := 0
	for i, ch := range s {
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[:i+1], true
			}
		}
	}
	return "", false
}
