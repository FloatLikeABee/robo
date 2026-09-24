package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGraphHealthOmitsURIAndRawError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handlers{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/graph/health", nil)
	h.GraphHealth(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["neo4j_uri"]; ok {
		t.Fatalf("response includes neo4j_uri: %s", w.Body.String())
	}
	if _, ok := body["neo4j_error"]; ok {
		t.Fatalf("response includes neo4j_error: %s", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "bolt://") || strings.Contains(w.Body.String(), "neo4j://") {
		t.Fatalf("response leaked a neo4j URI: %s", w.Body.String())
	}
}
