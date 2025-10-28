package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/seanmcgary/x402-go/pkg/discovery"
	"github.com/seanmcgary/x402-go/pkg/facilitator"
	x402types "github.com/seanmcgary/x402-go/pkg/types"
)

// Handler manages HTTP endpoints for the facilitator
type Handler struct {
	service          *facilitator.Service
	discoveryService *discovery.Service
}

// NewHandler creates a new HTTP handler
func NewHandler(service *facilitator.Service, discoveryService *discovery.Service) *Handler {
	return &Handler{
		service:          service,
		discoveryService: discoveryService,
	}
}

// RegisterRoutes registers all HTTP routes on the provided router
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Apply middleware
	router.Use(loggingMiddleware)
	router.Use(recoveryMiddleware)
	router.Use(contentTypeMiddleware)

	// Register endpoints
	router.HandleFunc("/verify", h.handleVerify).Methods("POST")
	router.HandleFunc("/settle", h.handleSettle).Methods("POST")
	router.HandleFunc("/supported", h.handleSupported).Methods("GET")
	router.HandleFunc("/discovery/resources", h.handleDiscoveryResources).Methods("GET")
}

// handleVerify implements POST /verify endpoint (spec section 7.1)
func (h *Handler) handleVerify(w http.ResponseWriter, r *http.Request) {
	var req x402types.VerifyRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Call service
	resp, err := h.service.Verify(r.Context(), req)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "verification failed", err)
		return
	}

	// Send response
	sendJSON(w, http.StatusOK, resp)
}

// handleSettle implements POST /settle endpoint (spec section 7.2)
func (h *Handler) handleSettle(w http.ResponseWriter, r *http.Request) {
	var req x402types.SettleRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Call service
	resp, err := h.service.Settle(r.Context(), req)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "settlement failed", err)
		return
	}

	// Send response
	sendJSON(w, http.StatusOK, resp)
}

// handleSupported implements GET /supported endpoint (spec section 7.3)
func (h *Handler) handleSupported(w http.ResponseWriter, r *http.Request) {
	resp := h.service.GetSupported()
	sendJSON(w, http.StatusOK, resp)
}

// handleDiscoveryResources implements GET /discovery/resources endpoint (spec section 8.1)
func (h *Handler) handleDiscoveryResources(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	resourceType := query.Get("type")

	limit := 20
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Query the discovery service
	resp := h.discoveryService.Query(resourceType, limit, offset)

	// Send response
	sendJSON(w, http.StatusOK, resp)
}

// sendJSON sends a JSON response
func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// sendError sends an error response
func sendError(w http.ResponseWriter, statusCode int, message string, err error) {
	log.Printf("Error: %s: %v", message, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   message,
		"details": err.Error(),
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware recovers from panics
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "internal server error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// contentTypeMiddleware validates Content-Type for POST requests
func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" && contentType != "" {
				log.Printf("Warning: Content-Type is %s, expected application/json", contentType)
			}
		}
		next.ServeHTTP(w, r)
	})
}
