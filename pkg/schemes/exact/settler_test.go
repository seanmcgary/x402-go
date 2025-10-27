package exact

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/seanmcgary/x402-go/pkg/blockchain/mocks"
	x402types "github.com/seanmcgary/x402-go/pkg/types"
)

func TestNewSettler(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	executorKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	settler := NewSettler(mockClient, executorKey, 60*time.Second)

	if settler == nil {
		t.Fatal("Expected non-nil settler")
	}

	if settler.client != mockClient {
		t.Error("Expected settler to use provided client")
	}

	if settler.receiptTimeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", settler.receiptTimeout)
	}
}

func TestNewSettlerDefaultTimeout(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	executorKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	settler := NewSettler(mockClient, executorKey, 0)

	if settler.receiptTimeout != 60*time.Second {
		t.Errorf("Expected default timeout 60s, got %v", settler.receiptTimeout)
	}
}

func TestSettleSuccess(t *testing.T) {
	t.Skip("Settle requires complex blockchain mocking - tested in integration tests")

	// This test would require:
	// - Mocking ChainID, PendingNonceAt, EstimateGas, SuggestGasPrice
	// - Mocking SendTransaction
	// - Mocking TransactionReceipt (multiple calls for polling)
	// The complexity makes it better suited for integration testing
}

func TestSettleInvalidPayload(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	executorKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	settler := NewSettler(mockClient, executorKey, 60*time.Second)

	// Create request with invalid "from" address
	req := x402types.SettleRequest{
		PaymentPayload: x402types.PaymentPayload{
			X402Version: 1,
			Scheme:      "exact",
			Network:     "base-sepolia",
			Payload: x402types.SchemePayload{
				Signature: "0x1234",
				Authorization: x402types.Authorization{
					From:        "invalid-address",
					To:          "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
					Value:       "10000",
					ValidAfter:  "1740672089",
					ValidBefore: "1740672154",
					Nonce:       "0xf3746613c2d920b5",
				},
			},
		},
		PaymentRequirements: x402types.PaymentRequirements{
			Scheme:            "exact",
			Network:           "base-sepolia",
			MaxAmountRequired: "10000",
			Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
			Resource:          "https://api.example.com/premium-data",
			Description:       "Test resource",
			MaxTimeoutSeconds: 60,
		},
	}

	resp, err := settler.Settle(context.Background(), req)

	if err != nil {
		t.Fatalf("Settle failed with error: %v", err)
	}

	if resp.Success {
		t.Error("Expected unsuccessful settlement for invalid address")
	}

	if resp.ErrorReason != x402types.ErrorInvalidPayload {
		t.Errorf("Expected error code %s, got %s", x402types.ErrorInvalidPayload, resp.ErrorReason)
	}
}

func TestSettleTransactionFailure(t *testing.T) {
	t.Skip("Complex transaction failure testing best suited for integration tests")
}
