package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"idongivaflyinfa/ai"

	"github.com/gin-gonic/gin"
	"github.com/robo/dorphai"
)

type coordinatorTurn struct {
	Plan        string   `json:"plan"`
	MemoryPatch string   `json:"memory_patch"`
	SubAgents   []string `json:"sub_agents"`
}

func (h *Handlers) runCoordinatorTurn(c *gin.Context, userID, sessionID, userMessage string) (coordinatorTurn, error) {
	var out coordinatorTurn
	if h == nil || h.aiService == nil || !h.aiService.CoordinatorConfigured() {
		return out, nil
	}
	if isSimpleChat(userMessage) {
		return out, nil
	}
	ctx := context.Background()
	prompt := coordinatorPrompt(userMessage)
	raw, err := h.aiService.CoordinatorChatCompletion(ctx, []ai.DashScopeMessage{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return out, err
	}
	parsed, ok := parseCoordinatorJSON(raw)
	if !ok {
		return out, nil
	}
	out = parsed
	if lessonUser, ok := h.trustedLessonUserID(c); ok && h.TranMySQL != nil {
		_ = h.TranMySQL.PatchAgentWorkingMemory(ctx, lessonUser, sessionID, out.MemoryPatch)
	}
	return out, nil
}

func coordinatorPrompt(userMessage string) string {
	return strings.TrimSpace(fmt.Sprintf(`You are the coordinator sub-agent (coord_v1). Do NOT answer the user. Return ONLY one JSON object with optional keys:
- plan: one short advisory paragraph for the primary agent
- memory_patch: facts/decisions to remember this session (plain text, under 2000 chars)
- sub_agents: array of zero to three ids from: files, knowledge, notes, dorphdata

User message:
%s`, truncateRunes(userMessage, 4000)))
}

func parseCoordinatorJSON(raw string) (coordinatorTurn, bool) {
	obj, ok := dorphai.ExtractJSONObject(raw)
	if !ok {
		return coordinatorTurn{}, false
	}
	var t coordinatorTurn
	if err := json.Unmarshal([]byte(obj), &t); err != nil {
		return coordinatorTurn{}, false
	}
	t.Plan = strings.TrimSpace(t.Plan)
	t.MemoryPatch = strings.TrimSpace(t.MemoryPatch)
	if len(t.SubAgents) > maxSubAgents {
		t.SubAgents = t.SubAgents[:maxSubAgents]
	}
	return t, true
}

func intersectSubAgents(keyword, coordinator []string) []string {
	if len(coordinator) == 0 {
		return keyword
	}
	allowed := map[string]struct{}{"files": {}, "knowledge": {}, "notes": {}, "dorphdata": {}}
	var out []string
	for _, id := range coordinator {
		id = strings.ToLower(strings.TrimSpace(id))
		if _, ok := allowed[id]; !ok {
			continue
		}
		out = append(out, id)
	}
	if len(keyword) == 0 {
		return out
	}
	kset := map[string]struct{}{}
	for _, w := range keyword {
		kset[strings.ToLower(w)] = struct{}{}
	}
	var both []string
	for _, id := range out {
		if _, ok := kset[id]; ok {
			both = append(both, id)
		}
	}
	if len(both) > 0 {
		return both
	}
	return out
}

func (h *Handlers) buildAgentMemoryPromptBlock(c *gin.Context, userID, sessionID, advisoryPlan string) string {
	var b strings.Builder
	if lessonUser, ok := h.trustedLessonUserID(c); ok && h.TranMySQL != nil {
		if summary, _ := h.TranMySQL.GetAgentWorkingMemory(context.Background(), lessonUser, sessionID); summary != "" {
			b.WriteString("--- Session working memory ---\n")
			b.WriteString(summary)
			b.WriteString("\n")
		}
	}
	if plan := strings.TrimSpace(advisoryPlan); plan != "" {
		b.WriteString("--- Coordinator plan (advisory) ---\n")
		b.WriteString(plan)
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}
