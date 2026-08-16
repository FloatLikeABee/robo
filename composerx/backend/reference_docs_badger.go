package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

const referenceDocKeyPrefix = "ref_doc:"

var errReferenceDocNotFound = errors.New("reference document not found")

// ReferenceDocsStore persists reference documents (RAG library) in Badger.
type ReferenceDocsStore struct {
	db *badger.DB
}

func NewReferenceDocsStore(db *badger.DB) *ReferenceDocsStore {
	return &ReferenceDocsStore{db: db}
}

func referenceDocKey(id string) []byte {
	return []byte(referenceDocKeyPrefix + id)
}

func openComposerXBadger(path string) (*badger.DB, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("badger path is empty")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, fmt.Errorf("badger mkdir: %w", err)
	}
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("badger open: %w", err)
	}
	return db, nil
}

func (s *ReferenceDocsStore) Put(ctx context.Context, doc *ReferenceDoc) error {
	_ = ctx
	if s == nil || s.db == nil {
		return errors.New("reference docs store not configured")
	}
	if doc == nil || strings.TrimSpace(doc.ID) == "" {
		return errors.New("reference doc id required")
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(referenceDocKey(doc.ID), raw)
	})
}

func (s *ReferenceDocsStore) Get(ctx context.Context, id string) (*ReferenceDoc, error) {
	_ = ctx
	if s == nil || s.db == nil {
		return nil, errors.New("reference docs store not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errReferenceDocNotFound
	}
	var doc ReferenceDoc
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(referenceDocKey(id))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &doc)
		})
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, errReferenceDocNotFound
	}
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *ReferenceDocsStore) Delete(ctx context.Context, id string) error {
	_ = ctx
	if s == nil || s.db == nil {
		return errors.New("reference docs store not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return errReferenceDocNotFound
	}
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(referenceDocKey(id))
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return errReferenceDocNotFound
	}
	return err
}

func (s *ReferenceDocsStore) List(ctx context.Context, limit int) ([]ReferenceDoc, error) {
	_ = ctx
	if s == nil || s.db == nil {
		return nil, errors.New("reference docs store not configured")
	}
	if limit <= 0 {
		limit = 500
	}
	docs := make([]ReferenceDoc, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()
		prefix := []byte(referenceDocKeyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var doc ReferenceDoc
			if err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &doc)
			}); err != nil {
				return err
			}
			docs = append(docs, doc)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].CreatedAt.After(docs[j].CreatedAt)
	})
	if len(docs) > limit {
		docs = docs[:limit]
	}
	return docs, nil
}

func (s *ReferenceDocsStore) GetByIDs(ctx context.Context, ids []string) ([]ReferenceDoc, error) {
	_ = ctx
	if len(ids) == 0 {
		return nil, nil
	}
	out := make([]ReferenceDoc, 0, len(ids))
	for _, id := range ids {
		doc, err := s.Get(ctx, id)
		if errors.Is(err, errReferenceDocNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, *doc)
	}
	return out, nil
}

func (s *ReferenceDocsStore) Search(ctx context.Context, query string, limit int) ([]ReferenceDoc, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}
	all, err := s.List(ctx, 500)
	if err != nil {
		return nil, err
	}
	matched := make([]ReferenceDoc, 0)
	for _, d := range all {
		name := strings.ToLower(d.Name)
		text := strings.ToLower(d.TextContent)
		if strings.Contains(name, q) || strings.Contains(text, q) {
			matched = append(matched, d)
			if len(matched) >= limit {
				break
			}
		}
	}
	return matched, nil
}
