package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const sealedVersion = "v1"

var (
	// ErrUnknownKeyID is returned when a v1 token names a key id this box does not hold.
	ErrUnknownKeyID = errors.New("secretbox: unknown key id")
	// ErrUnsupportedVersion is returned when the token version is not v1.
	ErrUnsupportedVersion = errors.New("secretbox: unsupported sealed version")
	// ErrDecrypt is returned when a v1 payload fails to authenticate.
	ErrDecrypt = errors.New("secretbox: decryption failed")
	// ErrKeyMissing is returned when MORPH_SECRETS_KEY is unset or blank.
	ErrKeyMissing = errors.New("secretbox: MORPH_SECRETS_KEY is not set")
	// ErrInvalidKey is returned when a key is the wrong length, the wrong encoding, or ambiguous.
	ErrInvalidKey = errors.New("secretbox: invalid secrets key")
)

// Box seals and opens secrets with the current master key.
// Open selects a key by id from the current key and any previous keys.
// It does not trial-decrypt.
type Box interface {
	// Seal encrypts plaintext with the current key and returns a v1 token.
	Seal(plaintext []byte, aad []byte) (string, error)
	// Open decrypts a token. On failure it returns ErrUnknownKeyID, ErrUnsupportedVersion, or ErrDecrypt.
	Open(sealed string, aad []byte) ([]byte, error)
	// NeedsRotation reports whether sealed is a well-formed v1 token whose key id is not the current key.
	// It does not decrypt. Malformed tokens and unsupported versions return false.
	NeedsRotation(sealed string) bool
}

type aesBox struct {
	currentID string
	keys      map[string][]byte
}

// New builds a box from a raw 32-byte current key and optional decrypt-only previous keys.
// A previous key that is byte-identical to one already in the ring is ignored.
// Two different keys that share a key id fail with ErrInvalidKey.
func New(current []byte, previous ...[]byte) (Box, error) {
	if len(current) != 32 {
		return nil, fmt.Errorf("%w: key must be 32 bytes", ErrInvalidKey)
	}
	key := append([]byte(nil), current...)
	id := keyID(key)
	b := &aesBox{
		currentID: id,
		keys:      map[string][]byte{id: key},
	}
	for _, prev := range previous {
		if err := b.addPrevious(prev); err != nil {
			return nil, err
		}
	}
	return b, nil
}

func (b *aesBox) addPrevious(prev []byte) error {
	if len(prev) != 32 {
		return fmt.Errorf("%w: key must be 32 bytes", ErrInvalidKey)
	}
	id := keyID(prev)
	if existing, ok := b.keys[id]; ok {
		if subtle.ConstantTimeCompare(existing, prev) == 1 {
			return nil
		}
		return fmt.Errorf("%w: duplicate key id", ErrInvalidKey)
	}
	b.keys[id] = append([]byte(nil), prev...)
	return nil
}

func (b *aesBox) Seal(plaintext []byte, aad []byte) (string, error) {
	aead, err := newAEAD(b.keys[b.currentID])
	if err != nil {
		return "", errors.New("secretbox: seal failed")
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", errors.New("secretbox: nonce")
	}
	dst := make([]byte, len(nonce), len(nonce)+len(plaintext)+aead.Overhead())
	copy(dst, nonce)
	dst = aead.Seal(dst, nonce, plaintext, aad)
	return sealedVersion + ":" + b.currentID + ":" + base64.StdEncoding.EncodeToString(dst), nil
}

func (b *aesBox) Open(sealed string, aad []byte) ([]byte, error) {
	ver, id, payload, err := splitToken(sealed)
	if err != nil {
		return nil, err
	}
	if ver != sealedVersion {
		return nil, ErrUnsupportedVersion
	}
	key, ok := b.keys[id]
	if !ok {
		return nil, ErrUnknownKeyID
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, ErrDecrypt
	}
	aead, err := newAEAD(key)
	if err != nil {
		return nil, ErrDecrypt
	}
	ns := aead.NonceSize()
	if len(raw) < ns+aead.Overhead() {
		return nil, ErrDecrypt
	}
	plain, err := aead.Open(nil, raw[:ns], raw[ns:], aad)
	if err != nil {
		return nil, ErrDecrypt
	}
	return plain, nil
}

// NeedsRotation reports whether sealed is a well-formed v1 token whose key id is not the current key.
// It does not decrypt. Malformed tokens and unsupported versions return false.
func (b *aesBox) NeedsRotation(sealed string) bool {
	ver, id, _, err := splitToken(sealed)
	if err != nil || ver != sealedVersion {
		return false
	}
	return id != b.currentID
}

func splitToken(sealed string) (version, id, payload string, err error) {
	parts := strings.Split(strings.TrimSpace(sealed), ":")
	if len(parts) == 0 || parts[0] != sealedVersion {
		return "", "", "", ErrUnsupportedVersion
	}
	if len(parts) != 3 || !validKeyID(parts[1]) {
		return "", "", "", ErrDecrypt
	}
	return parts[0], parts[1], parts[2], nil
}

func validKeyID(id string) bool {
	if len(id) != 8 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f':
		default:
			return false
		}
	}
	return true
}

func keyID(key []byte) string {
	sum := sha256.Sum256(key)
	return hex.EncodeToString(sum[:4])
}

// Reseal opens sealed and, when it was not sealed with the current key, seals it again.
// An already-current token is returned unchanged. A token Open rejects is returned as that error.
func Reseal(b Box, sealed string, aad []byte) (string, bool, error) {
	if b == nil {
		return "", false, errors.New("secretbox: nil box")
	}
	plain, err := b.Open(sealed, aad)
	if err != nil {
		return "", false, err
	}
	defer wipe(plain)
	if !b.NeedsRotation(sealed) {
		return sealed, false, nil
	}
	out, err := b.Seal(plain, aad)
	if err != nil {
		return "", false, err
	}
	return out, true, nil
}

func wipe(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func newAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
