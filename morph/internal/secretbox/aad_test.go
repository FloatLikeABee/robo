package secretbox_test

import (
	"bytes"
	"errors"
	"testing"

	"idongivaflyinfa/internal/secretbox"
)

func TestProviderKeyAADCanonicalAndRejects(t *testing.T) {
	got, err := secretbox.ProviderKeyAAD(secretbox.ScopeWorkspace, "owner-1", "openai")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "provider_key:workspace:owner-1:openai" {
		t.Fatal("workspace associated data mismatch")
	}
	got, err = secretbox.ProviderKeyAAD(secretbox.ScopeUser, "user-9", "anthropic")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "provider_key:user:user-9:anthropic" {
		t.Fatal("user associated data mismatch")
	}

	rejected := [][3]string{
		{"", "owner-1", "openai"},
		{"workspace", "", "openai"},
		{"workspace", "owner-1", ""},
		{"org", "owner-1", "openai"},
		{"workspace", "own:er", "openai"},
		{"workspace", "owner-1", "open:ai"},
		{"work:space", "owner-1", "openai"},
	}
	for _, tc := range rejected {
		if _, err := secretbox.ProviderKeyAAD(tc[0], tc[1], tc[2]); !errors.Is(err, secretbox.ErrInvalidAAD) {
			t.Fatal("expected invalid associated data")
		}
	}
}

func TestOpenRejectsMovedRow(t *testing.T) {
	box, err := secretbox.New(fakeKey(0x11))
	if err != nil {
		t.Fatal(err)
	}
	aad, err := secretbox.ProviderKeyAAD(secretbox.ScopeWorkspace, "owner-1", "openai")
	if err != nil {
		t.Fatal(err)
	}
	other, err := secretbox.ProviderKeyAAD(secretbox.ScopeWorkspace, "owner-2", "openai")
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("sk-test-row-binding")
	sealed, err := box.Seal(plain, aad)
	if err != nil {
		t.Fatal(err)
	}
	_, err = box.Open(sealed, other)
	if !errors.Is(err, secretbox.ErrDecrypt) {
		t.Fatal("expected decrypt failure for a moved row")
	}
	if _, err := box.Open(sealed, aad); err != nil || !bytes.Equal(mustOpen(t, box, sealed, aad), plain) {
		t.Fatal("original row should still open")
	}
}

func mustOpen(t *testing.T, box secretbox.Box, sealed string, aad []byte) []byte {
	t.Helper()
	got, err := box.Open(sealed, aad)
	if err != nil {
		t.Fatal(err)
	}
	return got
}
