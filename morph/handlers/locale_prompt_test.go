package handlers

import "strings"
import "testing"

func TestReplyLanguageLine(t *testing.T) {
	zh := replyLanguageLine("zh")
	if !strings.Contains(zh, "Simplified Chinese") || !strings.Contains(zh, "another language") {
		t.Fatalf("zh line: %q", zh)
	}
	en := replyLanguageLine(" EN ")
	if !strings.Contains(en, "English") || !strings.Contains(en, "another language") {
		t.Fatalf("en line: %q", en)
	}
	if replyLanguageLine("zh-TW") != "" || replyLanguageLine("fr") != "" || replyLanguageLine("") != "" {
		t.Fatal("only en and zh add a reply-language line")
	}
}
