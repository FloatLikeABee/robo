package secretbox_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"idongivaflyinfa/internal/secretbox"
)

func fakeKey(b byte) []byte {
	return bytes.Repeat([]byte{b}, 32)
}

func TestSealOpenRoundTripAndUniqueNonce(t *testing.T) {
	box, err := secretbox.New(fakeKey(0x11))
	if err != nil {
		t.Fatal(err)
	}
	aad := []byte("row-1")
	plain := []byte("sk-test-not-a-real-key")
	sealed, err := box.Seal(plain, aad)
	if err != nil {
		t.Fatal(err)
	}
	got, err := box.Open(sealed, aad)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatal("opened plaintext did not match")
	}
	if strings.Contains(sealed, string(plain)) {
		t.Fatal("sealed token contained plaintext")
	}
	sum := sha256.Sum256(fakeKey(0x11))
	wantID := hex.EncodeToString(sum[:4])
	prefix := "v1:" + wantID + ":"
	if !strings.HasPrefix(sealed, prefix) {
		t.Fatal("token was not v1 with the sha256 key id")
	}
	again, err := box.Seal(plain, aad)
	if err != nil {
		t.Fatal(err)
	}
	if sealed == again {
		t.Fatal("two seals were identical")
	}
}

func TestSealOpenEmptyPlaintext(t *testing.T) {
	box, err := secretbox.New(fakeKey(0x11))
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := box.Seal(nil, []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := box.Open(sealed, []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatal("expected empty plaintext")
	}
}

func TestOpenWrongAADIsDecrypt(t *testing.T) {
	box, err := secretbox.New(fakeKey(0x11))
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("sk-test-PLAINTEXT-SENTINEL")
	sealed, err := box.Seal(plain, []byte("aad-a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = box.Open(sealed, []byte("aad-b"))
	if !errors.Is(err, secretbox.ErrDecrypt) {
		t.Fatal("expected decrypt failure")
	}
	if strings.Contains(err.Error(), string(plain)) || strings.Contains(err.Error(), "0x11") {
		t.Fatal("error text contained secret material")
	}
}

func TestOpenTamperAndTruncationIsDecrypt(t *testing.T) {
	box, err := secretbox.New(fakeKey(0x11))
	if err != nil {
		t.Fatal(err)
	}
	aad := []byte("aad")
	sealed, err := box.Seal([]byte("payload"), aad)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(sealed, ":")
	if len(parts) != 3 || parts[0] != "v1" {
		t.Fatal("token was not v1")
	}
	raw, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 0x01
	parts[2] = base64.StdEncoding.EncodeToString(raw)
	_, err = box.Open(strings.Join(parts, ":"), aad)
	if !errors.Is(err, secretbox.ErrDecrypt) {
		t.Fatal("expected decrypt failure on tamper")
	}
	_, err = box.Open(parts[0]+":"+parts[1]+":"+parts[2][:8], aad)
	if !errors.Is(err, secretbox.ErrDecrypt) {
		t.Fatal("expected decrypt failure on truncation")
	}
	_, err = box.Open("v1:zz", aad)
	if !errors.Is(err, secretbox.ErrDecrypt) {
		t.Fatal("expected decrypt failure on malformed v1 token")
	}
}

func TestOpenUnknownKeyIDIsNotDecrypt(t *testing.T) {
	a, err := secretbox.New(fakeKey(0x11))
	if err != nil {
		t.Fatal(err)
	}
	b, err := secretbox.New(fakeKey(0x22))
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("sk-test-PLAINTEXT-SENTINEL")
	sealed, err := a.Seal(plain, []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.Open(sealed, []byte("aad"))
	if !errors.Is(err, secretbox.ErrUnknownKeyID) {
		t.Fatal("expected unknown key id")
	}
	if errors.Is(err, secretbox.ErrDecrypt) {
		t.Fatal("unknown key id was reported as decrypt failure")
	}
	if strings.Contains(err.Error(), string(plain)) {
		t.Fatal("error text contained plaintext")
	}
}

func TestOpenUnsupportedVersion(t *testing.T) {
	box, err := secretbox.New(fakeKey(0x11))
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"", "v2:00112233:aaaa", "nope"} {
		_, err = box.Open(token, nil)
		if !errors.Is(err, secretbox.ErrUnsupportedVersion) {
			t.Fatalf("expected unsupported version")
		}
	}
}
