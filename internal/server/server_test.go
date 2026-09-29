package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prateekbishtdev/onexo-poc/internal/handler"
)

func TestRoutes(t *testing.T) {
	health := handler.NewHealthHandler(map[string]handler.Checker{
		"postgres": handler.CheckerFunc(func(context.Context) error { return nil }),
	})
	srv := httptest.NewServer(New("", health, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler)
	defer srv.Close()

	tests := []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/health", http.StatusOK},
		{http.MethodGet, "/health/live", http.StatusOK},
		{http.MethodPost, "/health", http.StatusMethodNotAllowed},
		{http.MethodGet, "/nope", http.StatusNotFound},
	}
	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, srv.URL+tt.path, nil)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != tt.want {
			t.Errorf("%s %s = %d, want %d", tt.method, tt.path, res.StatusCode, tt.want)
		}
	}
}
