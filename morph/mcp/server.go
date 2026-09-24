package mcp

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// ServerName is the MCP serverInfo.name reported to clients.
	ServerName = "morph-mcp"
	// ServerVersion is the MCP serverInfo.version for this build.
	ServerVersion = "0.2.0"
)

const instructions = "Morph MCP stdio server. Read-only. whoami returns the Morph user from the verified session token. list_my_tasks and get_task return that user's own Notes and TODOs, not the shared MorphNotes Tasks board. HTTP JSON catalogs such as /ai/mcp-tools are not the Model Context Protocol."

// NewServer builds an MCP server that advertises whoami, list_my_tasks, and
// get_task. tasks is the read-only SQLite pool; nil makes the task tools
// return an error. recheck runs on every tool call; nil skips that check.
// logger receives server diagnostics. A nil logger discards them. The server
// does not bind a network port and does not open Badger.
func NewServer(id Identity, tasks *sql.DB, recheck func() error, logger *slog.Logger) (*sdkmcp.Server, error) {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	if strings.TrimSpace(id.UserID) == "" {
		return nil, errors.New("morph user id is required")
	}
	logger.Info("morph-mcp server ready")

	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    ServerName,
		Title:   "Morph",
		Version: ServerVersion,
	}, &sdkmcp.ServerOptions{
		Instructions: instructions,
		Logger:       logger,
		// A non-nil capabilities value suppresses the SDK's historical logging
		// capability. Tools are inferred when whoami is registered. Resources
		// are advertised with an empty list until a later story registers URIs.
		// ListChanged stays false: this process never emits list_changed.
		Capabilities: &sdkmcp.ServerCapabilities{
			Resources: &sdkmcp.ResourceCapabilities{ListChanged: false},
		},
	})
	closedWorld := false
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
	sdkmcp.AddTool(server, readOnly(
		"list_my_tasks",
		"List my tasks",
		"List the signed-in user's own Notes and TODOs (user_note_todo). Does not return the shared MorphNotes Tasks board. Optional type is all, note, or todo. Optional status is all, open, or done. Limit defaults to 50 and is capped at 100.",
	), listMyTasks(id, tasks, recheck))
	sdkmcp.AddTool(server, readOnly(
		"get_task",
		"Get task",
		"Read one of the signed-in user's Notes or TODOs by id. An id that is missing or belongs to someone else is not found.",
	), getTask(id, tasks, recheck))
	return server, nil
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
