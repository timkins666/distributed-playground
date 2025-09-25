package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type HTTPServer struct {
	server  *http.Server
	service *Service
	config  *Config
}

// creates a new HTTP server instance
func NewHTTPServer(service *Service, config *Config) *HTTPServer {
	return &HTTPServer{
		service: service,
		config:  config,
	}
}

func (h *HTTPServer) Start(ctx context.Context) error {
	mux := h.setupRoutes()

	h.server = &http.Server{
		Addr:         ":" + h.config.Server.Port,
		Handler:      h.setupMiddleware(mux),
		ReadTimeout:  h.config.Server.ReadTimeout,
		WriteTimeout: h.config.Server.WriteTimeout,
		IdleTimeout:  h.config.Server.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		h.service.appCtx.logger.Printf("Account service starting on port %s", h.config.Server.Port)
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.service.appCtx.logger.Printf("Server failed to start: %v", err)
		}
	}()

	// Start payment validator
	go paymentValidator(h.service.appCtx)

	// Wait for interrupt signal
	<-stop
	h.service.appCtx.logger.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := h.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	h.service.appCtx.logger.Println("Server stopped :)")
	return nil
}

// configures HTTP routes
func (h *HTTPServer) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", h.healthCheckHandler)

	// Business endpoints
	mux.HandleFunc("/banks", h.service.getAllBanksHandler)
	mux.HandleFunc("/myaccounts", h.service.getUserAccountsHandler)
	mux.HandleFunc("/new", h.service.createUserAccountHandler)

	return mux
}

// configures middleware chain
func (h *HTTPServer) setupMiddleware(handler http.Handler) http.Handler {
	return cmn.SetUserIDMiddlewareHandler(
		cmn.SetContextValuesMiddleware(
			map[cmn.ContextKey]any{cmn.AppCtx: h.service.appCtx})(handler))
}

// provides a health check endpoint
func (h *HTTPServer) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"account-service","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
}

// gracefully stop the server
func (h *HTTPServer) Stop(ctx context.Context) error {
	if h.server == nil {
		return nil
	}
	return h.server.Shutdown(ctx)
}
