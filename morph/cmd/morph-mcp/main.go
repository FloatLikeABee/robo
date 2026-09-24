// Command morph-mcp is a local Model Context Protocol server over stdio.
//
// Stdout is reserved for MCP JSON-RPC messages. Logs go to stderr.
// The process does not listen on a port and does not open Badger or SQLite.
package main

import (
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"net"
	"os"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/robo/repoenv"

	"idongivaflyinfa/mcp"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	if err := run(); err != nil {
		log.Fatalf("morph-mcp: %v", err)
	}
}

func run() error {
	if err := repoenv.Load(); err != nil {
		return err
	}
	id, err := mcp.IdentityFromEnv()
	if err != nil {
		return err
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	server, err := mcp.NewServer(id, logger)
	if err != nil {
		return err
	}
	err = server.Run(context.Background(), &sdkmcp.StdioTransport{})
	if err == nil || clientDisconnected(err) {
		return nil
	}
	return err
}

func clientDisconnected(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrClosedPipe) ||
		errors.Is(err, net.ErrClosed) ||
		errors.Is(err, context.Canceled)
}
