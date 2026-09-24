package secretbox_test

import (
	"bytes"
	"crypto/rand"
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

func b64(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

func TestFromEnvMissingAndInvalid(t *testing.T) {
	t.Setenv("MORPH_SECRETS_KEY", "")
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", "")
	_, err := secretbox.FromEnv()
	if !errors.Is(err, secretbox.ErrKeyMissing) {
		t.Fatal("expected missing key")
	}

	sentinel := "NOT-A-KEY-SENTINEL-XYZ"
	t.Setenv("MORPH_SECRETS_KEY", sentinel)
	_, err = secretbox.FromEnv()
	if !errors.Is(err, secretbox.ErrInvalidKey) {
		t.Fatal("expected invalid key")
	}
	if strings.Contains(err.Error(), sentinel) {
		t.Fatal("error echoed the env value")
	}

	short := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x11}, 16))
	t.Setenv("MORPH_SECRETS_KEY", short)
	_, err = secretbox.FromEnv()
	if !errors.Is(err, secretbox.ErrInvalidKey) || strings.Contains(err.Error(), short) {
		t.Fatal("expected invalid key without echoing the value")
	}

	url := base64.URLEncoding.EncodeToString(bytes.Repeat([]byte{0xff}, 32))
	if !strings.ContainsAny(url, "-_") {
		t.Fatal("fixture was not url-safe base64")
	}
	t.Setenv("MORPH_SECRETS_KEY", url)
	_, err = secretbox.FromEnv()
	if !errors.Is(err, secretbox.ErrInvalidKey) || strings.Contains(err.Error(), url) {
		t.Fatal("expected url-safe encoding to be rejected")
	}
}

func TestFromEnvAcceptsWhitespaceAndUnpadded(t *testing.T) {
	key := fakeKey(0x41)
	raw := b64(key)
	broken := raw[:20] + "\n" + raw[20:]
	t.Setenv("MORPH_SECRETS_KEY", broken)
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", "")
	box, err := secretbox.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := box.Seal([]byte("x"), []byte("a"))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(key)
	if !strings.Contains(sealed, hex.EncodeToString(sum[:4])) {
		t.Fatal("sealed token did not use the env key")
	}

	unpadded := strings.TrimRight(raw, "=")
	t.Setenv("MORPH_SECRETS_KEY", unpadded)
	if _, err := secretbox.FromEnv(); err != nil {
		t.Fatal(err)
	}
}

func TestFromEnvEmptyPreviousSlot(t *testing.T) {
	t.Setenv("MORPH_SECRETS_KEY", b64(fakeKey(0x11)))
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", b64(fakeKey(0x22))+",")
	_, err := secretbox.FromEnv()
	if !errors.Is(err, secretbox.ErrInvalidKey) {
		t.Fatal("expected invalid key for empty previous slot")
	}
}

func TestFromEnvDoesNotReadDotEnv(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("MORPH_SECRETS_KEY="+b64(fakeKey(0x11))+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("MORPH_SECRETS_KEY", "")
	t.Setenv("MORPH_SECRETS_KEY_PREVIOUS", "")
	_, err := secretbox.FromEnv()
	if !errors.Is(err, secretbox.ErrKeyMissing) {
		t.Fatal("FromEnv loaded a dotenv file")
	}
}

func TestPreviousKeyOpensAndReseals(t *testing.T) {
	prev := fakeKey(0x21)
	cur := fakeKey(0x22)
	older := fakeKey(0x23)
	plain := []byte("sk-test-rotate-me")
	aad := []byte("provider_key:workspace:owner-1:openai")

	oldBox, err := secretbox.New(prev)
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := oldBox.Seal(plain, aad)
	if err != nil {
		t.Fatal(err)
	}
	box, err := secretbox.New(cur, prev, older)
	if err != nil {
		t.Fatal(err)
	}
	got, err := box.Open(sealed, aad)
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatal("previous key did not open")
	}
	if !box.NeedsRotation(sealed) {
		t.Fatal("expected rotation")
	}
	out, rotated, err := secretbox.Reseal(box, sealed, aad)
	if err != nil || !rotated {
		t.Fatal("expected reseal")
	}
	if box.NeedsRotation(out) {
		t.Fatal("resealed token still needs rotation")
	}
	got, err = box.Open(out, aad)
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatal("resealed plaintext mismatch")
	}
	sum := sha256.Sum256(cur)
	if !strings.HasPrefix(out, "v1:"+hex.EncodeToString(sum[:4])+":") {
		t.Fatal("reseal did not use the current key")
	}

	fresh, err := box.Seal(plain, aad)
	if err != nil {
		t.Fatal(err)
	}
	if box.NeedsRotation(fresh) {
		t.Fatal("current token reported rotation")
	}
	same, rotated, err := secretbox.Reseal(box, fresh, aad)
	if err != nil || rotated || same != fresh {
		t.Fatal("current token was rewritten")
	}
}

func TestNeedsRotationMalformedIsFalse(t *testing.T) {
	box, err := secretbox.New(fakeKey(0x11))
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"", "v2:00112233:aaaa", "v1:zz", "not-a-token"} {
		if box.NeedsRotation(token) {
			t.Fatal("malformed token reported rotation")
		}
	}
	other, err := secretbox.New(fakeKey(0x44))
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := other.Seal([]byte("x"), []byte("a"))
	if err != nil {
		t.Fatal(err)
	}
	if !box.NeedsRotation(sealed) {
		t.Fatal("unknown key id should need rotation")
	}
}

func TestIdenticalPreviousKeyIsIgnored(t *testing.T) {
	key := fakeKey(0x31)
	box, err := secretbox.New(key, key)
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := box.Seal([]byte("x"), []byte("a"))
	if err != nil {
		t.Fatal(err)
	}
	if box.NeedsRotation(sealed) {
		t.Fatal("duplicate current key looked like rotation")
	}
}

func TestDistinctKeyIDCollisionFails(t *testing.T) {
	a, b := collidingKeys(t)
	_, err := secretbox.New(a, b)
	if !errors.Is(err, secretbox.ErrInvalidKey) {
		t.Fatal("expected invalid key on key id collision")
	}
}

func TestResealTamperedPreviousFails(t *testing.T) {
	prev := fakeKey(0x51)
	cur := fakeKey(0x52)
	oldBox, err := secretbox.New(prev)
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := oldBox.Seal([]byte("x"), []byte("a"))
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(sealed, ":")
	raw, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 0x01
	parts[2] = base64.StdEncoding.EncodeToString(raw)
	tampered := strings.Join(parts, ":")
	box, err := secretbox.New(cur, prev)
	if err != nil {
		t.Fatal(err)
	}
	if !box.NeedsRotation(tampered) {
		t.Fatal("tampered previous token should still report rotation")
	}
	_, _, err = secretbox.Reseal(box, tampered, []byte("a"))
	if !errors.Is(err, secretbox.ErrDecrypt) {
		t.Fatal("expected decrypt failure")
	}
}

func collidingKeys(t *testing.T) ([]byte, []byte) {
	t.Helper()
	seen := make(map[string][]byte)
	buf := make([]byte, 32)
	for i := 0; i < 2_000_000; i++ {
		if _, err := rand.Read(buf); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(buf)
		id := hex.EncodeToString(sum[:4])
		prev, ok := seen[id]
		if ok && !bytes.Equal(prev, buf) {
			return append([]byte(nil), prev...), append([]byte(nil), buf...)
		}
		if !ok {
			seen[id] = append([]byte(nil), buf...)
		}
	}
	t.Fatal("no colliding key ids")
	return nil, nil
}
