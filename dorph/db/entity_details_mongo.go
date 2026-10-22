package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const entityDetailsColl = "entity_details"

// bodyValueToJSONString turns stored `body` into strict JSON text. Documents may be stored as a UTF-8 string
// (what SetEntityDetailJSON writes) or as a BSON object/array if edited in Compass.
func bodyValueToJSONString(v interface{}) (string, error) {
	if v == nil {
		return "", errors.New("nil body")
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return "", errors.New("empty body string")
		}
		if !json.Valid([]byte(s)) {
			return "", errors.New("invalid JSON in body string")
		}
		return s, nil
	case []byte:
		s := strings.TrimSpace(string(t))
		if s == "" {
			return "", errors.New("empty body bytes")
		}
		if !json.Valid([]byte(s)) {
			return "", errors.New("invalid JSON in body bytes")
		}
		return s, nil
	case primitive.Decimal128:
		return "", errors.New("unsupported Decimal128 body")
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return "", err
		}
		s := strings.TrimSpace(string(b))
		if !json.Valid([]byte(s)) {
			return "", errors.New("marshal produced invalid JSON")
		}
		return s, nil
	}
}

func uniqueRecordIDs(id int) []interface{} {
	if id <= 0 {
		return nil
	}
	raw := []interface{}{id, int32(id), int64(id), strconv.Itoa(id)}
	out := make([]interface{}, 0, len(raw))
	done := map[string]struct{}{}
	for _, v := range raw {
		key := fmt.Sprintf("%T:%#v", v, v)
		if _, ok := done[key]; ok {
			continue
		}
		done[key] = struct{}{}
		out = append(out, v)
	}
	return out
}

func dedupeStrings(vals []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// entityLookupVariants expands entity keys users may write in Compass (`staff`, `Staff`, etc.).
// Staff rows in MySQL are matched to Mongo docs with entity `"staff"` and/or mirrored `"employee"`.
func entityLookupVariants(entity string) []string {
	e := strings.TrimSpace(entity)
	if e == "" {
		return nil
	}
	el := strings.ToLower(e)
	var out []string
	switch el {
	case "staff", "employee":
		out = append(out, e, el, "staff", "Staff", "STAFF", "employee", "Employee", "EMPLOYEE")
	case "student":
		out = append(out, e, el, "student", "Student", "STUDENT")
	default:
		out = append(out, e, el)
	}
	return dedupeStrings(out)
}

func detailPayloadFromDoc(doc bson.M) (interface{}, bool) {
	for _, key := range []string{"body", "Body", "detail", "Detail", "payload", "Payload"} {
		v, ok := doc[key]
		if ok && v != nil {
			return v, true
		}
	}
	return nil, false
}

// entityDetailFilter matches Tran writes and common Compass/import variants (entity casing, staff|employee mirrors, record_id vs recordId).
func entityDetailFilter(entity string, recordID int) (bson.M, bool) {
	ids := uniqueRecordIDs(recordID)
	ents := entityLookupVariants(entity)
	if len(ids) == 0 || len(ents) == 0 {
		return nil, false
	}
	return bson.M{
		"$or": []bson.M{
			{
				"entity":    bson.M{"$in": ents},
				"record_id": bson.M{"$in": ids},
			},
			{
				"entity":   bson.M{"$in": ents},
				"recordId": bson.M{"$in": ids},
			},
		},
	}, true
}

// GetEntityDetailJSON returns stored JSON text for a record, or "{}" if missing / mongo unavailable.
func (m *TranMongo) GetEntityDetailJSON(ctx context.Context, entity string, recordID int) (string, error) {
	if m == nil || m.Database == nil {
		return "{}", nil
	}
	if entity == "" || recordID <= 0 {
		return "{}", nil
	}
	filter, ok := entityDetailFilter(entity, recordID)
	if !ok {
		return "{}", nil
	}

	var doc bson.M
	err := m.Database.Collection(entityDetailsColl).FindOne(ctx, filter).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "{}", nil
	}
	if err != nil {
		return "", err
	}
	raw, ok := detailPayloadFromDoc(doc)
	if !ok {
		return "{}", nil
	}
	js, err := bodyValueToJSONString(raw)
	if err != nil {
		return "{}", nil
	}
	return js, nil
}

// SetEntityDetailJSON stores JSON (object or any valid JSON). Empty string is stored as "{}".
func (m *TranMongo) SetEntityDetailJSON(ctx context.Context, entity string, recordID int, body string) error {
	if m == nil || m.Database == nil {
		return fmt.Errorf("MongoDB is not configured")
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
	now := time.Now()
	_, err := m.Database.Collection(entityDetailsColl).UpdateOne(
		ctx,
		bson.M{"entity": entity, "record_id": recordID},
		bson.M{"$set": bson.M{
			"entity":     entity,
			"record_id":  recordID,
			"body":       body,
			"updated_at": now,
		}},
		options.Update().SetUpsert(true),
	)
	return err
}

// DeleteEntityDetail removes stored detail (e.g. when MySQL row is deleted).
func (m *TranMongo) DeleteEntityDetail(ctx context.Context, entity string, recordID int) error {
	if m == nil || m.Database == nil {
		return nil
	}
	if entity == "" || recordID <= 0 {
		return nil
	}
	filter, ok := entityDetailFilter(entity, recordID)
	if !ok {
		return nil
	}
	_, err := m.Database.Collection(entityDetailsColl).DeleteMany(ctx, filter)
	return err
}
