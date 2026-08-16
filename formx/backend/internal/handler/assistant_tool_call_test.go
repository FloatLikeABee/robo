package handler

import (
	"context"
	"strings"
	"testing"
)

func TestParseFormsXToolCallInitializesArgs(t *testing.T) {
	call, err := parseFormsXToolCall(`{"tool":"search_system_documents"}`)
	if err != nil {
		t.Fatalf("parseFormsXToolCall: %v", err)
	}
	if call.Tool != "search_system_documents" {
		t.Fatalf("unexpected tool %q", call.Tool)
	}
	if call.Args == nil {
		t.Fatal("expected args to be initialized")
	}
}

func TestExecFormsXSearchSystemDocumentsRequiresRepo(t *testing.T) {
	h := &Handler{}
	_, _, err := h.execFormsXTool(context.Background(), &formsXToolCall{
		Tool: "search_system_documents",
		Args: map[string]interface{}{"query": "registration"},
	})
	if err == nil || !strings.Contains(err.Error(), "ai document repo is not configured") {
		t.Fatalf("expected repo configuration error, got %v", err)
	}
}

func TestExecFormsXSearchSystemDocumentsRequiresQuery(t *testing.T) {
	h := &Handler{}
	_, err := h.execMongoMCPTool(context.Background(), "search_system_documents", map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "ai document repo is not configured") {
		t.Fatalf("expected repo configuration error before query validation, got %v", err)
	}
}
