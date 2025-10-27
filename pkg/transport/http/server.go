package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/seanmcgary/x402-go/pkg/discovery"
	"github.com/seanmcgary/x402-go/pkg/facilitator"
)

// Server wraps the HTTP server
type Server struct {
	httpServer *http.Server
	handler    *Handler
}

// NewServer creates a new HTTP server
func NewServer(host string, port int, service *facilitator.Service, discoveryService *discovery.Service) *Server {
	handler := NewHandler(service, discoveryService)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	addr := fmt.Sprintf("%s:%d", host, port)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		handler:    handler,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting HTTP server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Printf("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}
