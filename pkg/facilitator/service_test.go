package facilitator

import (
	"context"
	"testing"

	"github.com/seanmcgary/x402-go/pkg/chains"
	"github.com/seanmcgary/x402-go/pkg/schemes/exact"
	x402types "github.com/seanmcgary/x402-go/pkg/types"
)

func TestNewService(t *testing.T) {
	registry := chains.NewRegistry()
	verifiers := make(map[string]exact.Verifier)
	settlers := make(map[string]exact.Settler)

	service := NewService(verifiers, settlers, registry)

	if service == nil {
		t.Fatal("Expected non-nil service")
	}

	if service.registry != registry {
		t.Error("Expected service to use provided registry")
	}
}

func TestGetSupported(t *testing.T) {
	registry := chains.NewRegistry()

	// Register some chains
	registry.Register(chains.NewBaseSepolia("https://sepolia.base.org"))
	registry.Register(chains.NewBase("https://mainnet.base.org"))

	service := NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)

	resp := service.GetSupported()

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if len(resp.Kinds) != 2 {
		t.Errorf("Expected 2 supported kinds, got %d", len(resp.Kinds))
	}

	// Verify all kinds have correct scheme
	for _, kind := range resp.Kinds {
		if kind.Scheme != "exact" {
			t.Errorf("Expected scheme 'exact', got %s", kind.Scheme)
		}
		if kind.X402Version != 1 {
			t.Errorf("Expected x402Version 1, got %d", kind.X402Version)
		}
	}

	// Check for specific networks
	hasBaseSepolia := false
	hasBase := false
	for _, kind := range resp.Kinds {
		if kind.Network == chains.NetworkBaseSepolia {
			hasBaseSepolia = true
		}
		if kind.Network == chains.NetworkBase {
			hasBase = true
		}
	}

	if !hasBaseSepolia {
		t.Error("Expected base-sepolia in supported kinds")
	}
	if !hasBase {
		t.Error("Expected base in supported kinds")
	}
}

func TestVerifyUnsupportedScheme(t *testing.T) {
	registry := chains.NewRegistry()
	service := NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)

	req := x402types.VerifyRequest{
		PaymentPayload: x402types.PaymentPayload{
			Scheme: "unsupported-scheme",
			Payload: x402types.SchemePayload{
				Authorization: x402types.Authorization{
					From: "0x1234567890123456789012345678901234567890",
				},
			},
		},
	}

	resp, err := service.Verify(context.Background(), req)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if resp.IsValid {
		t.Error("Expected invalid response for unsupported scheme")
	}

	if resp.InvalidReason != x402types.ErrorUnsupportedScheme {
		t.Errorf("Expected error code %s, got %s", x402types.ErrorUnsupportedScheme, resp.InvalidReason)
	}
}

func TestVerifyInvalidNetwork(t *testing.T) {
	registry := chains.NewRegistry()
	service := NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)

	req := x402types.VerifyRequest{
		PaymentPayload: x402types.PaymentPayload{
			Scheme:  "exact",
			Network: "unsupported-network",
			Payload: x402types.SchemePayload{
				Authorization: x402types.Authorization{
					From: "0x1234567890123456789012345678901234567890",
				},
			},
		},
	}

	resp, err := service.Verify(context.Background(), req)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if resp.IsValid {
		t.Error("Expected invalid response for unsupported network")
	}

	if resp.InvalidReason != x402types.ErrorInvalidNetwork {
		t.Errorf("Expected error code %s, got %s", x402types.ErrorInvalidNetwork, resp.InvalidReason)
	}
}

func TestSettleUnsupportedScheme(t *testing.T) {
	registry := chains.NewRegistry()
	service := NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)

	req := x402types.SettleRequest{
		PaymentPayload: x402types.PaymentPayload{
			Scheme:  "unsupported-scheme",
			Network: "base-sepolia",
			Payload: x402types.SchemePayload{
				Authorization: x402types.Authorization{
					From: "0x1234567890123456789012345678901234567890",
				},
			},
		},
	}

	resp, err := service.Settle(context.Background(), req)
	if err != nil {
		t.Fatalf("Settle failed: %v", err)
	}

	if resp.Success {
		t.Error("Expected unsuccessful settlement for unsupported scheme")
	}

	if resp.ErrorReason != x402types.ErrorUnsupportedScheme {
		t.Errorf("Expected error code %s, got %s", x402types.ErrorUnsupportedScheme, resp.ErrorReason)
	}
}

func TestSettleInvalidNetwork(t *testing.T) {
	registry := chains.NewRegistry()
	service := NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)

	req := x402types.SettleRequest{
		PaymentPayload: x402types.PaymentPayload{
			Scheme:  "exact",
			Network: "unsupported-network",
			Payload: x402types.SchemePayload{
				Authorization: x402types.Authorization{
					From: "0x1234567890123456789012345678901234567890",
				},
			},
		},
	}

	resp, err := service.Settle(context.Background(), req)
	if err != nil {
		t.Fatalf("Settle failed: %v", err)
	}

	if resp.Success {
		t.Error("Expected unsuccessful settlement for unsupported network")
	}

	if resp.ErrorReason != x402types.ErrorInvalidNetwork {
		t.Errorf("Expected error code %s, got %s", x402types.ErrorInvalidNetwork, resp.ErrorReason)
	}
}

func TestNewServiceBuilder(t *testing.T) {
	registry := chains.NewRegistry()
	builder := NewServiceBuilder(registry)

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}

	if builder.registry != registry {
		t.Error("Expected builder to use provided registry")
	}
}
