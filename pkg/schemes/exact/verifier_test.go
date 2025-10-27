package exact

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/seanmcgary/x402-go/pkg/blockchain/mocks"
	crypto2 "github.com/seanmcgary/x402-go/pkg/crypto"
	x402types "github.com/seanmcgary/x402-go/pkg/types"
	"github.com/stretchr/testify/mock"
)

func TestNewVerifier(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	verifier := NewVerifier(mockClient)

	if verifier == nil {
		t.Fatal("Expected non-nil verifier")
	}

	if verifier.client != mockClient {
		t.Error("Expected verifier to use provided client")
	}
}

func TestVerifySuccess(t *testing.T) {
	// Generate a private key for testing
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	toAddr := common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C")
	assetAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")
	chainID := big.NewInt(84532)

	// Build authorization
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

	// Sign the authorization
	signature := signEIP712(t, privateKey, domain, params)
	signatureHex := "0x" + common.Bytes2Hex(signature)

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

	mockClient := mocks.NewMockEVMClient(t)

	// Mock ChainID call
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

	verifier := NewVerifier(mockClient)
	resp, err := verifier.Verify(context.Background(), req)

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

func TestVerifyInsufficientFunds(t *testing.T) {
	// Generate a private key for testing
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
	nonce := [32]byte{0xf3, 0x74, 0x66, 0x13, 0xc2, 0xd9, 0x20, 0xb5}

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

	signature := signEIP712(t, privateKey, domain, params)
	signatureHex := "0x" + common.Bytes2Hex(signature)

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
					Nonce:       "0xf3746613c2d920b5000000000000000000000000000000000000000000000000",
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

	mockClient := mocks.NewMockEVMClient(t)

	// Mock ChainID call
	mockClient.EXPECT().
		ChainID(context.Background()).
		Return(chainID, nil)

	// Mock balance check - insufficient balance
	balanceData := common.LeftPadBytes(big.NewInt(5000).Bytes(), 32)
	mockClient.EXPECT().
		CallContract(context.Background(), mock.Anything, (*big.Int)(nil)).
		Return(balanceData, nil)

	verifier := NewVerifier(mockClient)
	resp, err := verifier.Verify(context.Background(), req)

	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if resp.IsValid {
		t.Error("Expected invalid response for insufficient funds")
	}

	if resp.InvalidReason != x402types.ErrorInsufficientFunds {
		t.Errorf("Expected error code %s, got %s", x402types.ErrorInsufficientFunds, resp.InvalidReason)
	}
}

func TestVerifyInvalidSignature(t *testing.T) {
	// Generate TWO different private keys
	privateKey1, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}

	privateKey2, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	address1 := crypto.PubkeyToAddress(privateKey1.PublicKey)
	toAddr := common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C")
	assetAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")
	chainID := big.NewInt(84532)

	currentTime := time.Now()
	validAfter := big.NewInt(currentTime.Add(-10 * time.Minute).Unix())
	validBefore := big.NewInt(currentTime.Add(10 * time.Minute).Unix())
	nonce := [32]byte{0xf3, 0x74, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	domain := crypto2.EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           chainID,
		VerifyingContract: assetAddr,
	}

	params := crypto2.TransferWithAuthorizationParams{
		From:        address1,
		To:          toAddr,
		Value:       big.NewInt(10000),
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	// Sign with privateKey2 (wrong key)
	signature := signEIP712(t, privateKey2, domain, params)
	signatureHex := "0x" + common.Bytes2Hex(signature)

	req := x402types.VerifyRequest{
		PaymentPayload: x402types.PaymentPayload{
			X402Version: 1,
			Scheme:      "exact",
			Network:     "base-sepolia",
			Payload: x402types.SchemePayload{
				Signature: signatureHex,
				Authorization: x402types.Authorization{
					From:        address1.Hex(), // Claims to be from address1
					To:          toAddr.Hex(),
					Value:       "10000",
					ValidAfter:  validAfter.String(),
					ValidBefore: validBefore.String(),
					Nonce:       "0xf374000000000000000000000000000000000000000000000000000000000000",
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

	mockClient := mocks.NewMockEVMClient(t)

	// Mock ChainID call
	mockClient.EXPECT().
		ChainID(context.Background()).
		Return(chainID, nil)

	verifier := NewVerifier(mockClient)
	resp, err := verifier.Verify(context.Background(), req)

	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if resp.IsValid {
		t.Error("Expected invalid response for wrong signature")
	}

	if resp.InvalidReason != x402types.ErrorInvalidExactEVMPayloadSignature {
		t.Errorf("Expected error code %s, got %s", x402types.ErrorInvalidExactEVMPayloadSignature, resp.InvalidReason)
	}
}

func TestVerifyInsufficientAmount(t *testing.T) {
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
	nonce := [32]byte{0xf3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	domain := crypto2.EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           chainID,
		VerifyingContract: assetAddr,
	}

	// Sign for 5000 but require 10000
	params := crypto2.TransferWithAuthorizationParams{
		From:        address,
		To:          toAddr,
		Value:       big.NewInt(5000), // Less than required
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	signature := signEIP712(t, privateKey, domain, params)
	signatureHex := "0x" + common.Bytes2Hex(signature)

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
					Value:       "5000",
					ValidAfter:  validAfter.String(),
					ValidBefore: validBefore.String(),
					Nonce:       "0xf300000000000000000000000000000000000000000000000000000000000000",
				},
			},
		},
		PaymentRequirements: x402types.PaymentRequirements{
			Scheme:            "exact",
			Network:           "base-sepolia",
			MaxAmountRequired: "10000", // Requires more
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

	mockClient := mocks.NewMockEVMClient(t)

	// Mock ChainID call
	mockClient.EXPECT().
		ChainID(context.Background()).
		Return(chainID, nil)

	// Mock balance check - sufficient balance for the payment but not for requirement
	balanceData := common.LeftPadBytes(big.NewInt(20000).Bytes(), 32)
	mockClient.EXPECT().
		CallContract(context.Background(), mock.Anything, (*big.Int)(nil)).
		Return(balanceData, nil)

	verifier := NewVerifier(mockClient)
	resp, err := verifier.Verify(context.Background(), req)

	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if resp.IsValid {
		t.Error("Expected invalid response for insufficient amount")
	}

	if resp.InvalidReason != x402types.ErrorInvalidExactEVMPayloadAuthorizationValue {
		t.Errorf("Expected error code %s, got %s", x402types.ErrorInvalidExactEVMPayloadAuthorizationValue, resp.InvalidReason)
	}
}

func TestVerifyExpiredAuthorization(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	toAddr := common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C")
	assetAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")
	chainID := big.NewInt(84532)

	// Create an expired authorization
	validAfter := big.NewInt(time.Now().Add(-2 * time.Hour).Unix())
	validBefore := big.NewInt(time.Now().Add(-1 * time.Hour).Unix()) // Expired
	nonce := [32]byte{0xf3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

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

	signature := signEIP712(t, privateKey, domain, params)
	signatureHex := "0x" + common.Bytes2Hex(signature)

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
					Nonce:       "0xf300000000000000000000000000000000000000000000000000000000000000",
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

	mockClient := mocks.NewMockEVMClient(t)

	// Mock ChainID call
	mockClient.EXPECT().
		ChainID(context.Background()).
		Return(chainID, nil)

	// Mock balance check
	balanceData := common.LeftPadBytes(big.NewInt(20000).Bytes(), 32)
	mockClient.EXPECT().
		CallContract(context.Background(), mock.Anything, (*big.Int)(nil)).
		Return(balanceData, nil)

	verifier := NewVerifier(mockClient)
	resp, err := verifier.Verify(context.Background(), req)

	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if resp.IsValid {
		t.Error("Expected invalid response for expired authorization")
	}

	if resp.InvalidReason != x402types.ErrorInvalidExactEVMPayloadAuthorizationValidBefore {
		t.Errorf("Expected error code %s, got %s", x402types.ErrorInvalidExactEVMPayloadAuthorizationValidBefore, resp.InvalidReason)
	}
}

func TestVerifyRecipientMismatch(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	toAddr := common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C")
	wrongToAddr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	assetAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")
	chainID := big.NewInt(84532)

	currentTime := time.Now()
	validAfter := big.NewInt(currentTime.Add(-10 * time.Minute).Unix())
	validBefore := big.NewInt(currentTime.Add(10 * time.Minute).Unix())
	nonce := [32]byte{0xf3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

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

	signature := signEIP712(t, privateKey, domain, params)
	signatureHex := "0x" + common.Bytes2Hex(signature)

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
					Nonce:       "0xf300000000000000000000000000000000000000000000000000000000000000",
				},
			},
		},
		PaymentRequirements: x402types.PaymentRequirements{
			Scheme:            "exact",
			Network:           "base-sepolia",
			MaxAmountRequired: "10000",
			Asset:             assetAddr.Hex(),
			PayTo:             wrongToAddr.Hex(), // Wrong recipient
			Resource:          "https://api.example.com/premium-data",
			Description:       "Test resource",
			MaxTimeoutSeconds: 60,
			Extra: map[string]interface{}{
				"name":    "USD Coin",
				"version": "2",
			},
		},
	}

	mockClient := mocks.NewMockEVMClient(t)

	// Mock ChainID call
	mockClient.EXPECT().
		ChainID(context.Background()).
		Return(chainID, nil)

	// Mock balance check
	balanceData := common.LeftPadBytes(big.NewInt(20000).Bytes(), 32)
	mockClient.EXPECT().
		CallContract(context.Background(), mock.Anything, (*big.Int)(nil)).
		Return(balanceData, nil)

	verifier := NewVerifier(mockClient)
	resp, err := verifier.Verify(context.Background(), req)

	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if resp.IsValid {
		t.Error("Expected invalid response for recipient mismatch")
	}

	if resp.InvalidReason != x402types.ErrorInvalidExactEVMPayloadRecipientMismatch {
		t.Errorf("Expected error code %s, got %s", x402types.ErrorInvalidExactEVMPayloadRecipientMismatch, resp.InvalidReason)
	}
}

// Helper function to sign EIP-712 data
func signEIP712(t *testing.T, privateKey *ecdsa.PrivateKey, domain crypto2.EIP712Domain, params crypto2.TransferWithAuthorizationParams) []byte {
	t.Helper()

	// Compute domain separator
	domainSeparator, err := crypto2.EIP712DomainSeparator(domain)
	if err != nil {
		t.Fatalf("Failed to compute domain separator: %v", err)
	}

	// Hash the struct
	structHash := crypto2.HashTransferWithAuthorization(params)

	// Compute final EIP-712 hash
	hash := crypto2.EIP712Hash(domainSeparator, structHash)

	// Sign
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		t.Fatalf("Failed to sign: %v", err)
	}

	return signature
}
