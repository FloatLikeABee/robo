package secretbox_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"idongivaflyinfa/internal/secretbox"
)

func TestDevKeyFileCreateReuseAndMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "morph-secrets.key")
	key, created, err := secretbox.LoadOrCreateDevKeyFile(path)
	if err != nil || !created || len(key) != 32 {
		t.Fatal("expected a new 32-byte key")
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		t.Fatal("key file should be a regular file")
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatal("key file mode should be 0600")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(body), "\n") {
		t.Fatal("key file should end with a newline")
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(body)))
	if err != nil || !bytes.Equal(decoded, key) {
		t.Fatal("key file did not contain the returned key")
	}

	again, created, err := secretbox.LoadOrCreateDevKeyFile(path)
	if err != nil || created || !bytes.Equal(again, key) {
		t.Fatal("existing key file was not reused")
	}
	body2, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(body, body2) {
		t.Fatal("existing key file was rewritten")
	}
}

func TestResolveDevCreatesThenReuses(t *testing.T) {
	t.Setenv("MORPH_SECRETS_KEY", "")
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", "")
	path := filepath.Join(t.TempDir(), "morph-secrets.key")
	box, created, err := secretbox.Resolve(false, path)
	if err != nil || !created {
		t.Fatal("expected a created dev key")
	}
	sealed, err := box.Seal([]byte("local"), []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	again, created, err := secretbox.Resolve(false, path)
	if err != nil || created {
		t.Fatal("second resolve should reuse the file")
	}
	if _, err := again.Open(sealed, []byte("aad")); err != nil {
		t.Fatal("reused dev key did not open")
	}
}

func TestResolveProductionAndInvalidEnvDoNotCreate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "morph-secrets.key")
	t.Setenv("MORPH_SECRETS_KEY", "")
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", "")
	_, _, err := secretbox.Resolve(true, path)
	if !errors.Is(err, secretbox.ErrKeyMissing) {
		t.Fatal("expected missing key in production")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("production created a key file")
	}

	t.Setenv("MORPH_SECRETS_KEY", "NOT-A-REAL-KEY")
	_, _, err = secretbox.Resolve(false, path)
	if !errors.Is(err, secretbox.ErrInvalidKey) {
		t.Fatal("expected invalid key")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("invalid env fell back to a key file")
	}
}

func TestResolvePreviousWithoutCurrentDoesNotGenerate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "morph-secrets.key")
	t.Setenv("MORPH_SECRETS_KEY", "")
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", b64(fakeKey(0x11)))
	_, _, err := secretbox.Resolve(false, path)
	if !errors.Is(err, secretbox.ErrInvalidKey) {
		t.Fatal("expected invalid key")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("generated a key while previous keys were set")
	}
}

func TestResolveUsesEnvInsteadOfFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "morph-secrets.key")
	envKey := fakeKey(0x61)
	t.Setenv("MORPH_SECRETS_KEY", b64(envKey))
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", "")
	box, created, err := secretbox.Resolve(false, path)
	if err != nil || created {
		t.Fatal("env key should be used without creating a file")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("env key still created a file")
	}
	sealed, err := box.Seal([]byte("x"), []byte("a"))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(envKey)
	if !strings.Contains(sealed, hex.EncodeToString(sum[:4])) {
		t.Fatal("seal did not use the env key")
	}
}

func TestDevKeyFileLooseModeAndSymlinkAndMissingDir(t *testing.T) {
	dir := t.TempDir()
	loose := filepath.Join(dir, "loose.key")
	encoded := b64(fakeKey(0x11)) + "\n"
	if err := os.WriteFile(loose, []byte(encoded), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := secretbox.LoadOrCreateDevKeyFile(loose); !errors.Is(err, secretbox.ErrKeyFile) {
		t.Fatal("expected key file error for loose permissions")
	}
	info, err := os.Stat(loose)
	if err != nil || info.Mode().Perm()&0o077 == 0 {
		t.Fatal("loose key file should not be chmodded into use")
	}
	t.Setenv("MORPH_SECRETS_KEY", "")
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", "")
	if _, _, err := secretbox.Resolve(false, loose); !errors.Is(err, secretbox.ErrKeyFile) {
		t.Fatal("resolve should refuse a loose key file")
	}
	_, _, looseErr := secretbox.LoadOrCreateDevKeyFile(loose)
	if looseErr == nil || strings.Contains(looseErr.Error(), strings.TrimSpace(encoded)) {
		t.Fatal("error contained key material")
	}

	target := filepath.Join(dir, "target.key")
	if err := os.WriteFile(target, []byte(encoded), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.key")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, _, err := secretbox.LoadOrCreateDevKeyFile(link); !errors.Is(err, secretbox.ErrKeyFile) {
		t.Fatal("expected key file error for a symlink")
	}

	missing := filepath.Join(dir, "no-such-dir", "morph-secrets.key")
	if _, _, err := secretbox.LoadOrCreateDevKeyFile(missing); !errors.Is(err, secretbox.ErrKeyFile) {
		t.Fatal("expected key file error for a missing directory")
	}
}

func TestResolveIgnoresExistingFileWhenEnvOrProduction(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "morph-secrets.key")
	fileKey := fakeKey(0x99)
	body := []byte(b64(fileKey) + "\n")
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MORPH_SECRETS_KEY", "")
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", "")
	if _, _, err := secretbox.Resolve(true, path); !errors.Is(err, secretbox.ErrKeyMissing) {
		t.Fatal("production should ignore an existing dev key file")
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, body) {
		t.Fatal("production changed the dev key file")
	}

	envKey := fakeKey(0x61)
	t.Setenv("MORPH_SECRETS_KEY", b64(envKey))
	box, created, err := secretbox.Resolve(false, path)
	if err != nil || created {
		t.Fatal("env key should not create or replace the file")
	}
	sealed, err := box.Seal([]byte("x"), []byte("a"))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(envKey)
	fileSum := sha256.Sum256(fileKey)
	if !strings.Contains(sealed, hex.EncodeToString(sum[:4])) || strings.Contains(sealed, hex.EncodeToString(fileSum[:4])) {
		t.Fatal("seal used the file key instead of the env key")
	}
	got, err = os.ReadFile(path)
	if err != nil || !bytes.Equal(got, body) {
		t.Fatal("env resolve rewrote the dev key file")
	}
}

func TestDevKeyFileCorruptAndNonRegularAreRefused(t *testing.T) {
	dir := t.TempDir()
	corrupt := filepath.Join(dir, "corrupt.key")
	original := []byte("not-a-key\n")
	if err := os.WriteFile(corrupt, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := secretbox.LoadOrCreateDevKeyFile(corrupt); !errors.Is(err, secretbox.ErrKeyFile) {
		t.Fatal("expected key file error for corrupt contents")
	}
	got, err := os.ReadFile(corrupt)
	if err != nil || !bytes.Equal(got, original) {
		t.Fatal("corrupt key file was overwritten")
	}
	if _, _, err := secretbox.LoadOrCreateDevKeyFile(dir); !errors.Is(err, secretbox.ErrKeyFile) {
		t.Fatal("expected key file error for a directory")
	}
}
