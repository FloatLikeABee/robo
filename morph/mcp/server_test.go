package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
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

	server, err := mcp.NewServer(id, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
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
	if len(listed.Tools) != 3 || !toolNamed(listed.Tools, "whoami") {
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
	server, err := mcp.NewServer(id, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
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
	if len(listed.Tools) != 3 || !toolNamed(listed.Tools, "whoami") {
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
	server, err := mcp.NewServer(id, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	session := connectInMemory(t, server, "2025-06-18")
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, tool := range listed.Tools {
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Fatalf("%s annotations = %+v", tool.Name, tool.Annotations)
		}
		got[tool.Name] = true
	}
	for _, name := range []string{"whoami", "list_my_tasks", "get_task"} {
		if !got[name] {
			t.Fatalf("missing %s in %+v", name, listed.Tools)
		}
	}
	if len(listed.Tools) != 3 {
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
	server, err := mcp.NewServer(id, nil, recheck, nil)
	if err != nil {
		t.Fatal(err)
	}
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

func TestNewServerRequiresUser(t *testing.T) {
	_, err := mcp.NewServer(mcp.Identity{}, nil, nil, nil)
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
	server, err := mcp.NewServer(id, nil, nil, logger)
	if err != nil {
		t.Fatal(err)
	}

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
