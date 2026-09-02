package repoenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromUsesRepoRootNotNested(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "start-all.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("REPOENV_TEST_KEY=from-root\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "formx", "backend")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, ".env"), []byte("REPOENV_TEST_KEY=\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("REPOENV_TEST_KEY") })
	_ = os.Unsetenv("REPOENV_TEST_KEY")

	if err := LoadFrom(nested); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("REPOENV_TEST_KEY"); got != "from-root" {
		t.Fatalf("got %q, want from-root (nested empty .env must be ignored)", got)
	}
}

func TestLoadFromIgnoresEmptyNestedMorphAIKey(t *testing.T) {
	orig, had := os.LookupEnv("MORPH_AI_API_KEY")
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("MORPH_AI_API_KEY", orig)
		} else {
			_ = os.Unsetenv("MORPH_AI_API_KEY")
		}
	})
	_ = os.Unsetenv("MORPH_AI_API_KEY")

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "start-all.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("MORPH_AI_API_KEY=root-formsx-key\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "formx", "backend")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, ".env"), []byte("MORPH_AI_API_KEY=\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := LoadFrom(nested); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("MORPH_AI_API_KEY"); got != "root-formsx-key" {
		t.Fatalf("got %q, want root-formsx-key (leftover empty formx/backend/.env must be ignored)", got)
	}
}

func TestFindRepoRoot(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, ok := FindRepoRoot(nested); ok {
		t.Fatal("expected no root without start-all.sh")
	}
	if err := os.WriteFile(filepath.Join(root, "start-all.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, ok := FindRepoRoot(nested)
	if !ok {
		t.Fatal("expected to find repo root")
	}
	want, _ := filepath.Abs(root)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
