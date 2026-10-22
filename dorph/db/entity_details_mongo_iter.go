package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// CountEntityDetails returns the number of documents in entity_details.
func (m *TranMongo) CountEntityDetails(ctx context.Context) (int64, error) {
	if m == nil || m.Database == nil {
		return 0, nil
	}
	n, err := m.Database.Collection(entityDetailsColl).CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	return n, nil
}

// EntityDetailDoc is one Mongo entity_details row for migration.
type EntityDetailDoc struct {
	Entity   string
	RecordID int
	Body     string
}

// ForEachEntityDetail walks all entity_details documents (read-only).
// Invalid / empty bodies are reported as "{}" so empty-detail semantics are preserved.
func (m *TranMongo) ForEachEntityDetail(ctx context.Context, fn func(EntityDetailDoc) error) (int, error) {
	if m == nil || m.Database == nil {
		return 0, nil
	}
	cur, err := m.Database.Collection(entityDetailsColl).Find(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	defer cur.Close(ctx)

	n := 0
	for cur.Next(ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			return n, err
		}
		entity, recordID, ok := parseEntityDetailIdentity(doc)
		if !ok {
			continue
		}
		body := "{}"
		if raw, ok := detailPayloadFromDoc(doc); ok {
			if js, err := bodyValueToJSONString(raw); err == nil {
				body = js
			}
		}
		if err := fn(EntityDetailDoc{Entity: entity, RecordID: recordID, Body: body}); err != nil {
			return n, err
		}
		n++
	}
	if err := cur.Err(); err != nil {
		return n, err
	}
	return n, nil
}

func parseEntityDetailIdentity(doc bson.M) (entity string, recordID int, ok bool) {
	entity = strings.TrimSpace(fmt.Sprint(doc["entity"]))
	if entity == "" || entity == "<nil>" {
		return "", 0, false
	}
	var raw interface{}
	if v, has := doc["record_id"]; has && v != nil {
		raw = v
	} else if v, has := doc["recordId"]; has && v != nil {
		raw = v
	} else {
		return "", 0, false
	}
	id, err := coerceIntID(raw)
	if err != nil || id <= 0 {
		return "", 0, false
	}
	return entity, id, true
}

func coerceIntID(v interface{}) (int, error) {
	switch t := v.(type) {
	case int:
		return t, nil
	case int32:
		return int(t), nil
	case int64:
		return int(t), nil
	case float64:
		return int(t), nil
	case string:
		return strconv.Atoi(strings.TrimSpace(t))
	default:
		s := strings.TrimSpace(fmt.Sprint(t))
		if s == "" || s == "<nil>" {
			return 0, fmt.Errorf("empty id")
		}
		return strconv.Atoi(s)
	}
}
