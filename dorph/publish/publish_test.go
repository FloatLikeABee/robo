package publish

import "testing"

func TestVisible(t *testing.T) {
	if Visible("") {
		t.Fatal("empty slug is private")
	}
	if Visible("   ") {
		t.Fatal("whitespace slug is private")
	}
	if !Visible("public-topic") {
		t.Fatal("non-empty slug is published")
	}
}

func TestPageRouteMatchesPublishedPagesOnly(t *testing.T) {
	if !PageRoute("GET", "/api/tran/public/research/public-topic") {
		t.Fatal("published research GET is a public page")
	}
	if !PageRoute("HEAD", "/api/tran/public/big-notes/a-note") {
		t.Fatal("published big-note HEAD is a public page")
	}
	if PageRoute("GET", "/api/tran/research") {
		t.Fatal("private list is not a public page")
	}
	if PageRoute("POST", "/api/tran/public/research/public-topic") {
		t.Fatal("POST is not a public page")
	}
}
