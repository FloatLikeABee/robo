package handlers

import (
	"database/sql"
	"strings"
	"time"
)

// nullTimeFromAny scans SQLite TEXT / MySQL TIME values into sql.NullTime semantics.
func nullTimeFromAny(v interface{}) sql.NullTime {
	if v == nil {
		return sql.NullTime{}
	}
	switch t := v.(type) {
	case time.Time:
		return sql.NullTime{Time: t, Valid: true}
	case *time.Time:
		if t == nil {
			return sql.NullTime{}
		}
		return sql.NullTime{Time: *t, Valid: true}
	case []byte:
		return parseFlexibleNullTime(string(t))
	case string:
		return parseFlexibleNullTime(t)
	default:
		return sql.NullTime{}
	}
}

func parseFlexibleNullTime(s string) sql.NullTime {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "null") {
		return sql.NullTime{}
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if tm, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return sql.NullTime{Time: tm, Valid: true}
		}
	}
	return sql.NullTime{}
}

// scanDestTime is a destination for rows.Scan that accepts string or time.Time.
type scanDestTime struct {
	nt *sql.NullTime
}

func (d scanDestTime) Scan(src interface{}) error {
	*d.nt = nullTimeFromAny(src)
	return nil
}
