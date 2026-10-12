package grill

import _ "embed"

// AppHTML is the MCP app page. The SPA modal loads the same bytes.
//
//go:embed app.html
var AppHTML string
