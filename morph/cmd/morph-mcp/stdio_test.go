package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"idongivaflyinfa/auth"
)

func TestMain(m *testing.M) {
	if os.Getenv("MORPH_MCP_STDIO_CHILD") == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestStdioHandshakeAndWhoami(t *testing.T) {
	secret := "stdio-test-secret"
	cfg := auth.TokenConfig{Secret: []byte(secret), ExpiryHours: 24}
	tok, err := auth.EncodeToken(cfg, "user-stdio", "ada@example.com", "ada", []string{"Admin"}, "ch")
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Dir = t.TempDir()
	cmd.Env = childEnv(
		"MORPH_MCP_STDIO_CHILD=1",
		"JWT_SECRET="+secret,
		"MORPH_MCP_TOKEN="+tok,
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	lines := startLineReader(stdout)
	if err := writeLine(stdin, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"stdio-test","version":"0"}}}`); err != nil {
		t.Fatal(err)
	}
	initLine := waitID(t, lines, "1")
	assertJSONRPC(t, initLine, tok)
	initResult := resultObject(t, initLine)
	if initResult["protocolVersion"] != "2025-06-18" {
		t.Fatalf("protocolVersion = %#v", initResult["protocolVersion"])
	}
	info, _ := initResult["serverInfo"].(map[string]any)
	if info["name"] != "morph-mcp" || info["version"] == "" {
		t.Fatalf("serverInfo = %#v", initResult["serverInfo"])
	}
	caps, _ := initResult["capabilities"].(map[string]any)
	if caps["tools"] == nil || caps["resources"] == nil {
		t.Fatalf("capabilities = %#v", caps)
	}
	resourcesCap, _ := caps["resources"].(map[string]any)
	if resourcesCap == nil {
		t.Fatalf("capabilities.resources = %#v", caps["resources"])
	}
	if changed, ok := resourcesCap["listChanged"]; ok {
		t.Fatalf("resources.listChanged = %#v; the server does not emit list_changed", changed)
	}
	if err := assertNotListening(t, cmd.Process.Pid); err != nil {
		t.Fatal(err)
	}

	if err := writeLine(stdin, `{"jsonrpc":"2.0","method":"notifications/initialized"}`); err != nil {
		t.Fatal(err)
	}
	if err := writeLine(stdin, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`); err != nil {
		t.Fatal(err)
	}
	listLine := waitID(t, lines, "2")
	assertJSONRPC(t, listLine, tok)
	tools := resultObject(t, listLine)["tools"].([]any)
	if len(tools) != 1 || tools[0].(map[string]any)["name"] != "whoami" {
		t.Fatalf("tools = %#v", tools)
	}

	if err := writeLine(stdin, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"whoami","arguments":{}}}`); err != nil {
		t.Fatal(err)
	}
	callLine := waitID(t, lines, "3")
	assertJSONRPC(t, callLine, tok)
	if !strings.Contains(callLine, "user-stdio") {
		t.Fatalf("whoami result = %s", callLine)
	}
	if strings.Contains(initLine+listLine+callLine, "server ready") {
		t.Fatal("stdout included a log line")
	}

	if err := stdin.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-waitDone:
		if err != nil {
			t.Fatalf("exit: %v\nstderr: %s", err, stderr.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("process did not exit after stdin closed")
	}

	// cmd.Wait returns only after the stderr copy goroutine finishes.
	// bytes.Buffer is not safe to read while that goroutine is still writing.
	stderrText := stderr.String()
	if strings.Contains(stderrText, tok) {
		t.Fatal("stderr included the token")
	}
	if !strings.Contains(stderrText, "morph-mcp server ready") {
		t.Fatalf("stderr = %q", stderrText)
	}
	if strings.Contains(stderrText, "user-stdio") || strings.Contains(stderrText, "user_id") {
		t.Fatalf("stderr included the user id: %q", stderrText)
	}
	drainLines(t, lines, tok)
}

func TestStdioRejectsMissingIdentity(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Dir = t.TempDir()
	cmd.Env = childEnv("MORPH_MCP_STDIO_CHILD=1", "MORPH_MCP_TOKEN=")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "MORPH_MCP_TOKEN") || !strings.Contains(stderr.String(), "required") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestStdioRejectsInvalidIdentity(t *testing.T) {
	const bogus = "not-a-jwt"
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Dir = t.TempDir()
	cmd.Env = childEnv(
		"MORPH_MCP_STDIO_CHILD=1",
		"JWT_SECRET=stdio-test-secret",
		"MORPH_MCP_TOKEN="+bogus,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	text := stderr.String()
	if strings.Contains(text, bogus) {
		t.Fatal("stderr included the token")
	}
	if !strings.Contains(text, "invalid") {
		t.Fatalf("stderr = %q", text)
	}
}

func childEnv(extra ...string) []string {
	drop := map[string]bool{
		"MORPH_MCP_TOKEN":       true,
		"JWT_SECRET":            true,
		"MORPH_MCP_STDIO_CHILD": true,
	}
	var out []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if drop[key] {
			continue
		}
		out = append(out, entry)
	}
	return append(out, extra...)
}

func writeLine(w io.Writer, line string) error {
	_, err := io.WriteString(w, line+"\n")
	return err
}

func startLineReader(r io.Reader) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
		for sc.Scan() {
			ch <- sc.Text()
		}
	}()
	return ch
}

func drainLines(t *testing.T, lines <-chan string, token string) {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			assertJSONRPC(t, line, token)
			if strings.Contains(line, "server ready") {
				t.Fatalf("stdout included a log line: %s", line)
			}
		case <-deadline:
			t.Fatal("stdout did not reach EOF")
		}
	}
}

func waitID(t *testing.T, lines <-chan string, id string) string {
	t.Helper()
	deadline := time.After(10 * time.Second)
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				t.Fatal("stdout closed before a response")
			}
			var msg map[string]any
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				t.Fatalf("stdout line is not JSON: %s", line)
			}
			if rpcID(msg["id"]) == id {
				return line
			}
		case <-deadline:
			t.Fatalf("timed out waiting for id %s", id)
		}
	}
}

func assertJSONRPC(t *testing.T, line, token string) {
	t.Helper()
	var msg map[string]any
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		t.Fatal(err)
	}
	if msg["jsonrpc"] != "2.0" {
		t.Fatalf("line = %s", line)
	}
	if token != "" && strings.Contains(line, token) {
		t.Fatal("stdout included the token")
	}
}

func resultObject(t *testing.T, line string) map[string]any {
	t.Helper()
	var msg map[string]any
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		t.Fatal(err)
	}
	result, ok := msg["result"].(map[string]any)
	if !ok {
		t.Fatalf("result = %#v", msg["result"])
	}
	return result
}

func assertNotListening(t *testing.T, pid int) error {
	t.Helper()
	inodes := map[string]struct{}{}
	fdDir := fmt.Sprintf("/proc/%d/fd", pid)
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		target, err := os.Readlink(filepath.Join(fdDir, entry.Name()))
		if err != nil || !strings.HasPrefix(target, "socket:[") {
			continue
		}
		inode := strings.TrimSuffix(strings.TrimPrefix(target, "socket:["), "]")
		inodes[inode] = struct{}{}
	}
	for _, name := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		body, err := os.ReadFile(name)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(body), "\n")[1:] {
			fields := strings.Fields(line)
			if len(fields) < 10 {
				continue
			}
			if !strings.EqualFold(fields[3], "0A") {
				continue
			}
			if _, ok := inodes[fields[9]]; ok {
				return fmt.Errorf("pid %d is listening on %s", pid, fields[1])
			}
		}
	}
	return nil
}

func rpcID(v any) string {
	switch id := v.(type) {
	case float64:
		return strconv.FormatInt(int64(id), 10)
	case string:
		return id
	default:
		return ""
	}
}
