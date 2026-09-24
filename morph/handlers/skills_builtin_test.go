package handlers

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	morphdb "idongivaflyinfa/db"

	_ "modernc.org/sqlite"
)

func openHarnessSQL(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:harness-"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	stmts := []string{
		`CREATE TABLE ai_skills (
			id TEXT NOT NULL PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			enabled INTEGER NOT NULL DEFAULT 1,
			owner_user_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE agent_lesson (
			id TEXT NOT NULL PRIMARY KEY,
			trigger TEXT NOT NULL,
			rule TEXT NOT NULL,
			source_session_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX idx_agent_lesson_session ON agent_lesson(source_session_id)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestSeedBuiltinSkillsInsertsMissingOnExistingStore(t *testing.T) {
	sqlDB := openHarnessSQL(t)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	ctx := context.Background()
	if err := h.TranMySQL.InsertAISkill(ctx, &morphdb.AISkill{
		ID: "operator-custom", Name: "Custom", Description: "mine",
		Enabled: true, OwnerUserID: "op", CreatedAt: "t", UpdatedAt: "t",
	}); err != nil {
		t.Fatal(err)
	}
	h.SeedBuiltinSkills()
	got, err := h.TranMySQL.GetAISkill(ctx, builtinResearchID)
	if err != nil || got == nil {
		t.Fatalf("research builtin missing: %v %#v", err, got)
	}
	design, err := h.TranMySQL.GetAISkill(ctx, builtinDesignID)
	if err != nil || design == nil {
		t.Fatalf("design builtin missing: %v %#v", err, design)
	}
	graphs, err := h.TranMySQL.GetAISkill(ctx, builtinGraphsID)
	if err != nil || graphs == nil {
		t.Fatalf("graphs builtin missing: %v %#v", err, graphs)
	}
	custom, err := h.TranMySQL.GetAISkill(ctx, "operator-custom")
	if err != nil || custom == nil {
		t.Fatal("operator skill must stay")
	}
	h.SeedBuiltinSkills()
	n, err := h.TranMySQL.CountAISkills(ctx)
	if err != nil || n != 7 {
		t.Fatalf("second seed must not duplicate, got n=%d err=%v", n, err)
	}
}

func TestBuildEnabledSkillsContextIncludesDefaultBodiesWithoutPicker(t *testing.T) {
	h := &Handlers{}
	got := h.buildEnabledSkillsContext(nil)
	if !strings.Contains(got, builtinResearchInstructions) {
		t.Fatalf("research body missing:\n%s", got)
	}
	if !strings.Contains(got, builtinDesignInstructions) {
		t.Fatalf("design body missing:\n%s", got)
	}
	if !strings.Contains(got, builtinGraphsInstructions) {
		t.Fatalf("graphs body missing:\n%s", got)
	}
}
