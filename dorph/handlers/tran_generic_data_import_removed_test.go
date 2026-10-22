package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGenericDataImportRouteNotRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAPIRoutes(r, &Handlers{})

	var hasImport, hasExtract, hasCreate bool
	for _, rt := range r.Routes() {
		if rt.Path == "/api/tran/generic-data/import" && rt.Method == http.MethodPost {
			hasImport = true
		}
		if rt.Path == "/api/tran/generic-data/extract" && rt.Method == http.MethodPost {
			hasExtract = true
		}
		if rt.Path == "/api/tran/generic-data" && rt.Method == http.MethodPost {
			hasCreate = true
		}
	}
	if hasImport {
		t.Fatal("POST /api/tran/generic-data/import must not be registered")
	}
	if !hasExtract {
		t.Fatal("POST /api/tran/generic-data/extract must remain")
	}
	if !hasCreate {
		t.Fatal("POST /api/tran/generic-data must remain")
	}
}
