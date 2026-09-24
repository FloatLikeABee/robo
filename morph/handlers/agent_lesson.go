package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"idongivaflyinfa/ai"
	"idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robo/morphai"
)

const agentLessonPromptCap = 8

func isLowContextGreeting(s string) bool {
	low := strings.ToLower(strings.TrimSpace(s))
	switch low {
	case "hi", "hello", "hey", "thanks", "thank you", "ok", "okay", "yes", "no":
		return true
	}
	if low == "" {
		return true
	}
	fields := strings.Fields(low)
	return utf8.RuneCountInString(low) <= 24 && !strings.Contains(low, "?") && len(fields) <= 4
}

func sessionIsSignificant(userTurns, toolRounds int, hasDocs bool) bool {
	return userTurns >= 6 || toolRounds >= 2 || hasDocs
}

func countUserTurns(h *Handlers, userID, sessionID string) int {
	if h == nil || h.db == nil {
		return 0
	}
	msgs, err := h.db.GetChatSessionMessages(userID, sessionID)
	if err != nil {
		return 0
	}
	n := 0
	for _, m := range msgs {
		if m.Role == "user" {
			n++
		}
	}
	return n
}

// buildAgentLessonsContext injects enabled lessons for a verified bearer user.
// A missing or untrusted identity (including a client X-User-ID) injects nothing.
func (h *Handlers) buildAgentLessonsContext(c *gin.Context) string {
	userID, ok := h.trustedLessonUserID(c)
	if !ok {
		return ""
	}
	return h.lessonsPromptForUser(userID)
}

func (h *Handlers) lessonsPromptForUser(userID string) string {
	if h == nil || h.TranMySQL == nil {
		return ""
	}
	rows, err := h.TranMySQL.ListAgentLessons(context.Background(), userID, true, agentLessonPromptCap)
	if err != nil || len(rows) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("--- Learned from sessions ---\n")
	b.WriteString("Follow these when they match the current request:\n")
	for _, l := range rows {
		b.WriteString(fmt.Sprintf("- When %s: %s\n", strings.TrimSpace(l.Trigger), strings.TrimSpace(l.Rule)))
	}
	return strings.TrimSpace(b.String())
}

func (h *Handlers) maybeHarvestSession(userID, sessionID, lastUserPrompt string, userTurns, toolRounds int, hasDocs bool) {
	if h == nil || h.TranMySQL == nil {
		return
	}
	userID = strings.TrimSpace(userID)
	sessionID = strings.TrimSpace(sessionID)
	if userID == "" || sessionID == "" {
		return
	}
	if isLowContextGreeting(lastUserPrompt) && !sessionIsSignificant(userTurns, toolRounds, hasDocs) {
		return
	}
	if !sessionIsSignificant(userTurns, toolRounds, hasDocs) {
		return
	}
	ctx := context.Background()
	existing, err := h.TranMySQL.GetAgentLessonBySession(ctx, userID, sessionID)
	if err != nil || existing != nil {
		return
	}
	transcript := strings.TrimSpace(lastUserPrompt)
	if h.db != nil {
		if hist := h.toolChatHistory(userID, sessionID); hist != "" {
			transcript = hist
		}
	}
	run := func() {
		trigger, rule, err := h.distillSessionLesson(ctx, transcript)
		if err != nil || trigger == "" || rule == "" {
			return
		}
		_ = h.TranMySQL.InsertAgentLesson(ctx, &db.AgentLesson{
			ID:              uuid.NewString(),
			Trigger:         trigger,
			Rule:            rule,
			SourceSessionID: sessionID,
			CreatedAt:       time.Now().UTC().Format(time.RFC3339),
			Enabled:         true,
			OwnerUserID:     userID,
		})
	}
	if h.distillLesson != nil {
		run()
		return
	}
	go run()
}

func (h *Handlers) distillSessionLesson(ctx context.Context, transcript string) (trigger, rule string, err error) {
	transcript = strings.TrimSpace(transcript)
	if transcript == "" {
		return "", "", fmt.Errorf("empty transcript")
	}
	if h.distillLesson != nil {
		return h.distillLesson(ctx, transcript)
	}
	if h.aiService == nil {
		return "", "", fmt.Errorf("AI service not configured")
	}
	raw, err := h.aiService.ChatCompletion(ctx, []ai.DashScopeMessage{
		{Role: "user", Content: agentLessonDistillPrompt(transcript)},
	})
	if err != nil {
		return "", "", err
	}
	return parseAgentLessonJSON(raw)
}

func agentLessonDistillPrompt(transcript string) string {
	return strings.TrimSpace(fmt.Sprintf(`Distill one reusable lesson from this Morph AI session. Return ONLY a JSON object with keys trigger and rule (short strings). Skip greetings and trivia. trigger is when the lesson applies; rule is what to do.

Transcript:
%s
`, truncateRunes(transcript, 6000)))
}

func parseAgentLessonJSON(raw string) (trigger, rule string, err error) {
	obj, ok := morphai.ExtractJSONObject(raw)
	if !ok {
		return "", "", fmt.Errorf("AI did not return JSON")
	}
	var parsed struct {
		Trigger string `json:"trigger"`
		Rule    string `json:"rule"`
	}
	if err := json.Unmarshal([]byte(obj), &parsed); err != nil {
		return "", "", fmt.Errorf("AI JSON was invalid")
	}
	trigger = strings.TrimSpace(parsed.Trigger)
	rule = strings.TrimSpace(parsed.Rule)
	if trigger == "" || rule == "" {
		return "", "", fmt.Errorf("AI draft is missing trigger or rule")
	}
	return trigger, rule, nil
}
