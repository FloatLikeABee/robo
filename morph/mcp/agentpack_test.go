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
	names := toolNames(t, filepath.Join(root, "morph", "mcp", "server.go"))
	for _, want := range []string{"create_note", "list_my_tasks", "get_task"} {
		if !names[want] {
			t.Fatalf("server.go tools %v missing %s", names, want)
		}
	}

	skill := readPack(t, filepath.Join(root, ".cursor", "skills", "morph-notes", "SKILL.md"))
	doc := readPack(t, filepath.Join(root, "docs", "agents", "15-morph-notes-self-use.md"))
	createCmd := readPack(t, filepath.Join(root, ".cursor", "commands", "morph-note.md"))
	listCmd := readPack(t, filepath.Join(root, ".cursor", "commands", "morph-notes.md"))
	getCmd := readPack(t, filepath.Join(root, ".cursor", "commands", "morph-note-get.md"))

	shared := []string{
		"create_note", "list_my_tasks", "get_task",
		"title", "body", "type", "status", "limit",
		"[morph-mcp]", "source: morph-mcp",
		"localhost:3031",
		"Do not POST /api/tran/notes-todos",
		"not a notes user",
	}
	mustContain(t, "skill", skill, shared...)
	mustContain(t, "instruction", doc, shared...)
	mustContain(t, "morph-note", createCmd, "create_note")
	mustContain(t, "morph-notes", listCmd, "list_my_tasks")
	mustContain(t, "morph-note-get", getCmd, "get_task")

	for name, body := range map[string]string{
		"skill": skill, "instruction": doc,
		"morph-note": createCmd, "morph-notes": listCmd, "morph-note-get": getCmd,
	} {
		mustContain(t, name, body,
			"Do not POST /api/tran/notes-todos",
			"Do not click Notes & TODOs",
			"morph-mcp is not connected",
		)
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
		if _, err := os.Stat(filepath.Join(dir, "docs", "agents", "14-morph-mcp.md")); err == nil {
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
