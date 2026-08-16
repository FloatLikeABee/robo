package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

const emailBodyKeyPrefix = "email_body:"

// EmailContentStore persists large markdown/HTML bodies in Badger.
type EmailContentStore struct {
	db *badger.DB
}

type emailBodyDoc struct {
	ID                string    `json:"id"`
	Markdown          string    `json:"markdown,omitempty"`
	HTML              string    `json:"html,omitempty"`
	ComposerAISession string    `json:"composer_ai_session,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func NewEmailContentStore(db *badger.DB) *EmailContentStore {
	return &EmailContentStore{db: db}
}

func emailBodyKey(id string) []byte {
	return []byte(emailBodyKeyPrefix + id)
}

func (s *EmailContentStore) InsertMarkdown(ctx context.Context, markdown, composerAISessionJSON string) (string, error) {
	_ = ctx
	if s == nil || s.db == nil {
		return "", errors.New("email content store not configured")
	}
	id := uuid.NewString()
	now := time.Now().UTC()
	doc := emailBodyDoc{
		ID:                id,
		Markdown:          markdown,
		ComposerAISession: composerAISessionJSON,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.put(doc); err != nil {
		return "", err
	}
	return id, nil
}

func (s *EmailContentStore) UpdateMarkdown(ctx context.Context, id, markdown, composerAISessionJSON string) error {
	_ = ctx
	doc, err := s.get(id)
	if err != nil {
		return err
	}
	doc.Markdown = markdown
	doc.ComposerAISession = composerAISessionJSON
	doc.UpdatedAt = time.Now().UTC()
	return s.put(*doc)
}

func (s *EmailContentStore) GetMarkdown(ctx context.Context, id string) (markdown string, composerAISessionJSON string, err error) {
	_ = ctx
	doc, err := s.get(id)
	if err != nil {
		return "", "", err
	}
	body := doc.Markdown
	if body == "" {
		body = doc.HTML
	}
	return body, doc.ComposerAISession, nil
}

func (s *EmailContentStore) InsertHTML(ctx context.Context, html, composerAISessionJSON string) (string, error) {
	_ = ctx
	if s == nil || s.db == nil {
		return "", errors.New("email content store not configured")
	}
	id := uuid.NewString()
	now := time.Now().UTC()
	doc := emailBodyDoc{
		ID:                id,
		HTML:              html,
		ComposerAISession: composerAISessionJSON,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.put(doc); err != nil {
		return "", err
	}
	return id, nil
}

func (s *EmailContentStore) UpdateHTML(ctx context.Context, id, html, composerAISessionJSON string) error {
	_ = ctx
	doc, err := s.get(id)
	if err != nil {
		return err
	}
	doc.HTML = html
	doc.ComposerAISession = composerAISessionJSON
	doc.UpdatedAt = time.Now().UTC()
	return s.put(*doc)
}

func (s *EmailContentStore) GetHTML(ctx context.Context, id string) (html string, composerAISessionJSON string, err error) {
	_ = ctx
	doc, err := s.get(id)
	if err != nil {
		return "", "", err
	}
	return doc.HTML, doc.ComposerAISession, nil
}

func (s *EmailContentStore) DeleteByHexID(ctx context.Context, id string) error {
	_ = ctx
	if s == nil || s.db == nil {
		return errors.New("email content store not configured")
	}
	id = trimID(id)
	if id == "" {
		return errors.New("empty content id")
	}
	return s.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(emailBodyKey(id))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	})
}

func (s *EmailContentStore) put(doc emailBodyDoc) error {
	if s == nil || s.db == nil {
		return errors.New("email content store not configured")
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(emailBodyKey(doc.ID), raw)
	})
}

func (s *EmailContentStore) get(id string) (*emailBodyDoc, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("email content store not configured")
	}
	id = trimID(id)
	if id == "" {
		return nil, errors.New("empty content id")
	}
	var doc emailBodyDoc
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(emailBodyKey(id))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &doc)
		})
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, fmt.Errorf("content %s: %w", id, err)
	}
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func trimID(id string) string {
	return strings.TrimSpace(id)
}
