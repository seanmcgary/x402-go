package crypto

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/seanmcgary/x402-go/pkg/types"
)

func TestValidateTimeWindow(t *testing.T) {
	tests := []struct {
		name        string
		validAfter  int64
		validBefore int64
		currentTime int64
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid - within window",
			validAfter:  1000,
			validBefore: 2000,
			currentTime: 1500,
			expectError: false,
		},
		{
			name:        "valid - at validAfter",
			validAfter:  1000,
			validBefore: 2000,
			currentTime: 1000,
			expectError: false,
		},
		{
			name:        "invalid - before validAfter",
			validAfter:  1000,
			validBefore: 2000,
			currentTime: 999,
			expectError: true,
			errorCode:   types.ErrorInvalidExactEVMPayloadAuthorizationValidAfter,
		},
		{
			name:        "invalid - at validBefore",
			validAfter:  1000,
			validBefore: 2000,
			currentTime: 2000,
			expectError: true,
			errorCode:   types.ErrorInvalidExactEVMPayloadAuthorizationValidBefore,
		},
		{
			name:        "invalid - after validBefore",
			validAfter:  1000,
			validBefore: 2000,
			currentTime: 2001,
			expectError: true,
			errorCode:   types.ErrorInvalidExactEVMPayloadAuthorizationValidBefore,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validAfter := big.NewInt(tt.validAfter)
			validBefore := big.NewInt(tt.validBefore)
			currentTime := time.Unix(tt.currentTime, 0)

			err := ValidateTimeWindow(validAfter, validBefore, currentTime)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				validationErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if validationErr.Code != tt.errorCode {
					t.Errorf("Expected error code %s, got %s", tt.errorCode, validationErr.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name           string
		paymentAmount  int64
		requiredAmount int64
		expectError    bool
	}{
		{
			name:           "valid - exact amount",
			paymentAmount:  10000,
			requiredAmount: 10000,
			expectError:    false,
		},
		{
			name:           "valid - more than required",
			paymentAmount:  15000,
			requiredAmount: 10000,
			expectError:    false,
		},
		{
			name:           "invalid - less than required",
			paymentAmount:  9999,
			requiredAmount: 10000,
			expectError:    true,
		},
		{
			name:           "invalid - zero payment",
			paymentAmount:  0,
			requiredAmount: 10000,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentAmount := big.NewInt(tt.paymentAmount)
			requiredAmount := big.NewInt(tt.requiredAmount)

			err := ValidateAmount(paymentAmount, requiredAmount)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				validationErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if validationErr.Code != types.ErrorInvalidExactEVMPayloadAuthorizationValue {
					t.Errorf("Expected error code %s, got %s", types.ErrorInvalidExactEVMPayloadAuthorizationValue, validationErr.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestValidateRecipient(t *testing.T) {
	tests := []struct {
		name              string
		paymentRecipient  string
		expectedRecipient string
		expectError       bool
	}{
		{
			name:              "valid - matching recipients",
			paymentRecipient:  "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
			expectedRecipient: "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
			expectError:       false,
		},
		{
			name:              "valid - case insensitive",
			paymentRecipient:  "0x209693bc6afc0c5328ba36faf03c514ef312287c",
			expectedRecipient: "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
			expectError:       false,
		},
		{
			name:              "invalid - different recipients",
			paymentRecipient:  "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
			expectedRecipient: "0x857b06519E91e3A54538791bDbb0E22373e36b66",
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentRecipient := common.HexToAddress(tt.paymentRecipient)
			expectedRecipient := common.HexToAddress(tt.expectedRecipient)

			err := ValidateRecipient(paymentRecipient, expectedRecipient)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				validationErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if validationErr.Code != types.ErrorInvalidExactEVMPayloadRecipientMismatch {
					t.Errorf("Expected error code %s, got %s", types.ErrorInvalidExactEVMPayloadRecipientMismatch, validationErr.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestValidateNonce(t *testing.T) {
	tests := []struct {
		name        string
		nonce       string
		expectError bool
	}{
		{
			name:        "valid - with 0x prefix",
			nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
			expectError: false,
		},
		{
			name:        "valid - without 0x prefix",
			nonce:       "f3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
			expectError: false,
		},
		{
			name:        "invalid - too short",
			nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f134",
			expectError: true,
		},
		{
			name:        "invalid - too long",
			nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f1348000",
			expectError: true,
		},
		{
			name:        "invalid - not hex",
			nonce:       "0xg3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
			expectError: true,
		},
		{
			name:        "invalid - empty",
			nonce:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonce, err := ValidateNonce(tt.nonce)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
					return
				}
				// Verify nonce is 32 bytes
				if len(nonce) != 32 {
					t.Errorf("Expected 32-byte nonce, got %d bytes", len(nonce))
				}
			}
		})
	}
}

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		expectError bool
	}{
		{
			name:        "valid - with 0x prefix",
			address:     "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
			expectError: false,
		},
		{
			name:        "valid - lowercase",
			address:     "0x209693bc6afc0c5328ba36faf03c514ef312287c",
			expectError: false,
		},
		{
			name:        "valid - without 0x prefix (accepted by go-ethereum)",
			address:     "209693Bc6afc0C5328bA36FaF03C514EF312287C",
			expectError: false,
		},
		{
			name:        "invalid - too short",
			address:     "0x209693",
			expectError: true,
		},
		{
			name:        "invalid - not hex",
			address:     "0xZZZ693Bc6afc0C5328bA36FaF03C514EF312287C",
			expectError: true,
		},
		{
			name:        "invalid - empty",
			address:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := ParseAddress(tt.address)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
					return
				}
				// Normalize both addresses to 0x-prefixed format for comparison
				expectedAddr := common.HexToAddress(tt.address)
				if addr != expectedAddr {
					t.Errorf("Address mismatch: expected %s, got %s", expectedAddr.Hex(), addr.Hex())
				}
			}
		})
	}
}

func TestParseBigInt(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expected    int64
		expectError bool
	}{
		{
			name:        "valid - positive integer",
			value:       "10000",
			expected:    10000,
			expectError: false,
		},
		{
			name:        "valid - zero",
			value:       "0",
			expected:    0,
			expectError: false,
		},
		{
			name:        "valid - large number",
			value:       "1000000000000000000",
			expected:    1000000000000000000,
			expectError: false,
		},
		{
			name:        "invalid - not a number",
			value:       "abc",
			expectError: true,
		},
		{
			name:        "invalid - hex format",
			value:       "0x1234",
			expectError: true,
		},
		{
			name:        "invalid - empty",
			value:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseBigInt(tt.value)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
					return
				}
				if result.Int64() != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result.Int64())
				}
			}
		})
	}
}

func TestValidateSignature(t *testing.T) {
	// Generate a private key for testing
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	domain := EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(84532),
		VerifyingContract: common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
	}

	params := TransferWithAuthorizationParams{
		From:        address,
		To:          common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C"),
		Value:       big.NewInt(10000),
		ValidAfter:  big.NewInt(1740672089),
		ValidBefore: big.NewInt(1740672154),
		Nonce:       [32]byte{0xf3, 0x74, 0x66, 0x13},
	}

	// Sign the data
	signature, err := signEIP712(privateKey, domain, params)
	if err != nil {
		t.Fatalf("Failed to sign: %v", err)
	}

	signatureHex := "0x" + common.Bytes2Hex(signature)

	tests := []struct {
		name           string
		signatureHex   string
		expectedSigner common.Address
		expectError    bool
	}{
		{
			name:           "valid - correct signature",
			signatureHex:   signatureHex,
			expectedSigner: address,
			expectError:    false,
		},
		{
			name:           "invalid - wrong expected signer",
			signatureHex:   signatureHex,
			expectedSigner: common.HexToAddress("0x0000000000000000000000000000000000000001"),
			expectError:    true,
		},
		{
			name:           "invalid - malformed signature",
			signatureHex:   "0xinvalid",
			expectedSigner: address,
			expectError:    true,
		},
		{
			name:           "invalid - empty signature",
			signatureHex:   "",
			expectedSigner: address,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSignature(domain, params, tt.signatureHex, tt.expectedSigner)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				validationErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if validationErr.Code != types.ErrorInvalidExactEVMPayloadSignature {
					t.Errorf("Expected error code %s, got %s", types.ErrorInvalidExactEVMPayloadSignature, validationErr.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestValidateAuthorization(t *testing.T) {
	// Generate a private key for testing
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	domain := EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(84532),
		VerifyingContract: common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
	}

	currentTime := time.Unix(1740672100, 0)

	params := TransferWithAuthorizationParams{
		From:        address,
		To:          common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C"),
		Value:       big.NewInt(10000),
		ValidAfter:  big.NewInt(1740672089),
		ValidBefore: big.NewInt(1740672154),
		Nonce:       [32]byte{0xf3, 0x74, 0x66, 0x13, 0xc2, 0xd9, 0x20, 0xb5, 0xfd, 0xab, 0xc0, 0x85, 0x6f, 0x2a, 0xeb, 0x2d, 0x4f, 0x88, 0xee, 0x60, 0x37, 0xb8, 0xcc, 0x5d, 0x04, 0xa7, 0x1a, 0x44, 0x62, 0xf1, 0x34, 0x80},
	}

	// Sign the data
	signature, err := signEIP712(privateKey, domain, params)
	if err != nil {
		t.Fatalf("Failed to sign: %v", err)
	}

	signatureHex := "0x" + common.Bytes2Hex(signature)

	// Valid authorization
	validAuth := types.Authorization{
		From:        address.Hex(),
		To:          params.To.Hex(),
		Value:       params.Value.String(),
		ValidAfter:  params.ValidAfter.String(),
		ValidBefore: params.ValidBefore.String(),
		Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
	}

	validReqs := types.PaymentRequirements{
		PayTo:             params.To.Hex(),
		MaxAmountRequired: "10000",
	}

	// Test valid authorization
	t.Run("valid authorization", func(t *testing.T) {
		err := ValidateAuthorization(domain, validAuth, validReqs, signatureHex, currentTime)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	// Test with insufficient amount
	t.Run("insufficient amount", func(t *testing.T) {
		insufficientReqs := validReqs
		insufficientReqs.MaxAmountRequired = "20000"

		err := ValidateAuthorization(domain, validAuth, insufficientReqs, signatureHex, currentTime)
		if err == nil {
			t.Error("Expected error for insufficient amount")
		}
	})

	// Test with wrong recipient
	t.Run("wrong recipient", func(t *testing.T) {
		wrongRecipientReqs := validReqs
		wrongRecipientReqs.PayTo = "0x0000000000000000000000000000000000000001"

		err := ValidateAuthorization(domain, validAuth, wrongRecipientReqs, signatureHex, currentTime)
		if err == nil {
			t.Error("Expected error for wrong recipient")
		}
	})

	// Test with expired authorization
	t.Run("expired authorization", func(t *testing.T) {
		expiredTime := time.Unix(1740672200, 0) // After validBefore

		err := ValidateAuthorization(domain, validAuth, validReqs, signatureHex, expiredTime)
		if err == nil {
			t.Error("Expected error for expired authorization")
		}
	})

	// Test with not yet valid authorization
	t.Run("not yet valid authorization", func(t *testing.T) {
		earlyTime := time.Unix(1740672000, 0) // Before validAfter

		err := ValidateAuthorization(domain, validAuth, validReqs, signatureHex, earlyTime)
		if err == nil {
			t.Error("Expected error for not yet valid authorization")
		}
	})
}
