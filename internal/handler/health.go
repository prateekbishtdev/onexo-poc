package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Checker reports whether a dependency is reachable.
type Checker interface {
	Check(ctx context.Context) error
}

// CheckerFunc adapts a plain function to the Checker interface.
type CheckerFunc func(ctx context.Context) error

func (f CheckerFunc) Check(ctx context.Context) error { return f(ctx) }

type HealthHandler struct {
	checks  map[string]Checker
	timeout time.Duration
	started time.Time
}

func NewHealthHandler(checks map[string]Checker) *HealthHandler {
	return &HealthHandler{
		checks:  checks,
		timeout: 2 * time.Second,
		started: time.Now(),
	}
}

type checkResult struct {
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

type healthResponse struct {
	Status    string                 `json:"status"`
	Uptime    string                 `json:"uptime"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]checkResult `json:"checks,omitempty"`
}

// Live reports that the process is running. It never touches dependencies.
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Uptime:    time.Since(h.started).Round(time.Second).String(),
		Timestamp: time.Now().UTC(),
	})
}

// Ready runs every dependency check concurrently and returns 503 if any fail.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results = make(map[string]checkResult, len(h.checks))
		healthy = true
	)
	for name, c := range h.checks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			err := c.Check(ctx)
			res := checkResult{Status: "up", LatencyMS: time.Since(start).Milliseconds()}
			if err != nil {
				res.Status = "down"
				res.Error = err.Error()
			}
			mu.Lock()
			results[name] = res
			if err != nil {
				healthy = false
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	status, code := "ok", http.StatusOK
	if !healthy {
		status, code = "degraded", http.StatusServiceUnavailable
	}
	writeJSON(w, code, healthResponse{
		Status:    status,
		Uptime:    time.Since(h.started).Round(time.Second).String(),
		Timestamp: time.Now().UTC(),
		Checks:    results,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
