package handlers

import "strings"
import "testing"

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
