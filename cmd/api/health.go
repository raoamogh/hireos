package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type healthResponse struct {
	Status string `json:"status"`
}

type readinessResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			Status: "ok",
		})
	}
}

type databasePinger interface {
	Ping(context.Context) error
}

func readinessHandler(pool databasePinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, readinessResponse{
				Status:   "not_ready",
				Database: "unavailable",
			})
			return
		}

		writeJSON(w, http.StatusOK, readinessResponse{
			Status:   "ready",
			Database: "available",
		})
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}
