package auth

import (
	"strings"
	"testing"
)

func TestHashAndEqual(t *testing.T) {
	id, tok, err := NewPATSecret()
	if err != nil || id == "" || !IsPAT(tok) {
		t.Fatalf("new = %q %q %v", id, tok, err)
	}
	h := HashToken(tok)
	if h == tok || !EqualHash(h, HashToken(tok)) {
		t.Fatal("hash mismatch")
	}
	if EqualHash(h, HashToken(tok+"x")) {
		t.Fatal("different token compared equal")
	}
}

func TestNormalizeScopes(t *testing.T) {
	got, err := NormalizeScopes(nil)
	if err != nil || strings.Join(got, ",") != "notes:read,notes:write,pool:read" {
		t.Fatalf("default = %#v %v", got, err)
	}
	if _, err := NormalizeScopes([]string{"notes:read", "nope"}); err == nil {
		t.Fatal("accepted invalid scope")
	}
	if HasScope(nil, ScopeNotesRead) && HasScope(nil, ScopePoolPublish) {
		t.Fatal("empty scopes must not grant publish")
	}
	if !HasScope(nil, ScopeNotesWrite) || HasScope([]string{ScopeNotesRead}, ScopeNotesWrite) {
		t.Fatal("scope check")
	}
}
