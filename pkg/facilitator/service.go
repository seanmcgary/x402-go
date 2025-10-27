package facilitator

import (
	"context"
	"crypto/ecdsa"

	"github.com/seanmcgary/x402-go/pkg/chains"
	"github.com/seanmcgary/x402-go/pkg/schemes/exact"
	x402types "github.com/seanmcgary/x402-go/pkg/types"
)

// Service provides the core business logic for the facilitator
type Service struct {
	verifiers map[string]exact.Verifier // Network ID -> Verifier
	settlers  map[string]exact.Settler  // Network ID -> Settler
	registry  *chains.Registry
}

// NewService creates a new facilitator service
func NewService(
	verifiers map[string]exact.Verifier,
	settlers map[string]exact.Settler,
	registry *chains.Registry,
) *Service {
	return &Service{
		verifiers: verifiers,
		settlers:  settlers,
		registry:  registry,
	}
}

// Verify verifies a payment authorization without executing it
func (s *Service) Verify(ctx context.Context, req x402types.VerifyRequest) (*x402types.VerifyResponse, error) {
	// Validate scheme
	if req.PaymentPayload.Scheme != "exact" {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorUnsupportedScheme,
			Payer:         req.PaymentPayload.Payload.Authorization.From,
		}, nil
	}

	// Get verifier for network
	verifier, ok := s.verifiers[req.PaymentPayload.Network]
	if !ok {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidNetwork,
			Payer:         req.PaymentPayload.Payload.Authorization.From,
		}, nil
	}

	// Verify the payment
	return verifier.Verify(ctx, req)
}

// Settle executes a verified payment on the blockchain
func (s *Service) Settle(ctx context.Context, req x402types.SettleRequest) (*x402types.SettlementResponse, error) {
	// Validate scheme
	if req.PaymentPayload.Scheme != "exact" {
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorUnsupportedScheme,
			Transaction: "",
			Network:     req.PaymentPayload.Network,
			Payer:       req.PaymentPayload.Payload.Authorization.From,
		}, nil
	}

	// Get settler for network
	settler, ok := s.settlers[req.PaymentPayload.Network]
	if !ok {
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorInvalidNetwork,
			Transaction: "",
			Network:     req.PaymentPayload.Network,
			Payer:       req.PaymentPayload.Payload.Authorization.From,
		}, nil
	}

	// Settle the payment
	return settler.Settle(ctx, req)
}

// GetSupported returns the list of supported payment schemes and networks
func (s *Service) GetSupported() *x402types.SupportedResponse {
	kinds := []x402types.SupportedKind{}

	// For each registered chain, add an exact scheme entry
	for _, chain := range s.registry.GetAll() {
		kinds = append(kinds, x402types.SupportedKind{
			X402Version: 1,
			Scheme:      "exact",
			Network:     chain.NetworkID(),
		})
	}

	return &x402types.SupportedResponse{
		Kinds: kinds,
	}
}

// ServiceBuilder helps construct a Service with proper dependencies
type ServiceBuilder struct {
	registry     *chains.Registry
	executorKeys map[string]*ecdsa.PrivateKey // Network ID -> Executor key
}

// NewServiceBuilder creates a new ServiceBuilder
func NewServiceBuilder(registry *chains.Registry) *ServiceBuilder {
	return &ServiceBuilder{
		registry:     registry,
		executorKeys: make(map[string]*ecdsa.PrivateKey),
	}
}

// WithExecutorKey adds an executor key for a specific network
func (b *ServiceBuilder) WithExecutorKey(networkID string, key *ecdsa.PrivateKey) *ServiceBuilder {
	b.executorKeys[networkID] = key
	return b
}

// Build constructs the Service with all dependencies
func (b *ServiceBuilder) Build() (*Service, error) {
	// Create verifiers and settlers for each chain in registry
	verifiers := make(map[string]exact.Verifier)
	settlers := make(map[string]exact.Settler)

	// Note: This is a simplified builder. In a real implementation, you would
	// create blockchain clients here. For now, this is a placeholder.
	// The actual client creation will be done in the main.go

	return NewService(verifiers, settlers, b.registry), nil
}
