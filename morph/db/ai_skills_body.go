package db

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

const aiSkillBodyKeyPrefix = "ai_skill:body:"

// AISkillBody is the Badger-stored skill definition.
type AISkillBody struct {
	Instructions string                 `json:"instructions"`
	Body         string                 `json:"body,omitempty"` // alias accepted on write
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

func aiSkillBodyKey(id string) []byte {
	return []byte(aiSkillBodyKeyPrefix + strings.TrimSpace(id))
}

// PutAISkillBody stores skill body JSON keyed by skill id.
func (d *DB) PutAISkillBody(id string, body AISkillBody) error {
	if d == nil || d.badgerDB == nil {
		return fmt.Errorf("badger not available")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("skill id required")
	}
	instructions := strings.TrimSpace(body.Instructions)
	if instructions == "" {
		instructions = strings.TrimSpace(body.Body)
	}
	body.Instructions = instructions
	body.Body = ""
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Set(aiSkillBodyKey(id), raw)
	})
}

// GetAISkillBody loads skill body JSON; returns empty body if missing.
func (d *DB) GetAISkillBody(id string) (AISkillBody, error) {
	var out AISkillBody
	if d == nil || d.badgerDB == nil {
		return out, fmt.Errorf("badger not available")
	}
	id = strings.TrimSpace(id)
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(aiSkillBodyKey(id))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &out)
		})
	})
	if err != nil {
		return out, err
	}
	if strings.TrimSpace(out.Instructions) == "" && strings.TrimSpace(out.Body) != "" {
		out.Instructions = out.Body
	}
	return out, nil
}

// DeleteAISkillBody removes the Badger body for a skill id.
func (d *DB) DeleteAISkillBody(id string) error {
	if d == nil || d.badgerDB == nil {
		return nil
	}
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		err := txn.Delete(aiSkillBodyKey(id))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		return err
	})
}
