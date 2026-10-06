package mcp

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"log"
	"log/slog"
	"strings"
	"sync"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed notes_app.html
var notesAppHTML string

const (
	notesAppURI  = "ui://morph/notes"
	notesAppMIME = "text/html;profile=mcp-app"
	uiExtension  = "io.modelcontextprotocol/ui"
)

const (
	// ServerName is the MCP serverInfo.name reported to clients.
	ServerName = "morph-mcp"
	// ServerVersion is the MCP serverInfo.version for this build.
	ServerVersion = "0.3.0"
)

const instructions = "Morph MCP stdio server. whoami, list_my_tasks, and get_task are read-only Notes and TODOs tools. create_note stores a note and marks it [morph-mcp]. create, list, and get tools also cover stick notes, timelines, stories, research, and event logs for the same session. Event log tools call the Morph events API and do not open Badger. An MCP Apps host renders the notes panel from ui://morph/notes. HTTP JSON catalogs such as /ai/mcp-tools are not the Model Context Protocol."

// NewServer builds an MCP server that advertises whoami, the Notes and TODOs
// tools, and create, list, and get for stick notes, timelines, stories,
// research, and event logs, plus the ui://morph/notes app resource.
// tasks is the read-only SQLite pool; nil makes the read tools return an
// error. sqlitePath is opened mode=rw on the first create_note. recheck runs
// on every tool call; nil skips that check. logger receives server diagnostics.
// A nil logger discards them. The closer releases the write connection. The
// server does not bind a network port and does not open Badger.
func NewServer(id Identity, tasks *sql.DB, recheck func() error, logger *slog.Logger, sqlitePath string) (*sdkmcp.Server, func(), error) {
	noop := func() {}
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	if strings.TrimSpace(id.UserID) == "" {
		return nil, noop, errors.New("morph user id is required")
	}
	notes := &noteStore{path: sqlitePath}
	logger.Info("morph-mcp server ready")

	caps := &sdkmcp.ServerCapabilities{
		// ListChanged stays false: this process never emits list_changed.
		Resources: &sdkmcp.ResourceCapabilities{ListChanged: false},
	}
	caps.AddExtension(uiExtension, map[string]any{
		"mimeTypes": []string{notesAppMIME},
	})
	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    ServerName,
		Title:   "Morph",
		Version: ServerVersion,
	}, &sdkmcp.ServerOptions{
		Instructions: instructions,
		Logger:       logger,
		// A non-nil capabilities value suppresses the SDK's historical logging
		// capability. Tools are inferred when whoami is registered.
		Capabilities: caps,
	})
	closedWorld := false
	additive := false
	readOnly := func(name, title, description string) *sdkmcp.Tool {
		return &sdkmcp.Tool{
			Name:        name,
			Description: description,
			Annotations: &sdkmcp.ToolAnnotations{
				Title:         title,
				ReadOnlyHint:  true,
				OpenWorldHint: &closedWorld,
			},
		}
	}
	sdkmcp.AddTool(server, readOnly(
		"whoami",
		"Who am I",
		"Return the Morph user this read-only MCP server is acting as.",
	), whoami(id, recheck))
	listTool := readOnly(
		"list_my_tasks",
		"List my tasks",
		"List the signed-in user's own Notes and TODOs (user_note_todo). Does not return the shared MorphNotes Tasks board. Optional type is all, note, or todo. Optional status is all, open, or done. Limit defaults to 50 and is capped at 100.",
	)
	listTool.Meta = notesAppToolMeta()
	sdkmcp.AddTool(server, listTool, listMyTasks(id, tasks, recheck))
	getTool := readOnly(
		"get_task",
		"Get task",
		"Read one of the signed-in user's Notes or TODOs by id. An id that is missing or belongs to someone else is not found.",
	)
	getTool.Meta = notesAppToolMeta()
	sdkmcp.AddTool(server, getTool, getTask(id, tasks, recheck))
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "create_note",
		Description: "Create a note in the signed-in user's Notes and TODOs (user_note_todo). The stored title starts with [morph-mcp] and the body starts with source: morph-mcp. Title or body is required. Title max 200 characters. Body max 32000 characters. Does not create a TODO or a shared MorphNotes Tasks board row.",
		Meta:        notesAppToolMeta(),
		Annotations: &sdkmcp.ToolAnnotations{
			Title:           "Create note",
			ReadOnlyHint:    false,
			DestructiveHint: &additive,
			OpenWorldHint:   &closedWorld,
		},
	}, createNote(id, notes, recheck))
	registerModuleTools(server, id, tasks, notes, recheck, readOnly)
	server.AddResource(notesAppResource(), readNotesApp)
	return server, notes.close, nil
}

func notesAppToolMeta() sdkmcp.Meta {
	return sdkmcp.Meta{
		"ui": map[string]any{
			"resourceUri": notesAppURI,
			"visibility":  []string{"model", "app"},
		},
		"ui/resourceUri": notesAppURI,
	}
}

func notesAppUIMeta() map[string]any {
	return map[string]any{
		"csp": map[string]any{
			"connectDomains":  []string{},
			"resourceDomains": []string{},
			"frameDomains":    []string{},
		},
		"prefersBorder": true,
	}
}

func notesAppResource() *sdkmcp.Resource {
	return &sdkmcp.Resource{
		URI:         notesAppURI,
		Name:        "morph-notes",
		Title:       "Morph notes",
		Description: "Create, list, and open the signed-in user's Morph notes.",
		MIMEType:    notesAppMIME,
		Meta:        sdkmcp.Meta{"ui": notesAppUIMeta()},
	}
}

func readNotesApp(context.Context, *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
	return &sdkmcp.ReadResourceResult{
		Contents: []*sdkmcp.ResourceContents{{
			URI:      notesAppURI,
			MIMEType: notesAppMIME,
			Text:     notesAppHTML,
			Meta:     sdkmcp.Meta{"ui": notesAppUIMeta()},
		}},
	}, nil
}

type noteStore struct {
	path string
	mu   sync.Mutex
	db   *sql.DB
}

func (s *noteStore) close() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		_ = s.db.Close()
		s.db = nil
	}
}

func (s *noteStore) writer() (*sql.DB, error) {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return nil, errors.New("note store is not open")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		return s.db, nil
	}
	db, err := OpenReadWrite(s.path)
	if err != nil {
		log.Printf("create note: %v", err)
		return nil, errors.New("note store is not open")
	}
	s.db = db
	return db, nil
}

func callRecheck(recheck func() error) error {
	if recheck == nil {
		return nil
	}
	return recheck()
}

type whoamiInput struct{}

type whoamiOutput struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

func whoami(id Identity, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, whoamiInput) (*sdkmcp.CallToolResult, whoamiOutput, error) {
	roles := id.Roles
	if roles == nil {
		roles = []string{}
	} else {
		roles = append([]string(nil), roles...)
	}
	out := whoamiOutput{
		ID:       id.UserID,
		Email:    id.Email,
		Username: id.Username,
		Roles:    roles,
	}
	return func(context.Context, *sdkmcp.CallToolRequest, whoamiInput) (*sdkmcp.CallToolResult, whoamiOutput, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, whoamiOutput{}, err
		}
		return nil, out, nil
	}
}

type listInput struct {
	Type   string `json:"type,omitempty" jsonschema:"all, note, or todo. Default all."`
	Status string `json:"status,omitempty" jsonschema:"all, open, or done. Default all."`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum rows. Default 50, capped at 100."`
}

type getInput struct {
	ID int `json:"id" jsonschema:"user_note_todo id"`
}

func listMyTasks(id Identity, tasks *sql.DB, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, listInput) (*sdkmcp.CallToolResult, ListResult, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in listInput) (*sdkmcp.CallToolResult, ListResult, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, ListResult{}, err
		}
		out, err := ListMyTasks(ctx, tasks, id.UserID, TaskFilter{Type: in.Type, Status: in.Status, Limit: in.Limit})
		if err != nil {
			return nil, ListResult{}, err
		}
		return nil, out, nil
	}
}

type createNoteInput struct {
	Title string `json:"title,omitempty" jsonschema:"Note title. Optional. At most 200 characters."`
	Body  string `json:"body,omitempty" jsonschema:"Note text. Optional. At most 32000 characters. Title or body is required."`
}

func createNote(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, createNoteInput) (*sdkmcp.CallToolResult, Task, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in createNoteInput) (*sdkmcp.CallToolResult, Task, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, Task{}, err
		}
		db, err := notes.writer()
		if err != nil {
			return nil, Task{}, err
		}
		task, err := CreateMyNote(ctx, db, id.UserID, in.Title, in.Body)
		if err != nil {
			return nil, Task{}, err
		}
		return nil, task, nil
	}
}

func getTask(id Identity, tasks *sql.DB, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, getInput) (*sdkmcp.CallToolResult, Task, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in getInput) (*sdkmcp.CallToolResult, Task, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, Task{}, err
		}
		task, err := GetMyTask(ctx, tasks, id.UserID, in.ID)
		if err != nil {
			return nil, Task{}, err
		}
		return nil, task, nil
	}
}
