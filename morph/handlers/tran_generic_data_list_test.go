package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

func TestListGenericDataReturnsEveryRow(t *testing.T) {
	sqlDB := openGenericDataDB(t)
	for i, title := range []string{"one", "two", "three", "four"} {
		_, err := sqlDB.Exec(
			`INSERT INTO generic_data (title, source_type, last_updated) VALUES (?, 'json', ?)`,
			title,
			fmt.Sprintf("2026-01-%02d 00:00:00", i+1),
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	gin.SetMode(gin.TestMode)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	r := gin.New()
	r.GET("/api/tran/generic-data", h.ListGenericData)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tran/generic-data", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var list []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 4 {
		t.Fatalf("got %d rows, want 4: %s", len(list), w.Body.String())
	}
	ids := map[int]bool{}
	for _, row := range list {
		id := int(row["id"].(float64))
		ids[id] = true
	}
	for want := 1; want <= 4; want++ {
		if !ids[want] {
			t.Fatalf("missing id %d in %v", want, ids)
		}
	}
	firstID := int(list[0]["id"].(float64))
	if firstID != 4 {
		t.Fatalf("newest id should be first, got id %d title %v", firstID, list[0]["title"])
	}
}

func TestCreateGenericDataThenListIncludesIt(t *testing.T) {
	sqlDB := openGenericDataDB(t)
	_, err := sqlDB.Exec(`INSERT INTO generic_data (title, source_type, last_updated) VALUES ('old', 'json', '2020-01-01 00:00:00')`)
	if err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}, EntityDetails: &memEntityDetails{}}
	r := gin.New()
	r.GET("/api/tran/generic-data", h.ListGenericData)
	r.POST("/api/tran/generic-data", h.CreateGenericData)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/generic-data", strings.NewReader(`{"title":"new-import","source_type":"json"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create status %d body %s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	newID := int(created["id"].(float64))
	if newID <= 0 {
		t.Fatalf("expected created id, got %s", w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/tran/generic-data", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status %d body %s", w.Code, w.Body.String())
	}
	var list []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	foundAt := -1
	for i, row := range list {
		if int(row["id"].(float64)) == newID {
			foundAt = i
			break
		}
	}
	if foundAt < 0 {
		t.Fatalf("created id %d missing from list %s", newID, w.Body.String())
	}
	if foundAt != 0 {
		t.Fatalf("new row should be first, index %d list %s", foundAt, w.Body.String())
	}
}

func TestGetGenericDataByIDStillWorks(t *testing.T) {
	sqlDB := openGenericDataDB(t)
	res, err := sqlDB.Exec(`INSERT INTO generic_data (title, source_type) VALUES ('solo', 'csv')`)
	if err != nil {
		t.Fatal(err)
	}
	id64, _ := res.LastInsertId()

	gin.SetMode(gin.TestMode)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	r := gin.New()
	r.GET("/api/tran/generic-data/:id", h.GetGenericData)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tran/generic-data/%d", id64), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var row map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &row); err != nil {
		t.Fatal(err)
	}
	if int(row["id"].(float64)) != int(id64) {
		t.Fatalf("id=%v want %d", row["id"], id64)
	}
	if row["title"] != "solo" {
		t.Fatalf("title=%v", row["title"])
	}
}
