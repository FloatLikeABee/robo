package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"idongivaflyinfa/auth"
	"idongivaflyinfa/mcp"
)

func TestHandshakeListAndWhoami(t *testing.T) {
	cfg := testTokenConfig()
	tok := signToken(t, cfg, "user-1", "ada@example.com", "ada", []string{"Admin"})
	id, err := mcp.ResolveIdentity(tok, cfg)
	if err != nil {
		t.Fatal(err)
	}

	server, cleanup, err := mcp.NewServer(id, nil, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	session := connectInMemory(t, server, "2025-06-18")

	init := session.InitializeResult()
	if init == nil {
		t.Fatal("missing initialize result")
	}
	if init.ProtocolVersion != "2025-06-18" {
		t.Fatalf("protocolVersion = %q", init.ProtocolVersion)
	}
	if init.ServerInfo == nil || init.ServerInfo.Name != mcp.ServerName || init.ServerInfo.Version == "" {
		t.Fatalf("serverInfo = %+v", init.ServerInfo)
	}
	if init.Capabilities == nil || init.Capabilities.Tools == nil || init.Capabilities.Resources == nil {
		t.Fatalf("capabilities = %+v", init.Capabilities)
	}
	if init.Capabilities.Resources.ListChanged {
		t.Fatal("resources listChanged must be false; the server does not emit list_changed")
	}
	if init.Capabilities.Prompts != nil || init.Capabilities.Logging != nil {
		t.Fatalf("unexpected capabilities = %+v", init.Capabilities)
	}
	if !strings.Contains(strings.ToLower(init.Instructions), "read-only") {
		t.Fatalf("instructions = %q", init.Instructions)
	}

	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tools) != 4 || !toolNamed(listed.Tools, "whoami") || !toolNamed(listed.Tools, "create_note") {
		t.Fatalf("tools = %+v", listed.Tools)
	}
	who := toolByName(t, listed.Tools, "whoami")
	if who.Annotations == nil || !who.Annotations.ReadOnlyHint {
		t.Fatalf("whoami annotations = %+v", who.Annotations)
	}

	res, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: "whoami"})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("whoami error result: %+v", res)
	}
	got := structuredMap(t, res.StructuredContent)
	if got["id"] != "user-1" || got["email"] != "ada@example.com" || got["username"] != "ada" {
		t.Fatalf("whoami = %#v", got)
	}
	roles, ok := got["roles"].([]any)
	if !ok || len(roles) != 1 || roles[0] != "Admin" {
		t.Fatalf("roles = %#v", got["roles"])
	}
	if strings.Contains(toolText(t, res), tok) {
		t.Fatal("tool result included the token")
	}

	resources, err := session.ListResources(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(resources.Resources) != 0 {
		t.Fatalf("resources = %+v", resources.Resources)
	}
}

func TestHandshakeNegotiatesNewerProtocol(t *testing.T) {
	id, err := mcp.ResolveIdentity(signToken(t, testTokenConfig(), "user-1", "ada@example.com", "ada", nil), testTokenConfig())
	if err != nil {
		t.Fatal(err)
	}
	server, cleanup, err := mcp.NewServer(id, nil, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	session := connectInMemory(t, server, "2026-07-28")
	init := session.InitializeResult()
	if init == nil || init.ProtocolVersion != "2026-07-28" {
		t.Fatalf("initialize = %+v", init)
	}
	if init.Capabilities == nil || init.Capabilities.Tools == nil {
		t.Fatalf("capabilities = %+v", init.Capabilities)
	}
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tools) != 4 || !toolNamed(listed.Tools, "whoami") || !toolNamed(listed.Tools, "create_note") {
		t.Fatalf("tools = %+v", listed.Tools)
	}
}

func toolNamed(tools []*sdkmcp.Tool, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func toolByName(t *testing.T, tools []*sdkmcp.Tool, name string) *sdkmcp.Tool {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("missing tool %s", name)
	return nil
}

func TestToolsAreWhoamiListAndGet(t *testing.T) {
	cfg := testTokenConfig()
	tok := signToken(t, cfg, "user-1", "ada@example.com", "ada", nil)
	id, err := mcp.ResolveIdentity(tok, cfg)
	if err != nil {
		t.Fatal(err)
	}
	server, cleanup, err := mcp.NewServer(id, nil, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	session := connectInMemory(t, server, "2025-06-18")
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, tool := range listed.Tools {
		if tool.Annotations == nil {
			t.Fatalf("%s annotations = nil", tool.Name)
		}
		if tool.Name == "create_note" {
			if tool.Annotations.ReadOnlyHint {
				t.Fatal("create_note is marked read-only")
			}
		} else if !tool.Annotations.ReadOnlyHint {
			t.Fatalf("%s annotations = %+v", tool.Name, tool.Annotations)
		}
		got[tool.Name] = true
	}
	for _, name := range []string{"whoami", "list_my_tasks", "get_task", "create_note"} {
		if !got[name] {
			t.Fatalf("missing %s in %+v", name, listed.Tools)
		}
	}
	if len(listed.Tools) != 4 {
		t.Fatalf("tools = %+v", listed.Tools)
	}
}

func TestExpiredTokenIsAToolError(t *testing.T) {
	secret := "stdio-test-secret"
	cfg := auth.TokenConfig{Secret: []byte(secret), ExpiryHours: 24}
	tok := signToken(t, cfg, "user-1", "ada@example.com", "ada", nil)
	t.Setenv("JWT_SECRET", secret)
	t.Setenv(mcp.TokenEnv, tok)
	id, err := mcp.ResolveIdentity(tok, cfg)
	if err != nil {
		t.Fatal(err)
	}
	recheck := func() error {
		return mcp.RecheckToken(os.Getenv(mcp.TokenEnv), auth.LoadTokenConfig())
	}
	server, cleanup, err := mcp.NewServer(id, nil, recheck, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	session := connectInMemory(t, server, "2025-06-18")
	ctx := context.Background()
	res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "whoami"})
	if err != nil || res.IsError {
		t.Fatalf("whoami err=%v res=%+v", err, res)
	}
	expired, err := auth.EncodeToken(auth.TokenConfig{Secret: []byte(secret), ExpiryHours: -1}, "user-1", "ada@example.com", "ada", nil, "ch")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(mcp.TokenEnv, expired)
	for _, name := range []string{"whoami", "list_my_tasks"} {
		res, err = session.CallTool(ctx, &sdkmcp.CallToolParams{Name: name})
		if err != nil {
			t.Fatalf("%s protocol error: %v", name, err)
		}
		if !res.IsError {
			t.Fatalf("%s succeeded with an expired token", name)
		}
		if strings.Contains(toolText(t, res), expired) {
			t.Fatalf("%s leaked the token", name)
		}
	}
}

func TestCreateNoteToolRoundTripAndFailClosed(t *testing.T) {
	path, writer := seedTwoUsers(t)
	ctx := context.Background()
	secret := "stdio-test-secret"
	cfg := auth.TokenConfig{Secret: []byte(secret), ExpiryHours: 24}
	tok := signToken(t, cfg, "ada-id", "ada@example.com", "ada", nil)
	t.Setenv("JWT_SECRET", secret)
	t.Setenv(mcp.TokenEnv, tok)
	id, err := mcp.ResolveIdentity(tok, cfg)
	if err != nil {
		t.Fatal(err)
	}
	recheck := func() error {
		return mcp.RecheckToken(os.Getenv(mcp.TokenEnv), auth.LoadTokenConfig())
	}
	ro, err := mcp.OpenReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ro.Close() })
	server, cleanup, err := mcp.NewServer(id, ro, recheck, nil, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	session := connectInMemory(t, server, "2025-06-18")

	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tools) != 4 {
		t.Fatalf("tools = %+v", listed.Tools)
	}
	for _, name := range []string{"whoami", "list_my_tasks", "get_task"} {
		tool := toolByName(t, listed.Tools, name)
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Fatalf("%s annotations = %+v", name, tool.Annotations)
		}
	}
	create := toolByName(t, listed.Tools, "create_note")
	if create.Annotations == nil || create.Annotations.ReadOnlyHint {
		t.Fatalf("create_note annotations = %+v", create.Annotations)
	}

	before := noteCount(t, writer)
	res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "create_note",
		Arguments: map[string]any{"title": "Shift report", "body": "dock 4 is clear"},
	})
	if err != nil || res.IsError {
		t.Fatalf("create err=%v res=%+v", err, res)
	}
	if strings.Contains(toolText(t, res), tok) {
		t.Fatal("create leaked the token")
	}
	created := structuredMap(t, res.StructuredContent)
	idNum, _ := created["id"].(float64)
	if idNum <= 0 || !strings.Contains(fmt.Sprint(created["title"]), "Shift report") || !strings.Contains(fmt.Sprint(created["body"]), "dock 4 is clear") {
		t.Fatalf("create = %#v", created)
	}
	got, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "get_task",
		Arguments: map[string]any{"id": idNum},
	})
	if err != nil || got.IsError {
		t.Fatalf("get err=%v res=%+v", err, got)
	}
	gotMap := structuredMap(t, got.StructuredContent)
	if gotMap["title"] != created["title"] || gotMap["body"] != created["body"] {
		t.Fatalf("get = %#v", gotMap)
	}
	listedNotes, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "list_my_tasks",
		Arguments: map[string]any{"type": "note", "limit": 10},
	})
	if err != nil || listedNotes.IsError {
		t.Fatalf("list err=%v res=%+v", err, listedNotes)
	}
	if !strings.Contains(toolText(t, listedNotes), "Shift report") || strings.Contains(toolText(t, listedNotes), "Bea private") {
		t.Fatalf("list = %s", toolText(t, listedNotes))
	}

	empty, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "create_note", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if !empty.IsError {
		t.Fatal("empty create succeeded")
	}
	if noteCount(t, writer) != before+1 {
		t.Fatalf("empty create changed rows: %d", noteCount(t, writer))
	}

	t.Setenv(mcp.TokenEnv, "")
	missing, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "create_note",
		Arguments: map[string]any{"title": "Nope", "body": "secret"},
	})
	if err != nil {
		t.Fatalf("missing token protocol error: %v", err)
	}
	if !missing.IsError {
		t.Fatal("missing token create succeeded")
	}
	if strings.Contains(toolText(t, missing), "Nope") {
		t.Fatal("missing token error included the title")
	}
	bogus := "not-a-jwt"
	t.Setenv(mcp.TokenEnv, bogus)
	bad, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "create_note",
		Arguments: map[string]any{"title": "Nope", "body": "secret"},
	})
	if err != nil {
		t.Fatalf("invalid token protocol error: %v", err)
	}
	if !bad.IsError {
		t.Fatal("invalid token create succeeded")
	}
	badText := toolText(t, bad)
	if strings.Contains(badText, bogus) || strings.Contains(badText, "Nope") || strings.Contains(badText, "Bea private") {
		t.Fatalf("invalid token error = %q", badText)
	}
	if noteCount(t, writer) != before+1 {
		t.Fatalf("rejected create changed rows: %d", noteCount(t, writer))
	}

	beaTok := signToken(t, cfg, "bea-id", "bea@example.com", "bea", nil)
	t.Setenv(mcp.TokenEnv, beaTok)
	beaID, err := mcp.ResolveIdentity(beaTok, cfg)
	if err != nil {
		t.Fatal(err)
	}
	beaServer, beaCleanup, err := mcp.NewServer(beaID, ro, recheck, nil, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(beaCleanup)
	beaSession := connectInMemory(t, beaServer, "2025-06-18")
	foreign, err := beaSession.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "get_task",
		Arguments: map[string]any{"id": idNum},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !foreign.IsError {
		t.Fatalf("bea get succeeded: %+v", foreign)
	}
	foreignText := toolText(t, foreign)
	if strings.Contains(foreignText, "Shift report") || strings.Contains(foreignText, "dock 4") || strings.Contains(foreignText, "forbidden") {
		t.Fatalf("bea get = %q", foreignText)
	}
}

func TestNewServerRequiresUser(t *testing.T) {
	_, _, err := mcp.NewServer(mcp.Identity{}, nil, nil, nil, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProtocolBytesStayOffTheLog(t *testing.T) {
	cfg := testTokenConfig()
	tok := signToken(t, cfg, "user-7", "ada@example.com", "ada", []string{"Admin"})
	id, err := mcp.ResolveIdentity(tok, cfg)
	if err != nil {
		t.Fatal(err)
	}

	var protocol lockedBuf
	var logs lockedBuf
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	server, cleanup, err := mcp.NewServer(id, nil, nil, logger, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()
	t.Cleanup(func() {
		_ = clientReader.Close()
		_ = serverWriter.Close()
		_ = serverReader.Close()
		_ = clientWriter.Close()
	})

	ctx := context.Background()
	ss, err := server.Connect(ctx, &sdkmcp.IOTransport{
		Reader: serverReader,
		Writer: teeWriteCloser{buf: &protocol, next: serverWriter},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "morph-mcp-test", Version: "0"}, nil)
	session, err := client.Connect(ctx, &sdkmcp.IOTransport{
		Reader: clientReader,
		Writer: clientWriter,
	}, &sdkmcp.ClientSessionOptions{ProtocolVersion: "2025-06-18"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })

	if _, err := session.ListTools(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "whoami"}); err != nil {
		t.Fatal(err)
	}

	wire := protocol.String()
	if strings.TrimSpace(wire) == "" {
		t.Fatal("protocol writer captured nothing")
	}
	for _, line := range strings.Split(wire, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var msg map[string]any
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			t.Fatalf("protocol line is not JSON: %s", line)
		}
		if msg["jsonrpc"] != "2.0" {
			t.Fatalf("protocol line is not JSON-RPC: %s", line)
		}
	}
	if strings.Contains(wire, tok) {
		t.Fatal("protocol stream included the token")
	}
	if strings.Contains(wire, "server ready") {
		t.Fatal("protocol stream included a log line")
	}
	logText := logs.String()
	if !strings.Contains(logText, "morph-mcp server ready") {
		t.Fatalf("log = %q", logText)
	}
	if strings.Contains(logText, "user-7") || strings.Contains(logText, "user_id") {
		t.Fatalf("log included the user id: %q", logText)
	}
	if strings.Contains(logText, tok) {
		t.Fatal("log included the token")
	}
}

func connectInMemory(t *testing.T, server *sdkmcp.Server, protocol string) *sdkmcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "morph-mcp-test", Version: "0"}, nil)
	session, err := client.Connect(ctx, clientTransport, &sdkmcp.ClientSessionOptions{
		ProtocolVersion: protocol,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func structuredMap(t *testing.T, v any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("structured content %s: %v", raw, err)
	}
	return got
}

func toolText(t *testing.T, res *sdkmcp.CallToolResult) string {
	t.Helper()
	var b strings.Builder
	for _, c := range res.Content {
		tc, ok := c.(*sdkmcp.TextContent)
		if !ok {
			t.Fatalf("content type %T", c)
		}
		b.WriteString(tc.Text)
	}
	return b.String()
}

type lockedBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuf) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuf) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

type teeWriteCloser struct {
	buf  *lockedBuf
	next io.WriteCloser
}

func (t teeWriteCloser) Write(p []byte) (int, error) {
	if _, err := t.buf.Write(p); err != nil {
		return 0, err
	}
	return t.next.Write(p)
}

func (t teeWriteCloser) Close() error {
	return t.next.Close()
}
