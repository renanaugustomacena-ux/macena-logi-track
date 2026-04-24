package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/handlers"
)

// TestHealthEndpoint is a table-driven smoke test for GET /api/health.
// The handler under test is constructed with nil repositories; the
// implementation must tolerate absent dependencies and report them as
// "unknown" or "degraded" without panicking.
//
// Integration tests that exercise the real MongoDB and Redis clients
// live under tests/integration/ (not yet wired into this template).
func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	cfg.App.Name = "logitrack-test"
	cfg.App.Version = "0.0.0-test"

	h := handlers.NewHealthHandler(cfg, nil, nil)

	cases := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantFields []string
	}{
		{
			name:       "GET /api/health returns 200 with correct shape",
			method:     http.MethodGet,
			path:       "/api/health",
			wantStatus: http.StatusOK,
			wantFields: []string{"status", "service", "version", "uptime_seconds", "time", "dependencies"},
		},
	}

	r := gin.New()
	r.GET("/api/health", h.Get)

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
			defer cancel()
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", w.Code, tc.wantStatus, w.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode body: %v; raw=%s", err, w.Body.String())
			}
			for _, f := range tc.wantFields {
				if _, ok := body[f]; !ok {
					t.Errorf("missing field %q in response: %v", f, body)
				}
			}
			if svc, _ := body["service"].(string); svc != "logitrack-test" {
				t.Errorf("service = %q, want logitrack-test", svc)
			}
		})
	}
}
