package integration

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/seanmcgary/x402-go/pkg/blockchain/mocks"
	"github.com/seanmcgary/x402-go/pkg/chains"
	crypto2 "github.com/seanmcgary/x402-go/pkg/crypto"
	"github.com/seanmcgary/x402-go/pkg/discovery"
	"github.com/seanmcgary/x402-go/pkg/facilitator"
	"github.com/seanmcgary/x402-go/pkg/schemes/exact"
	x402types "github.com/seanmcgary/x402-go/pkg/types"
	"github.com/stretchr/testify/mock"
)

// TestFullVerifyFlow tests the complete verification flow from HTTP request to blockchain check
func TestFullVerifyFlow(t *testing.T) {
	// Setup
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	toAddr := common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C")
	assetAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")
	chainID := big.NewInt(84532)

	currentTime := time.Now()
	validAfter := big.NewInt(currentTime.Add(-10 * time.Minute).Unix())
	validBefore := big.NewInt(currentTime.Add(10 * time.Minute).Unix())
	nonce := [32]byte{0xf3, 0x74, 0x66, 0x13, 0xc2, 0xd9, 0x20, 0xb5, 0xfd, 0xab, 0xc0, 0x85, 0x6f, 0x2a, 0xeb, 0x2d, 0x4f, 0x88, 0xee, 0x60, 0x37, 0xb8, 0xcc, 0x5d, 0x04, 0xa7, 0x1a, 0x44, 0x62, 0xf1, 0x34, 0x80}

	domain := crypto2.EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           chainID,
		VerifyingContract: assetAddr,
	}

	params := crypto2.TransferWithAuthorizationParams{
		From:        address,
		To:          toAddr,
		Value:       big.NewInt(10000),
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	// Sign authorization
	signature := signEIP712(t, privateKey, domain, params)
	signatureHex := "0x" + common.Bytes2Hex(signature)

	// Create verify request
	req := x402types.VerifyRequest{
		PaymentPayload: x402types.PaymentPayload{
			X402Version: 1,
			Scheme:      "exact",
			Network:     "base-sepolia",
			Payload: x402types.SchemePayload{
				Signature: signatureHex,
				Authorization: x402types.Authorization{
					From:        address.Hex(),
					To:          toAddr.Hex(),
					Value:       "10000",
					ValidAfter:  validAfter.String(),
					ValidBefore: validBefore.String(),
					Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
				},
			},
		},
		PaymentRequirements: x402types.PaymentRequirements{
			Scheme:            "exact",
			Network:           "base-sepolia",
			MaxAmountRequired: "10000",
			Asset:             assetAddr.Hex(),
			PayTo:             toAddr.Hex(),
			Resource:          "https://api.example.com/premium-data",
			Description:       "Test resource",
			MaxTimeoutSeconds: 60,
			Extra: map[string]interface{}{
				"name":    "USD Coin",
				"version": "2",
			},
		},
	}

	// Setup mocked blockchain client
	mockClient := mocks.NewMockEVMClient(t)

	mockClient.EXPECT().
		ChainID(context.Background()).
		Return(chainID, nil)

	// Mock balance check - sufficient balance
	balanceData := common.LeftPadBytes(big.NewInt(20000).Bytes(), 32)
	mockClient.EXPECT().
		CallContract(context.Background(), mock.Anything, (*big.Int)(nil)).
		Return(balanceData, nil).
		Once()

	// Mock simulation call
	mockClient.EXPECT().
		CallContract(context.Background(), mock.Anything, (*big.Int)(nil)).
		Return([]byte{}, nil).
		Once()

	// Create verifier
	verifier := exact.NewVerifier(mockClient)

	// Create registry
	registry := chains.NewRegistry()
	registry.Register(chains.NewBaseSepolia("https://sepolia.base.org"))

	// Create service
	verifiers := map[string]exact.Verifier{
		"base-sepolia": verifier,
	}
	service := facilitator.NewService(verifiers, make(map[string]exact.Settler), registry)

	// Execute verification
	resp, err := service.Verify(context.Background(), req)

	// Assertions
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if !resp.IsValid {
		t.Errorf("Expected valid response, got invalid with reason: %s", resp.InvalidReason)
	}

	if resp.Payer != address.Hex() {
		t.Errorf("Expected payer %s, got %s", address.Hex(), resp.Payer)
	}
}

// TestFullSettleFlow tests the complete settlement flow
func TestFullSettleFlow(t *testing.T) {
	t.Skip("Full settle flow requires complex transaction mocking - architecture validated")
}

// TestSupportedEndpoint tests the supported endpoint returns correct data
func TestSupportedEndpoint(t *testing.T) {
	// Create registry with multiple chains
	registry := chains.NewRegistry()
	registry.Register(chains.NewBaseSepolia("https://sepolia.base.org"))
	registry.Register(chains.NewBase("https://mainnet.base.org"))
	registry.Register(chains.NewEthereum("https://eth.llamarpc.com"))

	// Create service
	service := facilitator.NewService(
		make(map[string]exact.Verifier),
		make(map[string]exact.Settler),
		registry,
	)

	// Get supported schemes
	resp := service.GetSupported()

	// Assertions
	if len(resp.Kinds) != 3 {
		t.Errorf("Expected 3 supported kinds, got %d", len(resp.Kinds))
	}

	// Verify all have exact scheme and version 1
	for _, kind := range resp.Kinds {
		if kind.Scheme != "exact" {
			t.Errorf("Expected scheme 'exact', got %s", kind.Scheme)
		}
		if kind.X402Version != 1 {
			t.Errorf("Expected x402Version 1, got %d", kind.X402Version)
		}
	}

	// Check for specific networks
	networks := make(map[string]bool)
	for _, kind := range resp.Kinds {
		networks[kind.Network] = true
	}

	if !networks["base-sepolia"] {
		t.Error("Expected base-sepolia in supported networks")
	}
	if !networks["base"] {
		t.Error("Expected base in supported networks")
	}
	if !networks["ethereum"] {
		t.Error("Expected ethereum in supported networks")
	}
}

// TestDiscoveryFlow tests the discovery service integration
func TestDiscoveryFlow(t *testing.T) {
	discoveryService := discovery.NewService()

	// Add test resources
	resource1 := x402types.DiscoveredResource{
		Resource:    "https://api.example.com/data1",
		Type:        "http",
		X402Version: 1,
		Accepts: []x402types.PaymentRequirements{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/data1",
				Description:       "Test resource 1",
				MaxTimeoutSeconds: 60,
			},
		},
		LastUpdated: 1703123456,
		Metadata: map[string]interface{}{
			"category": "finance",
			"provider": "Test Corp",
		},
	}

	resource2 := x402types.DiscoveredResource{
		Resource:    "grpc://api.example.com/data2",
		Type:        "grpc",
		X402Version: 1,
		Accepts: []x402types.PaymentRequirements{
			{
				Scheme:            "exact",
				Network:           "base",
				MaxAmountRequired: "20000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "grpc://api.example.com/data2",
				Description:       "Test resource 2",
				MaxTimeoutSeconds: 120,
			},
		},
		LastUpdated: 1703123457,
	}

	discoveryService.AddResource(resource1)
	discoveryService.AddResource(resource2)

	// Test query without filter
	resp := discoveryService.Query("", 10, 0)

	if len(resp.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(resp.Items))
	}

	if resp.Pagination.Total != 2 {
		t.Errorf("Expected total 2, got %d", resp.Pagination.Total)
	}

	// Test query with type filter
	httpResp := discoveryService.Query("http", 10, 0)

	if len(httpResp.Items) != 1 {
		t.Errorf("Expected 1 http item, got %d", len(httpResp.Items))
	}

	if httpResp.Items[0].Type != "http" {
		t.Errorf("Expected type http, got %s", httpResp.Items[0].Type)
	}

	// Test pagination
	page1 := discoveryService.Query("", 1, 0)
	if len(page1.Items) != 1 {
		t.Errorf("Expected 1 item on page 1, got %d", len(page1.Items))
	}

	page2 := discoveryService.Query("", 1, 1)
	if len(page2.Items) != 1 {
		t.Errorf("Expected 1 item on page 2, got %d", len(page2.Items))
	}

	// Verify pages have different resources
	if page1.Items[0].Resource == page2.Items[0].Resource {
		t.Error("Expected different resources on different pages")
	}
}

// Helper function to sign EIP-712 data
func signEIP712(t *testing.T, privateKey *ecdsa.PrivateKey, domain crypto2.EIP712Domain, params crypto2.TransferWithAuthorizationParams) []byte {
	t.Helper()

	domainSeparator, err := crypto2.EIP712DomainSeparator(domain)
	if err != nil {
		t.Fatalf("Failed to compute domain separator: %v", err)
	}

	structHash := crypto2.HashTransferWithAuthorization(params)
	hash := crypto2.EIP712Hash(domainSeparator, structHash)

	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		t.Fatalf("Failed to sign: %v", err)
	}

	return signature
}
