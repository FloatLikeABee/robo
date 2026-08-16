package docextract

import (
	"strings"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		file string
		mime string
		want Kind
	}{
		{"pdf by mime", "brief.pdf", "application/pdf", KindDocument},
		{"pdf by extension only", "brief.pdf", "", KindDocument},
		{"txt", "notes.txt", "text/plain", KindDocument},
		{"md", "readme.md", "text/markdown", KindDocument},
		{"markdown long ext", "readme.markdown", "", KindDocument},
		{"csv", "costs.csv", "text/csv", KindDocument},
		{"jpeg", "form.jpg", "image/jpeg", KindImage},
		{"jpeg alt ext", "form.jpeg", "image/jpeg", KindImage},
		{"png", "form.png", "image/png", KindImage},
		{"gif", "anim.gif", "image/gif", KindImage},
		{"webp", "shot.webp", "image/webp", KindImage},
		{"unsupported cad", "plan.dwg", "", KindUnsupported},
		{"unsupported no extension", "datafile", "", KindUnsupported},
		{"unsupported image subtype", "scan.tiff", "image/tiff", KindUnsupported},
		{"mime with charset", "notes.csv", "text/csv; charset=utf-8", KindDocument},
		{"unnamed text mime", "data.log", "text/x-log", KindDocument},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Classify(tc.file, tc.mime); got != tc.want {
				t.Fatalf("Classify(%q, %q) = %q, want %q", tc.file, tc.mime, got, tc.want)
			}
		})
	}
}

func TestClassifyGenericMimeFallsBackToExtension(t *testing.T) {
	for _, mime := range []string{"application/octet-stream", "binary/octet-stream", ""} {
		if got := Classify("brief.pdf", mime); got != KindDocument {
			t.Errorf("Classify with generic mime %q = %q, want document", mime, got)
		}
		if got := Classify("form.png", mime); got != KindImage {
			t.Errorf("Classify with generic mime %q = %q, want image", mime, got)
		}
		if got := Classify("plan.dwg", mime); got != KindUnsupported {
			t.Errorf("Classify with generic mime %q = %q, want unsupported", mime, got)
		}
	}
}

func TestIsPDF(t *testing.T) {
	if !IsPDF("a.pdf", "") {
		t.Error("expected .pdf extension to be detected")
	}
	if !IsPDF("upload", "application/pdf") {
		t.Error("expected application/pdf mime to be detected")
	}
	if IsPDF("a.txt", "text/plain") {
		t.Error("did not expect a text file to be detected as PDF")
	}
}

func TestAcceptedTypesMessage(t *testing.T) {
	docsOnly := AcceptedTypesMessage(KindDocument)
	for _, want := range []string{"PDF", "TXT", "CSV", "MD"} {
		if !strings.Contains(docsOnly, want) {
			t.Errorf("document message %q missing %q", docsOnly, want)
		}
	}
	if strings.Contains(docsOnly, "PNG") {
		t.Errorf("document-only message should not mention images: %q", docsOnly)
	}

	both := AcceptedTypesMessage(KindDocument, KindImage)
	if !strings.Contains(both, "PDF") || !strings.Contains(both, "PNG") {
		t.Errorf("combined message missing a type: %q", both)
	}
}

func TestExtractTextReplacesInvalidUTF8(t *testing.T) {
	raw := []byte{'h', 'i', 0xff, 0xfe, '!'}
	got := ExtractText(raw)
	if got == "" {
		t.Fatal("expected text despite invalid UTF-8")
	}
	if !strings.Contains(got, "hi") || !strings.Contains(got, "!") {
		t.Fatalf("valid characters lost: %q", got)
	}
	if strings.ContainsRune(got, 0xff) {
		t.Fatalf("invalid byte survived: %q", got)
	}
}

func TestExtractTextPreservesCSVRows(t *testing.T) {
	raw := []byte("date,amount,note\r\n2026-01-04,120.50,cement\r\n2026-01-05,80,fuel\r\n")
	got := ExtractText(raw)

	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 rows, got %d: %q", len(lines), got)
	}
	if lines[0] != "date,amount,note" {
		t.Errorf("header row altered: %q", lines[0])
	}
	if lines[2] != "2026-01-05,80,fuel" {
		t.Errorf("final data row altered: %q", lines[2])
	}
	if strings.Contains(got, "\r") {
		t.Error("carriage returns should be normalized away")
	}
}

func TestExtractTextCollapsesBlankLineRuns(t *testing.T) {
	got := ExtractText([]byte("a\n\n\n\nb"))
	if got != "a\n\nb" {
		t.Fatalf("got %q, want %q", got, "a\n\nb")
	}
}

func TestCollapseWhitespace(t *testing.T) {
	got := CollapseWhitespace("  spread   out \n\n across\tlines  ")
	if got != "spread out across lines" {
		t.Fatalf("got %q", got)
	}
}

func TestTruncateMarksCutContent(t *testing.T) {
	long := strings.Repeat("x", 100)

	got, cut := Truncate(long, 10)
	if !cut {
		t.Fatal("expected truncation to be reported")
	}
	if !strings.HasSuffix(got, TruncationMarker) {
		t.Fatalf("missing truncation marker: %q", got)
	}
	if !strings.HasPrefix(got, strings.Repeat("x", 10)) {
		t.Fatalf("kept content is wrong: %q", got)
	}
}

func TestTruncateLeavesShortContentAlone(t *testing.T) {
	got, cut := Truncate("short", 100)
	if cut {
		t.Error("did not expect truncation")
	}
	if got != "short" {
		t.Fatalf("content altered: %q", got)
	}
}

func TestTruncateCountsRunesNotBytes(t *testing.T) {
	// Five multi-byte runes must survive a five-rune cap untouched.
	s := "日本語です！"
	got, cut := Truncate(s, 6)
	if cut {
		t.Errorf("unexpected truncation of %q", got)
	}
	if got != s {
		t.Fatalf("got %q, want %q", got, s)
	}

	got, cut = Truncate(s, 3)
	if !cut {
		t.Fatal("expected truncation")
	}
	if !strings.HasPrefix(got, "日本語") {
		t.Fatalf("runes split incorrectly: %q", got)
	}
}

func TestExtractPDFBytesRejectsGarbage(t *testing.T) {
	// Must return an error rather than panicking out of the package.
	if _, err := ExtractPDFBytes([]byte("this is definitely not a pdf")); err == nil {
		t.Fatal("expected an error for non-PDF input")
	}
}
