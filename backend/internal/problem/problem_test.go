package problem

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEmitShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/shipments", nil)
	BadRequest(c, "invalid_body", "missing reference")

	if ct := w.Header().Get("Content-Type"); ct != ContentType {
		t.Fatalf("content-type: %q", ct)
	}
	var p Problem
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.Status != 400 || p.Code != "invalid_body" {
		t.Fatalf("unexpected: %+v", p)
	}
	if p.Type != TypeBase+"invalid_body" {
		t.Fatalf("type: %q", p.Type)
	}
	if p.Title != "Invalid body" {
		t.Fatalf("title: %q", p.Title)
	}
	if p.Instance != "/api/v1/shipments" {
		t.Fatalf("instance: %q", p.Instance)
	}
}
