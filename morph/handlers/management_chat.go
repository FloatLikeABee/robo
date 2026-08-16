package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"idongivaflyinfa/ai"

	"github.com/gin-gonic/gin"
	gocache "github.com/patrickmn/go-cache"
	"github.com/robo/morphai"
)

const managementToolCatalog = `Tran admin (/api/tran/) — same JSON as the admin UI:
GET/POST /api/tran/districts ; GET/PUT/DELETE /api/tran/districts/:id
GET/POST /api/tran/facilities ; GET/PUT/DELETE /api/tran/facilities/:id
GET/POST /api/tran/members ; GET /api/tran/members/:id ; GET /api/tran/members/:id/full ; PUT/DELETE /api/tran/members/:id
GET/POST /api/tran/employees ; GET /api/tran/employees/:id ; GET /api/tran/employees/:id/full ; PUT/DELETE /api/tran/employees/:id
GET/POST /api/tran/assets ; GET /api/tran/assets/:id ; GET /api/tran/assets/:id/full ; PUT/DELETE /api/tran/assets/:id
GET/POST /api/tran/activities ; GET /api/tran/activities/:id ; GET /api/tran/activities/:id/full ; PUT/DELETE /api/tran/activities/:id
GET/POST /api/tran/generic-data ; GET /api/tran/generic-data/:id ; GET /api/tran/generic-data/:id/full ; POST /api/tran/generic-data/import ; POST /api/tran/generic-data/:id/analyze ; PUT/DELETE /api/tran/generic-data/:id
GET/POST /api/tran/mailing-cities ; DELETE /api/tran/mailing-cities/:id
GET/POST /api/tran/mailing-postal-codes ; DELETE /api/tran/mailing-postal-codes/:id
GET/POST /api/tran/mailing-states ; DELETE /api/tran/mailing-states/:id
GET /api/tran/mailing-reference
GET/POST /api/tran/grid-saved-filters ; DELETE /api/tran/grid-saved-filters/:id
GET /api/tran/grid-color-config ; PUT /api/tran/grid-color-config
GET /api/tran/platform-ui-config ; PUT /api/tran/platform-ui-config (display label overrides)
GET/POST /api/tran/users ; GET/PUT/DELETE /api/tran/users/:id
GET/POST /api/tran/tool-notes ; GET/PUT/DELETE /api/tran/tool-notes/:id ; POST /api/tran/tool-notes/:id/read
GET/POST /api/tran/notes-todos ; GET/PUT/DELETE /api/tran/notes-todos/:id
GET/POST /api/tran/big-notes ; GET/DELETE /api/tran/big-notes/:id ; POST /api/tran/big-notes/:id/regenerate ; POST /api/tran/big-notes/:id/publish
GET/POST /api/tran/big-notes/:id/responses ; POST /api/tran/big-notes/:id/responses/:responseId/analyze ; POST /api/tran/big-notes/:id/analyze
GET/POST /api/tran/comments ; PUT/DELETE /api/tran/comments/:id
GET/POST /api/tran/case-tasks ; GET /api/tran/case-tasks/:id ; GET /api/tran/case-tasks/:id/full ; PUT/DELETE /api/tran/case-tasks/:id
POST /api/tran/:entity/:id/attachments ; DELETE /api/tran/:entity/:id/attachments/:attachmentId ; GET /api/tran/:entity/:id/attachments/:attachmentId/download ; GET /api/tran/attachment-config
GET /api/tran/case-tasks/:id/pdf ; POST /api/tran/case-tasks/:id/send-email

Forms (/api/forms/):
GET/POST /api/forms/templates ; GET/PUT/DELETE /api/forms/templates/:id
GET/POST /api/forms/answers ; GET/PUT/DELETE /api/forms/answers/:id
GET /api/forms/assignees ; GET /api/forms/assignments ; GET /api/forms/assignments/:id ; POST /api/forms/assignments ; DELETE /api/forms/assignments/:id
GET /api/forms/notifications

SheetX Events & Info (/api/sheetx/ or /api/formsx/ — proxied to SheetX when TRANFORM_BASE_URL is set):
GET /api/sheetx/events-info ; query page=1&limit=50
GET /api/sheetx/events-info/:id
GET /api/sheetx/events-info/:id/ai-context — event plus related Contacts
POST /api/sheetx/events-info — body JSON: title (required), detail, reporter, time (RFC3339, required)
DELETE /api/sheetx/events-info/:id
GET /api/sheetx/ai/app-abilities — Morph AI form template + web search tool catalog
POST /api/sheetx/ai/web-search — body JSON: query
POST /api/sheetx/ai/form-template-chat — web-grounded HTML form template + proposed questions
PUT /api/sheetx/forms/:id — body may include landing_html to save edited template
GET /api/sheetx/survey-bot/templates ; GET /api/sheetx/survey-bot/templates/:id
GET /api/sheetx/survey-bot/results ; GET /api/sheetx/survey-bot/results/:id
POST /api/sheetx/survey-bot/templates/ai-draft — body: query / title_hint

Morph Knowledge + GraphRAG:
GET /api/knowledge/files
POST /api/knowledge/files — multipart file (md/json/csv/txt/pdf); form index_to_graph=true|false (default true)
DELETE /api/knowledge/files/:id
GET /api/graph/health
POST /api/graph/search — body JSON: query, limit

AI skills store:
GET /api/skills ; GET /api/skills/:id
POST /api/skills — body JSON: name, description, instructions|body (required), enabled?
PATCH /api/skills/:id — enable/disable or update fields
DELETE /api/skills/:id

ComposerX (/api/composerx/ — proxied when TRAN_MAIL_BASE_URL is set):
GET /api/composerx/ai/mcp-tools — email + publish page generation abilities
POST /api/composerx/ai/web-search — body JSON: query
POST /api/composerx/ai/composer-chat — web-grounded email HTML (body: messages, use_web_search, current_email_html)
POST /api/composerx/ai/publish-chat — web-grounded public page HTML (body: messages, use_web_search, theme, current_html)

Use query strings exactly as the admin app (e.g. list filters, user_id). Tran endpoints return 503 if MySQL is not configured. SheetX endpoints return 503 if TRANFORM_BASE_URL is not configured. ComposerX endpoints return 503 if TRAN_MAIL_BASE_URL is not configured.`

const managementToolInstructions = `You are Morph AI (MorphData) in the management site. You help staff in plain language.

Fast source selection:
- ` + morphai.FastToolFirstInstructions + `
- Platform knowledge / uploaded docs: POST /api/graph/search with body {"query":"…"} first when the question is about known Morph Knowledge Library content.
- MorphData operational records: use /api/tran/* and prefer /full routes for single records when available.
- SheetX form/events/template/survey-bot: use /api/sheetx/* (or legacy /api/formsx/*); survey-bot results at /api/sheetx/survey-bot/results when proxied.
- ComposerX email/publish operations: use /api/composerx/ai/mcp-tools before generation calls.
- HybridContext: users attach files and pasted notes in Morph AI chat; use excerpts when present in the conversation.

When you need to read or change data, respond with ONLY one JSON object (no markdown, no prose) with:
- "method": GET, POST, PUT, DELETE, or PATCH
- "path": must start with /api/tran/, /api/forms/, /api/formsx/, /api/sheetx/, /api/composerx/, /api/knowledge/, /api/graph/, or /api/skills
- "query": optional string WITHOUT leading ?, e.g. "school_id=1&limit=50"
- "body": optional JSON object for POST/PUT/PATCH (omit or null for GET/DELETE)

After each call you will receive a line TOOL_RESULT with HTTP status and response body. Then either explain results to the user in normal language, or issue another JSON API call if more steps are needed.

When you summarize Tran records (people: members/employees, places/facilities, districts, assets/vehicles, activities/trips, case tasks, etc.) for the user:
- For one record, use GET .../:id/full when the catalog shows a /full route (members, employees, assets, activities, case-tasks). That merges all main columns with extended data in the "detail" field.
- For districts or places (facilities), GET .../:id already includes the same kind of "detail" payload.
- Prefer user-facing terms People and Places over members/employees/facilities when summarizing.
- Use every meaningful field from the payload. The "detail" value is JSON (object or a JSON string in the body): treat it as structured fields and cover all keys inside it, not just top-level columns. If "detail" is empty or "{}", say there is no extended detail.
- Answer in readable markdown only: a one-line intro if helpful, then **Label:** value per field. Convert snake_case keys to readable labels (e.g. participant_type → Participant type). For nested maps inside "detail", add a small heading such as ### Extended detail (or ### Record detail) and the same **Label:** pattern; use bullet lists for arrays.
- Never paste the API response as a raw JSON blob or a markdown code fence unless the user explicitly asks for raw JSON.

If the user only wants general conversation or no data access, reply in normal prose only (not JSON).

Prefer fetching or listing before destructive actions. Keep paths exactly as in the catalog.

When the user asks for "all data", "everything", "all fields", "detail fields", or similarly broad coverage:
- Pull enough data to cover both top-level DB fields and extended detail fields before answering.
- Use /full record routes where available and include every meaningful field in the final answer.
- Never skip nested "detail" keys when summarizing records.

Catalog:
` + managementToolCatalog

const maxToolResultRunes = 40_000
const managementToolMaxRounds = 10
const managementQueryCacheTTL = time.Hour
const managementSessionCacheTTL = time.Hour

var managementExactQueryCache = gocache.New(managementQueryCacheTTL, 10*time.Minute)
var managementSessionSnapshotCache = gocache.New(managementSessionCacheTTL, 10*time.Minute)

type managementSessionSnapshot struct {
	LastPrompt  string   `json:"last_prompt"`
	LastReply   string   `json:"last_reply"`
	ToolRunLog  []string `json:"tool_run_log"`
	UpdatedUnix int64    `json:"updated_unix"`
}

type managementCall struct {
	Method string
	Path   string
	Query  string
	Body   []byte
}

func (h *Handlers) toolChatHistory(userID, sessionID string) string {
	msgs, err := h.db.GetChatSessionMessages(userID, sessionID)
	if err != nil || len(msgs) == 0 {
		return ""
	}
	const maxMsgs = 8
	const maxLen = 800
	start := 0
	if len(msgs) > maxMsgs {
		start = len(msgs) - maxMsgs
	}
	var b strings.Builder
	for _, m := range msgs[start:] {
		role := m.Role
		if role != "user" && role != "assistant" {
			continue
		}
		content := strings.TrimSpace(m.Content)
		if len([]rune(content)) > maxLen {
			content = truncateRunes(content, maxLen)
		}
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

func parseManagementCall(raw string) (*managementCall, error) {
	var aux struct {
		Method string          `json:"method"`
		Path   string          `json:"path"`
		Query  string          `json:"query"`
		Body   json.RawMessage `json:"body"`
		JSON   json.RawMessage `json:"json"`
	}
	if err := json.Unmarshal([]byte(raw), &aux); err != nil {
		return nil, err
	}
	if aux.Method == "" || aux.Path == "" {
		return nil, fmt.Errorf("missing method or path")
	}
	body := aux.Body
	if len(body) == 0 {
		body = aux.JSON
	}
	if len(body) > 0 && string(body) == "null" {
		body = nil
	}
	return &managementCall{Method: aux.Method, Path: aux.Path, Query: aux.Query, Body: body}, nil
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "\n…(truncated)"
}

func normalizeManagementPrompt(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

func managementExactQueryKey(userID, agentInstructions, prompt string) string {
	return fmt.Sprintf("mgmt:exact:%s:%s:%s", strings.TrimSpace(userID), normalizeManagementPrompt(agentInstructions), normalizeManagementPrompt(prompt))
}

func managementSessionSnapshotKey(userID, sessionID string) string {
	return fmt.Sprintf("mgmt:session:%s:%s", strings.TrimSpace(userID), strings.TrimSpace(sessionID))
}

func getManagementSessionSnapshot(userID, sessionID string) *managementSessionSnapshot {
	key := managementSessionSnapshotKey(userID, sessionID)
	raw, ok := managementSessionSnapshotCache.Get(key)
	if !ok {
		return nil
	}
	snap, ok := raw.(*managementSessionSnapshot)
	if !ok || snap == nil {
		return nil
	}
	return snap
}

func setManagementSessionSnapshot(userID, sessionID, prompt, reply string, toolLog []string) {
	if len(toolLog) > 24 {
		toolLog = toolLog[len(toolLog)-24:]
	}
	snap := &managementSessionSnapshot{
		LastPrompt:  strings.TrimSpace(prompt),
		LastReply:   strings.TrimSpace(reply),
		ToolRunLog:  append([]string(nil), toolLog...),
		UpdatedUnix: time.Now().Unix(),
	}
	managementSessionSnapshotCache.Set(managementSessionSnapshotKey(userID, sessionID), snap, managementSessionCacheTTL)
}

func persistManagementCachesAsync(cacheKey, userID, sessionID, prompt, reply string, toolLog []string) {
	reply = strings.TrimSpace(reply)
	if reply == "" {
		return
	}
	go func() {
		managementExactQueryCache.Set(cacheKey, reply, managementQueryCacheTTL)
		setManagementSessionSnapshot(userID, sessionID, prompt, reply, toolLog)
	}()
}

// chatWithManagementTools runs a multi-turn loop: the model may issue JSON API calls executed against the local Gin router.
// agentInstructions is optional extra persona text from a selected Morph AI agent.
// skillIDs optionally loads full skill bodies in addition to the enabled skills catalog.
func (h *Handlers) chatWithManagementTools(c *gin.Context, userID, sessionID, userPrompt, agentInstructions string, skillIDs []string) (string, error) {
	ctx := context.Background()
	cacheKey := managementExactQueryKey(userID, agentInstructions, userPrompt)
	hybridSkipCache := h.hybridStore != nil && h.hybridStore.IsAttached(userID, sessionID)
	if !hybridSkipCache {
		if cached, ok := managementExactQueryCache.Get(cacheKey); ok {
			if reply, ok := cached.(string); ok && strings.TrimSpace(reply) != "" {
				log.Printf("[MGMT-CHAT] exact-query cache hit user=%s session=%s", userID, sessionID)
				return reply, nil
			}
		}
	}

	hist := h.toolChatHistory(userID, sessionID)
	first := managementToolInstructions
	if skillsBlock := h.buildEnabledSkillsContext(skillIDs); skillsBlock != "" {
		first += "\n\n" + skillsBlock
	}
	if strings.TrimSpace(agentInstructions) != "" {
		first += "\n\n--- Selected agent focus ---\n" + strings.TrimSpace(agentInstructions)
	}
	if snap := getManagementSessionSnapshot(userID, sessionID); snap != nil {
		toolLogBlock := "(none)"
		if len(snap.ToolRunLog) > 0 {
			toolLogBlock = "- " + strings.Join(snap.ToolRunLog, "\n- ")
		}
		first += fmt.Sprintf(
			"\n\nSession cache (latest context, %s):\n- Last prompt: %s\n- Last reply: %s\n- Tool log:\n%s",
			time.Unix(snap.UpdatedUnix, 0).Format(time.RFC3339),
			truncateRunes(snap.LastPrompt, 300),
			truncateRunes(snap.LastReply, 500),
			toolLogBlock,
		)
	}
	if hist != "" {
		first += "\n\nRecent conversation:\n" + hist
	}
	first += "\n\nUser message:\n" + userPrompt

	messages := []ai.DashScopeMessage{{Role: "user", Content: first}}
	toolLog := make([]string, 0, managementToolMaxRounds)

	for round := 0; round < managementToolMaxRounds; round++ {
		reply, err := h.aiService.ChatCompletion(ctx, messages)
		if err != nil {
			return "", err
		}
		reply = strings.TrimSpace(reply)
		if reply == "" {
			return "", fmt.Errorf("empty model response")
		}

		obj, ok := morphai.ExtractJSONObject(reply)
		if !ok {
			persistManagementCachesAsync(cacheKey, userID, sessionID, userPrompt, reply, toolLog)
			return reply, nil
		}

		call, err := parseManagementCall(obj)
		if err != nil {
			persistManagementCachesAsync(cacheKey, userID, sessionID, userPrompt, reply, toolLog)
			return reply, nil
		}

		code, body := h.execManagementAPI(c, call.Method, call.Path, call.Query, call.Body)
		tb := truncateRunes(string(body), maxToolResultRunes)
		toolMsg := fmt.Sprintf("TOOL_RESULT http_status=%d\n%s", code, tb)
		toolLog = append(toolLog, fmt.Sprintf("round=%d %s %s status=%d", round+1, strings.ToUpper(call.Method), call.Path, code))
		log.Printf("[MGMT-CHAT] round=%d %s %s -> %d", round+1, call.Method, call.Path, code)

		followUp := morphai.ToolFollowUpPromptWithInstruction(toolMsg, "Reply in markdown with every meaningful field.")
		messages = append(messages, ai.DashScopeMessage{Role: "user", Content: followUp})
	}

	return "", fmt.Errorf("management tool loop exceeded %d rounds", managementToolMaxRounds)
}
