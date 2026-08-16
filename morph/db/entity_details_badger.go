package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// EntityDetailStore is the document API formerly backed by Mongo entity_details.
type EntityDetailStore interface {
	GetEntityDetailJSON(ctx context.Context, entity string, recordID int) (string, error)
	SetEntityDetailJSON(ctx context.Context, entity string, recordID int, body string) error
	DeleteEntityDetail(ctx context.Context, entity string, recordID int) error
}

const entityDetailKeyPrefix = "entity_detail:"

// BadgerEntityDetails stores entity detail JSON in a dedicated Badger directory.
type BadgerEntityDetails struct {
	db *badger.DB
}

// NewBadgerEntityDetails opens (or creates) a Badger DB at path for entity details.
func NewBadgerEntityDetails(path string) (*BadgerEntityDetails, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("entity details badger path is empty")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, fmt.Errorf("entity details mkdir: %w", err)
	}
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	bdb, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("entity details badger open: %w", err)
	}
	return &BadgerEntityDetails{db: bdb}, nil
}

func (s *BadgerEntityDetails) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func entityDetailKey(entity string, recordID int) []byte {
	return []byte(fmt.Sprintf("%s%s:%d", entityDetailKeyPrefix, strings.ToLower(strings.TrimSpace(entity)), recordID))
}

func entityDetailKeyVariants(entity string, recordID int) [][]byte {
	variants := entityLookupVariants(entity)
	keys := make([][]byte, 0, len(variants)+1)
	seen := map[string]struct{}{}
	for _, v := range variants {
		k := string(entityDetailKey(v, recordID))
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		keys = append(keys, []byte(k))
	}
	return keys
}

// GetEntityDetailJSON returns stored JSON text, or "{}" if missing.
func (s *BadgerEntityDetails) GetEntityDetailJSON(ctx context.Context, entity string, recordID int) (string, error) {
	_ = ctx
	if s == nil || s.db == nil {
		return "{}", nil
	}
	if entity == "" || recordID <= 0 {
		return "{}", nil
	}
	var out string
	err := s.db.View(func(txn *badger.Txn) error {
		for _, key := range entityDetailKeyVariants(entity, recordID) {
			item, err := txn.Get(key)
			if err == badger.ErrKeyNotFound {
				continue
			}
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				sval := strings.TrimSpace(string(val))
				if sval == "" {
					out = "{}"
					return nil
				}
				if !json.Valid([]byte(sval)) {
					out = "{}"
					return nil
				}
				out = sval
				return nil
			})
		}
		if out == "" {
			out = "{}"
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return out, nil
}

// SetEntityDetailJSON upserts JSON body. Empty string is stored as "{}".
func (s *BadgerEntityDetails) SetEntityDetailJSON(ctx context.Context, entity string, recordID int, body string) error {
	_ = ctx
	if s == nil || s.db == nil {
		return fmt.Errorf("entity details store is not configured")
	}
	if entity == "" || recordID <= 0 {
		return fmt.Errorf("invalid entity or record id")
	}
	if body == "" {
		body = "{}"
	}
	if !json.Valid([]byte(body)) {
		return fmt.Errorf("detail must be valid JSON")
	}
	key := entityDetailKey(entity, recordID)
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, []byte(body))
	})
}

// DeleteEntityDetail removes stored detail for entity/record (and lookup variants).
func (s *BadgerEntityDetails) DeleteEntityDetail(ctx context.Context, entity string, recordID int) error {
	_ = ctx
	if s == nil || s.db == nil {
		return nil
	}
	if entity == "" || recordID <= 0 {
		return nil
	}
	return s.db.Update(func(txn *badger.Txn) error {
		for _, key := range entityDetailKeyVariants(entity, recordID) {
			_ = txn.Delete(key)
		}
		return nil
	})
}

// CountEntityDetails returns approximate number of entity_detail keys (for migration verify).
func (s *BadgerEntityDetails) CountEntityDetails() (int, error) {
	if s == nil || s.db == nil {
		return 0, nil
	}
	n := 0
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = []byte(entityDetailKeyPrefix)
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			n++
		}
		return nil
	})
	return n, err
}

// Ensure TranMongo still satisfies EntityDetailStore for migrate / legacy.
var _ EntityDetailStore = (*TranMongo)(nil)
var _ EntityDetailStore = (*BadgerEntityDetails)(nil)
