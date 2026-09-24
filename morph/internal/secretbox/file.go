package secretbox

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ErrKeyFile is returned when a dev key file cannot be created or used.
// The error text does not include key material.
var ErrKeyFile = errors.New("secretbox: secrets key file")

// LoadOrCreateDevKeyFile reads a standard-base64 32-byte key from path.
// When path does not exist, it creates a regular file with mode 0600 and
// reports created=true. An existing file is never overwritten.
func LoadOrCreateDevKeyFile(path string) ([]byte, bool, error) {
	if path == "" {
		return nil, false, fmt.Errorf("%w: path is empty", ErrKeyFile)
	}
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return createDevKeyFile(path)
		}
		return nil, false, fmt.Errorf("%w: key file is not available", ErrKeyFile)
	}
	key, err := readDevKeyFile(path, info)
	if err != nil {
		return nil, false, err
	}
	return key, false, nil
}

// Resolve returns the box the process should use.
// production true, or MORPH_SECRETS_KEY set: FromEnv only, and no key file is created or read.
// production false with both env vars unset: load or create devKeyPath (mode 0600).
// createdDevKey is true only when this call wrote that file. The caller logs a warning
// that does not include the key. This function does not log.
// A previous-key list without a current key is ErrInvalidKey and does not create a file.
func Resolve(production bool, devKeyPath string) (Box, bool, error) {
	current := strings.TrimSpace(os.Getenv("MORPH_SECRETS_KEY"))
	previous := strings.TrimSpace(os.Getenv("MORPH_SECRETS_KEY_PREVIOUS"))
	if production || current != "" {
		box, err := FromEnv()
		return box, false, err
	}
	if previous != "" {
		return nil, false, fmt.Errorf("%w: MORPH_SECRETS_KEY_PREVIOUS is set but MORPH_SECRETS_KEY is not", ErrInvalidKey)
	}
	key, created, err := LoadOrCreateDevKeyFile(devKeyPath)
	if err != nil {
		return nil, false, err
	}
	box, err := New(key)
	if err != nil {
		return nil, created, err
	}
	return box, created, nil
}

func createDevKeyFile(path string) ([]byte, bool, error) {
	parent, err := os.Stat(filepath.Dir(path))
	if err != nil || !parent.IsDir() {
		return nil, false, fmt.Errorf("%w: key file directory does not exist", ErrKeyFile)
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, false, fmt.Errorf("%w: could not generate a key", ErrKeyFile)
	}
	encoded := base64.StdEncoding.EncodeToString(raw) + "\n"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if os.IsExist(err) {
			info, lerr := os.Lstat(path)
			if lerr != nil {
				return nil, false, fmt.Errorf("%w: key file is not available", ErrKeyFile)
			}
			key, rerr := readDevKeyFile(path, info)
			return key, false, rerr
		}
		return nil, false, fmt.Errorf("%w: could not create key file", ErrKeyFile)
	}
	defer f.Close()
	n, err := f.WriteString(encoded)
	if err != nil || n != len(encoded) {
		return nil, false, fmt.Errorf("%w: could not write key file", ErrKeyFile)
	}
	if err := f.Sync(); err != nil {
		return nil, false, fmt.Errorf("%w: could not write key file", ErrKeyFile)
	}
	if err := f.Chmod(0o600); err != nil {
		return nil, false, fmt.Errorf("%w: could not set key file mode", ErrKeyFile)
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return nil, false, fmt.Errorf("%w: key file mode is not private", ErrKeyFile)
	}
	if err := syncDir(filepath.Dir(path)); err != nil {
		return nil, false, fmt.Errorf("%w: could not write key file", ErrKeyFile)
	}
	return append([]byte(nil), raw...), true, nil
}

func syncDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

func readDevKeyFile(path string, lstated os.FileInfo) ([]byte, error) {
	if lstated.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%w: key file must not be a symlink", ErrKeyFile)
	}
	if !lstated.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: key file must be a regular file", ErrKeyFile)
	}
	if lstated.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("%w: key file is readable by group or others", ErrKeyFile)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: could not open key file", ErrKeyFile)
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(lstated, opened) {
		return nil, fmt.Errorf("%w: key file changed while opening", ErrKeyFile)
	}
	if !opened.Mode().IsRegular() || opened.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("%w: key file is not a private regular file", ErrKeyFile)
	}
	body, err := io.ReadAll(io.LimitReader(f, 128))
	if err != nil {
		return nil, fmt.Errorf("%w: could not read key file", ErrKeyFile)
	}
	var extra [1]byte
	n, rerr := f.Read(extra[:])
	if n > 0 || (rerr != nil && rerr != io.EOF) {
		return nil, fmt.Errorf("%w: key file must contain 32 bytes of standard base64", ErrKeyFile)
	}
	raw, err := decodeKey("key file", string(body))
	if err != nil {
		return nil, fmt.Errorf("%w: key file must contain 32 bytes of standard base64", ErrKeyFile)
	}
	return raw, nil
}
