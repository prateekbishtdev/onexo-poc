package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func ok(context.Context) error   { return nil }
func fail(context.Context) error { return errors.New("connection refused") }

func TestLive(t *testing.T) {
	h := NewHealthHandler(map[string]Checker{"postgres": CheckerFunc(fail)})
	rec := httptest.NewRecorder()
	h.Live(rec, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("live should ignore dependencies, got %d", rec.Code)
	}
}

func TestReady(t *testing.T) {
	tests := []struct {
		name       string
		checks     map[string]Checker
		wantCode   int
		wantStatus string
	}{
		{"all up", map[string]Checker{"postgres": CheckerFunc(ok), "redis": CheckerFunc(ok)}, http.StatusOK, "ok"},
		{"redis down", map[string]Checker{"postgres": CheckerFunc(ok), "redis": CheckerFunc(fail)}, http.StatusServiceUnavailable, "degraded"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewHealthHandler(tt.checks).Ready(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d", rec.Code, tt.wantCode)
			}
			var body healthResponse
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", body.Status, tt.wantStatus)
			}
			if len(body.Checks) != len(tt.checks) {
				t.Errorf("got %d checks, want %d", len(body.Checks), len(tt.checks))
			}
		})
	}
}
