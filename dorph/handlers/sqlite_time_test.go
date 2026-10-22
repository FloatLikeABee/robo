package handlers

import "testing"

func TestParseFlexibleNullTimeDateTimeLocal(t *testing.T) {
	got := parseFlexibleNullTime("2026-08-30T19:16")
	if !got.Valid {
		t.Fatal("datetime-local start_at should parse")
	}
	if got.Time.Year() != 2026 || got.Time.Month() != 8 || got.Time.Day() != 30 {
		t.Fatalf("got %v", got.Time)
	}
	if got.Time.Hour() != 19 || got.Time.Minute() != 16 {
		t.Fatalf("got %v", got.Time)
	}
	if parseFlexibleNullTime("").Valid {
		t.Fatal("empty should be invalid")
	}
}
