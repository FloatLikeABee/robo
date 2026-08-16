package main

import (
	"context"
	"strings"
	"testing"
)

func TestParseTranMailToolCallInitializesArgs(t *testing.T) {
	call, err := parseTranMailToolCall(`{"tool":"get_template"}`)
	if err != nil {
		t.Fatalf("parseTranMailToolCall: %v", err)
	}
	if call.Tool != "get_template" {
		t.Fatalf("unexpected tool %q", call.Tool)
	}
	if call.Args == nil {
		t.Fatal("expected args to be initialized")
	}
}

func TestInt64ArgAcceptsStrings(t *testing.T) {
	got := int64Arg(map[string]interface{}{"id": "42"}, "id")
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestGetTemplateRequiresID(t *testing.T) {
	a := &App{}
	_, _, err := a.execTranMailTool(context.Background(), "", &tranMailToolCall{
		Tool: "get_template",
		Args: map[string]interface{}{},
	})
	if err == nil || !strings.Contains(err.Error(), "get_template requires id") {
		t.Fatalf("expected id error, got %v", err)
	}
}

func TestSearchReferenceDocsRequiresQuery(t *testing.T) {
	a := &App{}
	_, _, err := a.execTranMailTool(context.Background(), "", &tranMailToolCall{
		Tool: "search_reference_docs",
		Args: map[string]interface{}{},
	})
	if err == nil || !strings.Contains(err.Error(), "search_reference_docs requires query") {
		t.Fatalf("expected query error, got %v", err)
	}
}
