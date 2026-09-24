package mcp

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// ServerName is the MCP serverInfo.name reported to clients.
	ServerName = "morph-mcp"
	// ServerVersion is the MCP serverInfo.version for this skeleton.
	ServerVersion = "0.1.0"
)

const instructions = "Morph MCP stdio server. Read-only. The whoami tool returns the Morph user from the verified session token. MorphNotes data tools are not registered. HTTP JSON catalogs such as /ai/mcp-tools are not the Model Context Protocol."

// NewServer builds an MCP server that advertises tools and the whoami tool.
// logger receives server diagnostics. A nil logger discards them. The server
// does not bind a network port and does not open Morph data stores.
func NewServer(id Identity, logger *slog.Logger) (*sdkmcp.Server, error) {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	if strings.TrimSpace(id.UserID) == "" {
		return nil, errors.New("morph user id is required")
	}
	logger.Info("morph-mcp server ready", "user_id", id.UserID)

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
		Capabilities: &sdkmcp.ServerCapabilities{
			Resources: &sdkmcp.ResourceCapabilities{ListChanged: true},
		},
	})
	closedWorld := false
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "whoami",
		Description: "Return the Morph user this read-only MCP server is acting as.",
		Annotations: &sdkmcp.ToolAnnotations{
			Title:         "Who am I",
			ReadOnlyHint:  true,
			OpenWorldHint: &closedWorld,
		},
	}, whoami(id))
	return server, nil
}

type whoamiInput struct{}

type whoamiOutput struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

func whoami(id Identity) func(context.Context, *sdkmcp.CallToolRequest, whoamiInput) (*sdkmcp.CallToolResult, whoamiOutput, error) {
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
		return nil, out, nil
	}
}
