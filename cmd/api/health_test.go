package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raoamogh/hireos/internal/platform/config"
)

type fakeDatabasePinger struct {
	err error
}

func (f fakeDatabasePinger) Ping(context.Context) error {
	return f.err
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d",
			http.StatusOK, rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected application/json, got %q", got)
	}

	var response healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected status ok, got %q", response.Status)
	}
}

func TestReadinessHandler(t *testing.T) {
	tests := []struct {
		name           string
		databaseErr    error
		expectedStatus int
		expectedHealth string
		expectedDB     string
	}{
		{
			name:           "database available",
			databaseErr:    nil,
			expectedStatus: http.StatusOK,
			expectedHealth: "ready",
			expectedDB:     "available",
		},
		{
			name:           "database unavailable",
			databaseErr:    errors.New("database connection failed"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "not_ready",
			expectedDB:     "unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := fakeDatabasePinger{err: tt.databaseErr}

			req := httptest.NewRequest(http.MethodGet, "/ready", nil)
			rec := httptest.NewRecorder()

			readinessHandler(pool).ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d",
					tt.expectedStatus, rec.Code)
			}

			var response readinessResponse
			if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if response.Status != tt.expectedHealth {
				t.Errorf("expected status %q, got %q",
					tt.expectedHealth, response.Status)
			}

			if response.Database != tt.expectedDB {
				t.Errorf("expected database %q, got %q",
					tt.expectedDB, response.Database)
			}
		})
	}
}

func TestHTTPServerRoutes(t *testing.T) {
	cfg := config.Config{
		App: config.AppConfig{
			Port: 8080,
		},
	}

	pool := fakeDatabasePinger{}

	server := newHTTPServer(cfg, pool)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "health route",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "readiness route",
			method:         http.MethodGet,
			path:           "/ready",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "health route rejects POST",
			method:         http.MethodPost,
			path:           "/health",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "readiness route rejects POST",
			method:         http.MethodPost,
			path:           "/ready",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			server.Handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d",
					tt.expectedStatus, rec.Code)
			}
		})
	}
}
