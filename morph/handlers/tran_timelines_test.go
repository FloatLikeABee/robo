package handlers

import (
	"strings"
	"testing"
)

func TestTimelineFileExtOK(t *testing.T) {
	cases := map[string]bool{
		"a.txt":       true,
		"a.PDF":       true,
		"notes.md":    true,
		"x.markdown":  true,
		"photo.png":   false,
		"doc.docx":    false,
		"noext":       false,
		"archive.tar": false,
	}
	for name, want := range cases {
		if got := timelineFileExtOK(name); got != want {
			t.Errorf("timelineFileExtOK(%q)=%v want %v", name, got, want)
		}
	}
}

func TestIsBlockedTimelineHost(t *testing.T) {
	blocked := []string{"localhost", "127.0.0.1", "0.0.0.0", "10.0.0.1", "192.168.1.1", "169.254.169.254"}
	for _, h := range blocked {
		if !isBlockedTimelineHost(h) {
			t.Errorf("expected blocked host %q", h)
		}
	}
	if isBlockedTimelineHost("example.com") {
		t.Errorf("example.com should not be blocked by IP private rules alone")
	}
}

func TestExtractTimelineFileTextTxt(t *testing.T) {
	text, err := extractTimelineFileText("sample.txt", []byte("  Hello timeline  "))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Hello timeline") {
		t.Fatalf("unexpected text %q", text)
	}
	if _, err := extractTimelineFileText("x.png", []byte("nope")); err == nil {
		t.Fatal("expected unsupported type error")
	}
	if _, err := extractTimelineFileText("empty.txt", []byte("   ")); err == nil {
		t.Fatal("expected empty text error")
	}
}

func TestBuildTimelineHTML(t *testing.T) {
	html := buildTimelineHTML("Demo", "## 2020\n\n- Event one")
	if !strings.Contains(html, "Demo") || !strings.Contains(html, "Event one") {
		t.Fatalf("html missing content: %s", truncateRunes(html, 200))
	}
	if !strings.Contains(html, "<!doctype html>") {
		t.Fatal("expected full html document")
	}
}
