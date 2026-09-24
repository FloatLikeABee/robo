package handlers

import (
	"context"
	"strings"
	"testing"

	morphdb "idongivaflyinfa/db"
)

func TestSessionIsSignificantAndGreetingSkip(t *testing.T) {
	if sessionIsSignificant(1, 0, false) {
		t.Fatal("greeting-scale session is not significant")
	}
	if !isLowContextGreeting("hi") || !isLowContextGreeting("hello") {
		t.Fatal("short greetings should skip")
	}
	if !sessionIsSignificant(6, 0, false) || !sessionIsSignificant(1, 2, false) || !sessionIsSignificant(1, 0, true) {
		t.Fatal("turns, tools, or docs should be significant")
	}
}

func TestMaybeHarvestSkipsGreeting(t *testing.T) {
	sqlDB := openHarnessSQL(t)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	called := false
	h.distillLesson = func(ctx context.Context, transcript string) (string, string, error) {
		called = true
		return "t", "r", nil
	}
	h.maybeHarvestSession("u", "sess-hi", "hi", 1, 0, false)
	if called {
		t.Fatal("greeting must not distill")
	}
	rows, err := h.TranMySQL.ListAgentLessons(context.Background(), "u", false, 8)
	if err != nil || len(rows) != 0 {
		t.Fatalf("expected no lessons, got %d err=%v", len(rows), err)
	}
}

func TestMaybeHarvestStoresOnceAndLaterContextIncludesRule(t *testing.T) {
	sqlDB := openHarnessSQL(t)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	h.distillLesson = func(ctx context.Context, transcript string) (string, string, error) {
		return "operator asks about pinned docs", "search the graph before answering", nil
	}
	h.maybeHarvestSession("u", "sess-docs", "What does the spec say about auth?", 1, 0, true)
	rows, err := h.TranMySQL.ListAgentLessons(context.Background(), "u", false, 8)
	if err != nil || len(rows) != 1 {
		t.Fatalf("expected one lesson, got %d err=%v", len(rows), err)
	}
	if rows[0].SourceSessionID != "sess-docs" {
		t.Fatalf("source session: %s", rows[0].SourceSessionID)
	}
	if rows[0].OwnerUserID != "u" || !rows[0].Enabled {
		t.Fatalf("harvest must record owner and enabled: %+v", rows[0])
	}
	h.maybeHarvestSession("u", "sess-docs", "follow up with more detail please", 6, 2, true)
	rows, err = h.TranMySQL.ListAgentLessons(context.Background(), "u", false, 8)
	if err != nil || len(rows) != 1 {
		t.Fatalf("duplicate harvest: %d err=%v", len(rows), err)
	}
	ctxBlock := h.buildAgentLessonsContext("u")
	if !strings.Contains(ctxBlock, "search the graph before answering") {
		t.Fatalf("later context missing rule:\n%s", ctxBlock)
	}
	if h.buildAgentLessonsContext("someone-else") != "" {
		t.Fatal("another user must not receive this lesson")
	}
	combined := h.agentSkillsAndLessonsContext("u", nil)
	if !strings.Contains(combined, "search the graph before answering") {
		t.Fatalf("skills+lessons missing rule:\n%s", combined)
	}
}
