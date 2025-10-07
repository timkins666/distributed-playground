package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPServer(t *testing.T) {
	service := &Service{}
	config := &Config{
		Server: ServerConfig{Port: "8080"},
	}

	server := NewHTTPServer(service, config)
	if server.service != service {
		t.Error("service not set correctly")
	}
	if server.config != config {
		t.Error("config not set correctly")
	}
}

func TestHealthCheckHandler(t *testing.T) {
	service := &Service{}
	config := &Config{}
	server := NewHTTPServer(service, config)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.healthCheckHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}
}

func TestHealthCheckHandlerMethodNotAllowed(t *testing.T) {
	service := &Service{}
	config := &Config{}
	server := NewHTTPServer(service, config)

	req := httptest.NewRequest("POST", "/health", nil)
	w := httptest.NewRecorder()

	server.healthCheckHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestSetupRoutes(t *testing.T) {
	service := &Service{}
	config := &Config{}
	server := NewHTTPServer(service, config)

	mux := server.setupRoutes()
	if mux == nil {
		t.Error("mux should not be nil")
	}
}

func TestStop(t *testing.T) {
	service := &Service{}
	config := &Config{}
	server := NewHTTPServer(service, config)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := server.Stop(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}