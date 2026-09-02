package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type memEntityDetails struct {
	mu sync.Mutex
	m  map[string]string
}

func (s *memEntityDetails) key(entity string, recordID int) string {
	return fmt.Sprintf("%s:%d", entity, recordID)
}

func (s *memEntityDetails) GetEntityDetailJSON(_ context.Context, entity string, recordID int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.m == nil {
		return "{}", nil
	}
	v, ok := s.m[s.key(entity, recordID)]
	if !ok {
		return "{}", nil
	}
	return v, nil
}

func (s *memEntityDetails) SetEntityDetailJSON(_ context.Context, entity string, recordID int, body string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.m == nil {
		s.m = map[string]string{}
	}
	s.m[s.key(entity, recordID)] = body
	return nil
}

func (s *memEntityDetails) DeleteEntityDetail(_ context.Context, entity string, recordID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.m != nil {
		delete(s.m, s.key(entity, recordID))
	}
	return nil
}

func openGenericDataDB(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", "file:generic-data-"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	_, err = sqlDB.Exec(`CREATE TABLE generic_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		source_type TEXT NOT NULL,
		source_filename TEXT NULL,
		record_count INTEGER NOT NULL DEFAULT 0,
		description TEXT NULL,
		ai_analysis TEXT NULL,
		created_on TEXT DEFAULT CURRENT_TIMESTAMP,
		last_updated TEXT DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatal(err)
	}
	return sqlDB
}

func genericDataReviewRouter(sqlDB *sql.DB, details morphdb.EntityDetailStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &Handlers{
		TranMySQL:     &morphdb.TranSQL{DB: sqlDB},
		EntityDetails: details,
	}
	r := gin.New()
	r.POST("/api/tran/generic-data/extract", h.ExtractGenericDataFile)
	r.POST("/api/tran/generic-data", h.CreateGenericData)
	return r
}

func postGenericDataExtract(r http.Handler, filename string, body []byte) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		panic(err)
	}
	_, _ = fw.Write(body)
	_ = mw.Close()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/generic-data/extract", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)
	return w
}

func genericDataRowCount(t *testing.T, sqlDB *sql.DB) int {
	t.Helper()
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM generic_data`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestExtractGenericDataFileCSVReturnsDraftsWithoutInsert(t *testing.T) {
	sqlDB := openGenericDataDB(t)
	r := genericDataReviewRouter(sqlDB, nil)

	w := postGenericDataExtract(r, "fleet.csv", []byte("name,qty\nbus,2\nvan,1\n"))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["json"] == nil {
		t.Fatalf("expected json draft, got %s", w.Body.String())
	}
	md, _ := out["markdown"].(string)
	if !strings.Contains(md, "bus") {
		t.Fatalf("expected markdown table to include row values, got %q", md)
	}
	if out["source_type"] != "csv" {
		t.Fatalf("source_type=%v", out["source_type"])
	}
	if out["filename"] != "fleet.csv" {
		t.Fatalf("filename=%v", out["filename"])
	}
	if genericDataRowCount(t, sqlDB) != 0 {
		t.Fatal("extract must not insert a generic_data row")
	}
}

func TestExtractGenericDataFileAIDownUsesParseSeed(t *testing.T) {
	sqlDB := openGenericDataDB(t)
	r := genericDataReviewRouter(sqlDB, nil) // aiService is nil

	w := postGenericDataExtract(r, "notes.md", []byte("# Hello\n\nBody of the note.\n"))
	if w.Code != http.StatusOK {
		t.Fatalf("markdown extract status %d body %s", w.Code, w.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	md, _ := out["markdown"].(string)
	if !strings.Contains(md, "Hello") {
		t.Fatalf("expected parse-seeded markdown, got %q", md)
	}
	if genericDataRowCount(t, sqlDB) != 0 {
		t.Fatal("extract must not insert a row when AI is down")
	}
}

func TestExtractGenericDataFileEmptyPDFAIDownErrors(t *testing.T) {
	sqlDB := openGenericDataDB(t)
	r := genericDataReviewRouter(sqlDB, nil)

	w := postGenericDataExtract(r, "scan.pdf", []byte("not a pdf"))
	if w.Code != http.StatusBadRequest && w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected extract error, got %d body %s", w.Code, w.Body.String())
	}
	if genericDataRowCount(t, sqlDB) != 0 {
		t.Fatal("failed extract must not insert a row")
	}
}

func TestCreateGenericDataPersistsExtractShapedDetail(t *testing.T) {
	sqlDB := openGenericDataDB(t)
	details := &memEntityDetails{}
	r := genericDataReviewRouter(sqlDB, details)

	body := `{
		"title": "Fleet excerpt",
		"source_type": "csv",
		"source_filename": "fleet.csv",
		"detail": {
			"columns": ["name", "qty"],
			"rows": [{"name": "bus", "qty": "2"}],
			"content_markdown": "| name | qty |\n| --- | --- |\n| bus | 2 |"
		}
	}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/generic-data", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	id, _ := out["id"].(float64)
	if id <= 0 {
		t.Fatalf("expected id, got %s", w.Body.String())
	}
	raw, err := details.GetEntityDetailJSON(context.Background(), entityKeyGenericData, int(id))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(raw, "content_markdown") || !strings.Contains(raw, "bus") {
		t.Fatalf("detail missing json+markdown: %s", raw)
	}
}

func TestCreateGenericDataRejectsInvalidDetailJSON(t *testing.T) {
	sqlDB := openGenericDataDB(t)
	r := genericDataReviewRouter(sqlDB, &memEntityDetails{})

	body := `{"title":"Bad","source_type":"json","detail":"not json"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/generic-data", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if genericDataRowCount(t, sqlDB) != 0 {
		t.Fatal("invalid JSON must not create a row")
	}
}

func TestGenericDataExcerptIncludesContentMarkdown(t *testing.T) {
	m := map[string]interface{}{
		"source_type": "json",
		"detail": map[string]interface{}{
			"payload":          map[string]interface{}{"a": 1},
			"content_markdown": "# Hello from file\n\nBody",
		},
	}
	got := genericDataExcerptForAnalysis(m)
	if !strings.Contains(got, "Hello from file") {
		t.Fatalf("excerpt missing content_markdown: %s", got)
	}
}
