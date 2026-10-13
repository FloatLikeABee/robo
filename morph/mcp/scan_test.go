package mcp_test

import (
	"testing"

	"idongivaflyinfa/mcp"
)

func TestScanShareTextBlocksSecretsAndEmail(t *testing.T) {
	if hits := mcp.ScanShareText("dock is clear"); len(hits) != 0 {
		t.Fatalf("clean = %#v", hits)
	}
	if hits := mcp.ScanShareText("key sk-abcdefghijklmnopqrstuv"); len(hits) == 0 {
		t.Fatal("missed api key")
	}
	if hits := mcp.ScanShareText("token eyJhbGciOiJIUzI1NiJ9.aaaa.bbbb"); len(hits) == 0 {
		t.Fatal("missed jwt")
	}
	if hits := mcp.ScanShareText("write ada@example.com later"); len(hits) == 0 {
		t.Fatal("missed email")
	}
	if hits := mcp.ScanShareText("API_KEY=supersecretvalue"); len(hits) == 0 {
		t.Fatal("missed env secret")
	}
}
