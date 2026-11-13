package main

import (
	"testing"
	"time"

	"github.com/raoamogh/hireos/internal/platform/config"
)

func TestNewHTTPServerConfiguration(t *testing.T) {
	cfg := config.Config{
		App: config.AppConfig{
			Port: 8080,
		},
	}

	server := newHTTPServer(cfg)

	if server.Addr != ":8080" {
		t.Errorf("expected address :8080, got %q", server.Addr)
	}

	if server.ReadHeaderTimeout != 5*time.Second {
		t.Errorf(
			"expected ReadHeaderTimeout of 5 seconds, got %v",
			server.ReadHeaderTimeout,
		)
	}

	if server.IdleTimeout != 60*time.Second {
		t.Errorf(
			"expected IdleTimeout of 60 seconds, got %v",
			server.IdleTimeout,
		)
	}

	if server.Handler == nil {
		t.Error("expected HTTP handler to be initialized")
	}
}
