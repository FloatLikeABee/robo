package mcp_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestAgentPack(t *testing.T) {
	root := repoRoot(t)
	names := toolNames(t, filepath.Join(root, "dorph", "mcp", "server.go"))
	for extra := range toolNames(t, filepath.Join(root, "dorph", "mcp", "agent_runtime.go")) {
		names[extra] = true
	}
	for extra := range toolNames(t, filepath.Join(root, "dorph", "mcp", "pool_tools.go")) {
		names[extra] = true
	}
	for _, want := range []string{"create_note", "list_my_tasks", "get_task", "suggest_note_kind", "dorph_guide", "preview_share", "share_note", "pool_search"} {
		if !names[want] {
			t.Fatalf("mcp tools %v missing %s", names, want)
		}
	}

	skill := readPack(t, filepath.Join(root, ".cursor", "skills", "dorph-notes", "SKILL.md"))
	htmlSkill := readPack(t, filepath.Join(root, ".cursor", "skills", "html-anything", "SKILL.md"))
	doc := readPack(t, filepath.Join(root, "docs", "agents", "15-dorph-notes-self-use.md"))
	guide := readPack(t, filepath.Join(root, "docs", "agents", "16-dorph-for-agents.md"))
	createCmd := readPack(t, filepath.Join(root, ".cursor", "commands", "dorph-note.md"))
	listCmd := readPack(t, filepath.Join(root, ".cursor", "commands", "dorph-notes.md"))
	getCmd := readPack(t, filepath.Join(root, ".cursor", "commands", "dorph-note-get.md"))
	todoCmd := readPack(t, filepath.Join(root, ".cursor", "commands", "dorph-todo.md"))
	logCmd := readPack(t, filepath.Join(root, ".cursor", "commands", "dorph-log-session.md"))
	reviewCmd := readPack(t, filepath.Join(root, ".cursor", "commands", "dorph-review.md"))
	shareCmd := readPack(t, filepath.Join(root, ".cursor", "commands", "dorph-share.md"))
	poolCmd := readPack(t, filepath.Join(root, ".cursor", "commands", "dorph-pool-find.md"))
	claude := readPack(t, filepath.Join(root, "agent-pack", "dorph", "CLAUDE.md"))
	rule := readPack(t, filepath.Join(root, "agent-pack", "dorph", "cursor-rule.mdc"))

	shared := []string{
		"create_note", "list_my_tasks", "get_task",
		"title", "body", "type", "status", "limit",
		"[dorph-mcp]", "source: dorph-mcp",
		"localhost:3031",
		"Do not POST /api/tran/notes-todos",
		"not a notes user",
	}
	mustContain(t, "skill", skill, shared...)
	mustContain(t, "instruction", doc, shared...)
	mustContain(t, "skill", skill, "suggest_note_kind", "dorph_guide", "16-dorph-for-agents.md", "Never share without explicit human consent", "html-anything")
	mustContain(t, "guide", guide, "html-anything", "skills/vendor/html-anything/")
	mustContain(t, "html-anything", htmlSkill, "html-anything", "skills/vendor/html-anything/", "16-dorph-for-agents.md")
	if n := strings.Count(htmlSkill, "\n") + 1; n > 150 {
		t.Fatalf("html-anything skill is %d lines; pack files stay under 150", n)
	}
	mustContain(t, "guide", guide, "preview_share", "share_note", "unshare_note", "list_my_shares", "pool_topics", "pool_search", "pool_get", "pool_save_copy", "pool_report", "not_found", "invalid_input", "rate_limited", "forbidden_scope", "consent_required", "CC BY 4.0", "untrusted")
	if strings.Contains(guide, "coming in phase 1B") {
		t.Fatal("guide still marks pool tools as phase 1B")
	}
	mustContain(t, "dorph-note", createCmd, "create_note")
	mustContain(t, "dorph-notes", listCmd, "list_my_tasks")
	mustContain(t, "dorph-note-get", getCmd, "get_task")
	mustContain(t, "dorph-todo", todoCmd, "suggest_note_kind", "16-dorph-for-agents.md")
	mustContain(t, "dorph-log-session", logCmd, "create_note", "16-dorph-for-agents.md")
	mustContain(t, "dorph-review", reviewCmd, "list_my_tasks", "16-dorph-for-agents.md")
	mustContain(t, "dorph-share", shareCmd, "Never share without explicit human consent", "preview_share", "yes")
	mustContain(t, "dorph-pool-find", poolCmd, "pool_search", "untrusted", "16-dorph-for-agents.md")
	mustContain(t, "claude", claude, "suggest_note_kind", "16-dorph-for-agents.md")
	mustContain(t, "rule", rule, "suggest_note_kind", "16-dorph-for-agents.md")

	for name, body := range map[string]string{
		"skill": skill, "instruction": doc,
		"dorph-note": createCmd, "dorph-notes": listCmd, "dorph-note-get": getCmd,
		"dorph-todo": todoCmd, "dorph-log-session": logCmd, "dorph-review": reviewCmd,
		"dorph-share": shareCmd, "dorph-pool-find": poolCmd,
		"claude": claude, "rule": rule, "guide": guide,
	} {
		mustContain(t, name, body,
			"Do not POST /api/tran/notes-todos",
		)
		if name != "guide" && name != "claude" && name != "rule" {
			mustContain(t, name, body,
				"Do not click Notes & TODOs",
				"dorph-mcp is not connected",
			)
		}
		if strings.Contains(body, "eyJ") {
			t.Errorf("%s contains a JWT prefix", name)
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "docs", "agents", "14-dorph-mcp.md")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}

func toolNames(t *testing.T, path string) map[string]bool {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]bool{}
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`readOnly\(\s*"([^"]+)"`),
		regexp.MustCompile(`readOnlyTool\(\s*"([^"]+)"`),
		regexp.MustCompile(`writeTool\(\s*"([^"]+)"`),
		regexp.MustCompile(`Name:\s*"([^"]+)"`),
	} {
		for _, m := range re.FindAllStringSubmatch(string(b), -1) {
			out[m[1]] = true
		}
	}
	if len(out) == 0 {
		t.Fatal("no tool names extracted from server.go")
	}
	return out
}

func readPack(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func mustContain(t *testing.T, name, body string, phrases ...string) {
	t.Helper()
	for _, phrase := range phrases {
		if !strings.Contains(body, phrase) {
			t.Errorf("%s missing %q", name, phrase)
		}
	}
}
