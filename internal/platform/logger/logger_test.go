package logger

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNewWritesStructuredJSON(t *testing.T) {
	var output bytes.Buffer

	log := New("api", &output)

	log.Info(
		"application started",
		"port", 8080,
	)

	var record map[string]any

	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("expected valid JSON log record: %v", err)
	}

	if record["level"] != "INFO" {
		t.Errorf("expected INFO level, got %v", record["level"])
	}

	if record["msg"] != "application started" {
		t.Errorf("unexpected message: %v", record["msg"])
	}

	if record["service"] != "api" {
		t.Errorf("expected service api, got %v", record["service"])
	}

	if record["port"] != float64(8080) {
		t.Errorf("expected port 8080, got %v", record["port"])
	}

	if _, exists := record["time"]; !exists {
		t.Error("expected timestamp in log record")
	}
}

func TestNewSuppressesDebugLogs(t *testing.T) {
	var output bytes.Buffer

	log := New("api", &output)

	log.Debug("debug message")

	if output.Len() != 0 {
		t.Errorf("expected DEBUG message to be suppressed, got %s", output.String())
	}

	log.Info("info message")

	if output.Len() == 0 {
		t.Error("expected INFO message to be written")
	}
}

func TestNewUsesServiceName(t *testing.T) {
	var output bytes.Buffer

	log := New("worker", &output)

	log.Info("worker started")

	var record map[string]any

	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}

	if record["service"] != "worker" {
		t.Errorf("expected worker service, got %v", record["service"])
	}
}
