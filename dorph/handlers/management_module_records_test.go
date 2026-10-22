package handlers

import (
	"strings"
	"testing"

	"github.com/robo/dorphai"
)

func TestManagementChatLeadStatesReplyShapeAndIdentity(t *testing.T) {
	text := managementChatLead(dorphai.ModelIdentityInstructions("ollama", "llama3.2"))
	for _, needle := range []string{
		"ordinary question",
		"provider ollama",
		"model llama3.2",
		"only for a real data read or change",
		"GET /api/tran/case-tasks",
	} {
		if !strings.Contains(text, needle) {
			t.Errorf("lead missing %q", needle)
		}
	}
	if strings.Contains(text, "Visual-first") || strings.Contains(text, "short caption") || strings.Contains(text, "5.1") {
		t.Fatalf("diagram-first or invented version leaked:\n%s", text)
	}
}

func TestModuleRecordInstructionsNameRoutes(t *testing.T) {
	text := managementToolInstructions
	for _, needle := range []string{
		"GET /api/tran/case-tasks",
		"GET /api/tran/case-tasks/:id/full",
		"POST /api/tran/case-tasks",
		"GET /api/tran/timelines",
		"GET /api/tran/timelines/:id",
		"POST /api/tran/timelines",
		"GET /api/tran/big-notes",
		"GET /api/tran/big-notes/:id",
		"POST /api/tran/big-notes",
		"GET /api/tran/research",
		"GET /api/tran/research/:id",
		"POST /api/tran/research",
		"GET /api/sheetx/events-info",
		"GET /api/sheetx/events-info/:id",
		"POST /api/sheetx/events-info",
		"do not call POST /api/tran/notes-todos for those requests",
		"Do not file stick notes, timelines, stories, research, or event logs as Notes & TODOs",
	} {
		if !strings.Contains(text, needle) {
			t.Errorf("instructions missing %q", needle)
		}
	}
}
