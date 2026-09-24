package main

import (
	"bytes"
	"encoding/base64"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"idongivaflyinfa/internal/secretbox"
)

func TestOpenSecretsBoxProductionMissingOrInvalidRefuses(t *testing.T) {
	// Break this catches: production start with a missing or invalid MORPH_SECRETS_KEY, or an error that prints the value.
	dir := t.TempDir()
	sqlitePath := filepath.Join(dir, "tran.sqlite")
	keyPath := filepath.Join(dir, "morph-secrets.key")
	if err := os.WriteFile(keyPath, []byte("CANARY-FILE-KEY\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MORPH_SECRETS_KEY", "")
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", "")

	_, err := openSecretsBox(true, sqlitePath)
	if err == nil {
		t.Fatal("expected production to refuse a missing key")
	}
	msg := err.Error()
	for _, want := range []string{"refusing to start", "MORPH_SECRETS_KEY"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %s", msg, want)
		}
	}
	if strings.Contains(msg, "CANARY-FILE-KEY") {
		t.Fatalf("error leaked key file contents: %s", msg)
	}
	body, err := os.ReadFile(keyPath)
	if err != nil || string(body) != "CANARY-FILE-KEY\n" {
		t.Fatal("production rewrote or removed the key file")
	}

	const canary = "CANARY-INVALID-SECRETS-KEY-VALUE"
	t.Setenv("MORPH_SECRETS_KEY", canary)
	_, err = openSecretsBox(true, sqlitePath)
	if err == nil {
		t.Fatal("expected production to refuse an invalid key")
	}
	msg = err.Error()
	if !strings.Contains(msg, "MORPH_SECRETS_KEY") || !strings.Contains(msg, "refusing to start") {
		t.Fatalf("error %q", msg)
	}
	if strings.Contains(msg, canary) {
		t.Fatalf("error leaked MORPH_SECRETS_KEY: %s", msg)
	}
}

func TestOpenSecretsBoxProductionLoadsPreviousKey(t *testing.T) {
	// Break this catches: MORPH_SECRETS_KEY_PREVIOUS ignored, so rotation tokens cannot be opened at startup.
	dir := t.TempDir()
	sqlitePath := filepath.Join(dir, "tran.sqlite")
	current := bytes.Repeat([]byte{0x21}, 32)
	previous := bytes.Repeat([]byte{0x43}, 32)
	old, err := secretbox.New(previous)
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := old.Seal([]byte("provider-key"), []byte("row"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MORPH_SECRETS_KEY", base64.StdEncoding.EncodeToString(current))
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", base64.StdEncoding.EncodeToString(previous))

	box, err := openSecretsBox(true, sqlitePath)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := box.Open(sealed, []byte("row"))
	if err != nil || string(plain) != "provider-key" {
		t.Fatalf("previous key was not loaded: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "morph-secrets.key")); !os.IsNotExist(err) {
		t.Fatal("production with an env key created a key file")
	}

	const badPrevious = "CANARY-BAD-PREVIOUS-KEY"
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", badPrevious)
	_, err = openSecretsBox(true, sqlitePath)
	if err == nil {
		t.Fatal("expected invalid previous key to refuse startup")
	}
	msg := err.Error()
	if !strings.Contains(msg, "MORPH_SECRETS_KEY_PREVIOUS") || strings.Contains(msg, badPrevious) {
		t.Fatalf("previous-key error %q", msg)
	}
}

func TestOpenSecretsBoxLocalFileAndWarning(t *testing.T) {
	// Break this catches: local startup failing closed when the secrets env is unset, or logging the key itself.
	dir := t.TempDir()
	sqlitePath := filepath.Join(dir, "tran.sqlite")
	keyPath := filepath.Join(dir, "morph-secrets.key")
	t.Setenv("MORPH_SECRETS_KEY", "")
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", "")

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	box, err := openSecretsBox(false, sqlitePath)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 || !info.Mode().IsRegular() {
		t.Fatalf("key file mode %o", info.Mode().Perm())
	}
	sealed, err := box.Seal([]byte("local"), []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	wantWarn := "warning: generated dev MORPH_SECRETS_KEY file at " + keyPath + " (mode 0600); set MORPH_SECRETS_KEY before production"
	logged := buf.String()
	if !strings.Contains(logged, wantWarn) {
		t.Fatalf("log %q missing warning", logged)
	}
	if strings.Contains(logged, strings.TrimSpace(string(body))) {
		t.Fatal("warning included the key")
	}

	buf.Reset()
	again, err := openSecretsBox(false, sqlitePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "generated dev MORPH_SECRETS_KEY") {
		t.Fatal("reused key file logged the create warning")
	}
	if _, err := again.Open(sealed, []byte("aad")); err != nil {
		t.Fatal(err)
	}
	body2, err := os.ReadFile(keyPath)
	if err != nil || !bytes.Equal(body, body2) {
		t.Fatal("existing key file was rewritten")
	}
}

func TestOpenSecretsBoxLocalPreviousWithoutCurrentRefuses(t *testing.T) {
	// Break this catches: local startup minting a new current key while MORPH_SECRETS_KEY_PREVIOUS is set.
	dir := t.TempDir()
	sqlitePath := filepath.Join(dir, "tran.sqlite")
	t.Setenv("MORPH_SECRETS_KEY", "")
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x11}, 32)))

	_, err := openSecretsBox(false, sqlitePath)
	if err == nil {
		t.Fatal("expected refusal")
	}
	msg := err.Error()
	if !strings.Contains(msg, "MORPH_SECRETS_KEY") || !strings.Contains(msg, "MORPH_SECRETS_KEY_PREVIOUS") {
		t.Fatalf("error %q", msg)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "morph-secrets.key")); !os.IsNotExist(statErr) {
		t.Fatal("generated a key file")
	}
}
