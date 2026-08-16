package mongo

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/formsx/backend/internal/config"
)

// ErrNoDocuments mirrors mongo.ErrNoDocuments for handler compatibility.
var ErrNoDocuments = errors.New("document not found")

// Store is the shared Badger DB for former Mongo collections.
type Store struct {
	db *badger.DB
}

// NewStore opens (or creates) Badger at FORMSX_BADGER_PATH.
func NewStore(cfg *config.Config) (*Store, error) {
	path := strings.TrimSpace(cfg.FormsXBadgerPath)
	if path == "" {
		return nil, fmt.Errorf("FORMSX_BADGER_PATH is empty")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, fmt.Errorf("badger mkdir: %w", err)
	}
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	bdb, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("badger open: %w", err)
	}
	return &Store{db: bdb}, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func putJSON(txn *badger.Txn, key []byte, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return txn.Set(key, raw)
}

func getJSON(txn *badger.Txn, key []byte, dest any) error {
	item, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return ErrNoDocuments
		}
		return err
	}
	return item.Value(func(val []byte) error {
		return json.Unmarshal(val, dest)
	})
}

func scanPrefix(db *badger.DB, prefix []byte, fn func(key, val []byte) error) error {
	return db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.KeyCopy(nil)
			if err := item.Value(func(val []byte) error {
				return fn(k, val)
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func containsFold(hay, needle string) bool {
	if needle == "" {
		return true
	}
	return strings.Contains(strings.ToLower(hay), strings.ToLower(needle))
}
