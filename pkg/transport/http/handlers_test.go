package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/seanmcgary/x402-go/pkg/chains"
	"github.com/seanmcgary/x402-go/pkg/discovery"
	"github.com/seanmcgary/x402-go/pkg/facilitator"
	"github.com/seanmcgary/x402-go/pkg/schemes/exact"
	x402types "github.com/seanmcgary/x402-go/pkg/types"
)

func TestNewHandler(t *testing.T) {
	registry := chains.NewRegistry()
	service := facilitator.NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)
	discoveryService := discovery.NewService()

	handler := NewHandler(service, discoveryService)

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}

	if handler.service != service {
		t.Error("Expected handler to use provided service")
	}

	if handler.discoveryService != discoveryService {
		t.Error("Expected handler to use provided discovery service")
	}
}

func TestHandleSupported(t *testing.T) {
	registry := chains.NewRegistry()
	registry.Register(chains.NewBaseSepolia("https://sepolia.base.org"))
	registry.Register(chains.NewBase("https://mainnet.base.org"))

	service := facilitator.NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)
	discoveryService := discovery.NewService()

	handler := NewHandler(service, discoveryService)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create request
	req := httptest.NewRequest("GET", "/supported", nil)
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp x402types.SupportedResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.Kinds) != 2 {
		t.Errorf("Expected 2 supported kinds, got %d", len(resp.Kinds))
	}
}

func TestHandleVerifyUnsupportedScheme(t *testing.T) {
	registry := chains.NewRegistry()
	service := facilitator.NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)
	discoveryService := discovery.NewService()

	handler := NewHandler(service, discoveryService)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create request with unsupported scheme
	reqBody := x402types.VerifyRequest{
		PaymentPayload: x402types.PaymentPayload{
			Scheme: "unsupported",
			Payload: x402types.SchemePayload{
				Authorization: x402types.Authorization{
					From: "0x1234567890123456789012345678901234567890",
				},
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp x402types.VerifyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.IsValid {
		t.Error("Expected invalid response for unsupported scheme")
	}

	if resp.InvalidReason != x402types.ErrorUnsupportedScheme {
		t.Errorf("Expected error %s, got %s", x402types.ErrorUnsupportedScheme, resp.InvalidReason)
	}
}

func TestHandleVerifyInvalidJSON(t *testing.T) {
	registry := chains.NewRegistry()
	service := facilitator.NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)
	discoveryService := discovery.NewService()

	handler := NewHandler(service, discoveryService)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/verify", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleDiscoveryResources(t *testing.T) {
	registry := chains.NewRegistry()
	service := facilitator.NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)
	discoveryService := discovery.NewService()

	// Add test resources
	discoveryService.AddResource(x402types.DiscoveredResource{
		Resource:    "https://api.example.com/data1",
		Type:        "http",
		X402Version: 1,
		LastUpdated: 1703123456,
	})

	discoveryService.AddResource(x402types.DiscoveredResource{
		Resource:    "https://api.example.com/data2",
		Type:        "http",
		X402Version: 1,
		LastUpdated: 1703123457,
	})

	handler := NewHandler(service, discoveryService)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		url            string
		expectedCount  int
		expectedLimit  int
		expectedOffset int
		expectedTotal  int
	}{
		{
			name:           "no parameters",
			url:            "/discovery/resources",
			expectedCount:  2,
			expectedLimit:  20,
			expectedOffset: 0,
			expectedTotal:  2,
		},
		{
			name:           "with limit",
			url:            "/discovery/resources?limit=1",
			expectedCount:  1,
			expectedLimit:  1,
			expectedOffset: 0,
			expectedTotal:  2,
		},
		{
			name:           "with offset",
			url:            "/discovery/resources?offset=1",
			expectedCount:  1,
			expectedLimit:  20,
			expectedOffset: 1,
			expectedTotal:  2,
		},
		{
			name:           "with type filter",
			url:            "/discovery/resources?type=http",
			expectedCount:  2,
			expectedLimit:  20,
			expectedOffset: 0,
			expectedTotal:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			var resp x402types.DiscoveryResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if len(resp.Items) != tt.expectedCount {
				t.Errorf("Expected %d items, got %d", tt.expectedCount, len(resp.Items))
			}

			if resp.Pagination.Limit != tt.expectedLimit {
				t.Errorf("Expected limit %d, got %d", tt.expectedLimit, resp.Pagination.Limit)
			}

			if resp.Pagination.Offset != tt.expectedOffset {
				t.Errorf("Expected offset %d, got %d", tt.expectedOffset, resp.Pagination.Offset)
			}

			if resp.Pagination.Total != tt.expectedTotal {
				t.Errorf("Expected total %d, got %d", tt.expectedTotal, resp.Pagination.Total)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	registry := chains.NewRegistry()
	service := facilitator.NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)
	discoveryService := discovery.NewService()

	handler := NewHandler(service, discoveryService)
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test that middleware is applied
	req := httptest.NewRequest("GET", "/supported", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should have Content-Type header from sendJSON
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}
