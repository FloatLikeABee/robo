package handlers

import (
	"database/sql"
	"strings"
	"time"
)

// normalizeTranDateTimeColVal coerces API date strings (e.g. RFC3339) to MySQL DATETIME/DATE literals.
func normalizeTranDateTimeColVal(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil
		}
		if len(s) >= 10 {
			s = s[:10]
		}
		if d, err := time.Parse("2006-01-02", s); err == nil {
			return d.Format("2006-01-02")
		}
		return s
	default:
		return v
	}
}

// querySingleRowMap runs a query expected to return at most 1 row and returns it as a JSON-friendly map.
// MySQL driver typically returns []byte for strings; we convert those to string.
func querySingleRowMap(db *sql.DB, query string, args ...interface{}) (map[string]interface{}, bool, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, false, err
	}

	if !rows.Next() {
		return nil, false, nil
	}

	vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	if err := rows.Scan(ptrs...); err != nil {
		return nil, false, err
	}

	out := make(map[string]interface{}, len(cols))
	for i, c := range cols {
		v := vals[i]
		switch t := v.(type) {
		case nil:
			out[c] = nil
		case []byte:
			out[c] = string(t)
		case time.Time:
			out[c] = t.Format(time.RFC3339)
		default:
			out[c] = v
		}
	}
	return out, true, nil
}

