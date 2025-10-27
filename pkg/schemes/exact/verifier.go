package exact

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/seanmcgary/x402-go/pkg/blockchain"
	"github.com/seanmcgary/x402-go/pkg/crypto"
	x402types "github.com/seanmcgary/x402-go/pkg/types"
)

// Verifier defines the interface for verifying payment authorizations
type Verifier interface {
	// Verify verifies a payment authorization without executing it on the blockchain
	Verify(ctx context.Context, req x402types.VerifyRequest) (*x402types.VerifyResponse, error)
}

// ExactSchemeVerifier implements payment verification for the exact scheme
type ExactSchemeVerifier struct {
	client blockchain.EVMClient
}

// NewVerifier creates a new ExactSchemeVerifier
func NewVerifier(client blockchain.EVMClient) *ExactSchemeVerifier {
	return &ExactSchemeVerifier{
		client: client,
	}
}

// Verify implements all verification steps from section 6.1.2 of the x402 specification:
// 1. Signature Validation
// 2. Balance Verification
// 3. Amount Validation
// 4. Time Window Check
// 5. Parameter Matching
// 6. Transaction Simulation
func (v *ExactSchemeVerifier) Verify(ctx context.Context, req x402types.VerifyRequest) (*x402types.VerifyResponse, error) {
	payload := req.PaymentPayload
	requirements := req.PaymentRequirements

	// Validate x402 version
	if payload.X402Version != 1 {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidX402Version,
			Payer:         payload.Payload.Authorization.From,
		}, nil
	}

	// Validate scheme and network match
	if payload.Scheme != requirements.Scheme {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidScheme,
			Payer:         payload.Payload.Authorization.From,
		}, nil
	}

	if payload.Network != requirements.Network {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidNetwork,
			Payer:         payload.Payload.Authorization.From,
		}, nil
	}

	// Parse authorization fields
	authorization := payload.Payload.Authorization
	signature := payload.Payload.Signature

	// Parse addresses
	fromAddr, err := crypto.ParseAddress(authorization.From)
	if err != nil {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidPayload,
			Payer:         authorization.From,
		}, nil
	}

	toAddr, err := crypto.ParseAddress(authorization.To)
	if err != nil {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidPayload,
			Payer:         authorization.From,
		}, nil
	}

	assetAddr, err := crypto.ParseAddress(requirements.Asset)
	if err != nil {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidPaymentRequirements,
			Payer:         authorization.From,
		}, nil
	}

	// Parse amounts
	value, err := crypto.ParseBigInt(authorization.Value)
	if err != nil {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidPayload,
			Payer:         authorization.From,
		}, nil
	}

	requiredAmount, err := crypto.ParseBigInt(requirements.MaxAmountRequired)
	if err != nil {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidPaymentRequirements,
			Payer:         authorization.From,
		}, nil
	}

	// Parse time values
	validAfter, err := crypto.ParseBigInt(authorization.ValidAfter)
	if err != nil {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidPayload,
			Payer:         authorization.From,
		}, nil
	}

	validBefore, err := crypto.ParseBigInt(authorization.ValidBefore)
	if err != nil {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidPayload,
			Payer:         authorization.From,
		}, nil
	}

	// Parse nonce
	nonce, err := crypto.ValidateNonce(authorization.Nonce)
	if err != nil {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidPayload,
			Payer:         authorization.From,
		}, nil
	}

	// Get chain ID for EIP-712 domain
	chainID, err := v.client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Get token name and version from extra fields for EIP-712 domain
	tokenName := "USD Coin"
	tokenVersion := "2"
	if requirements.Extra != nil {
		if name, ok := requirements.Extra["name"].(string); ok {
			tokenName = name
		}
		if version, ok := requirements.Extra["version"].(string); ok {
			tokenVersion = version
		}
	}

	// Build EIP-712 domain
	domain := crypto.EIP712Domain{
		Name:              tokenName,
		Version:           tokenVersion,
		ChainID:           chainID,
		VerifyingContract: assetAddr,
	}

	// Step 1: Signature Validation
	params := crypto.TransferWithAuthorizationParams{
		From:        fromAddr,
		To:          toAddr,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	err = crypto.ValidateSignature(domain, params, signature, fromAddr)
	if err != nil {
		if validationErr, ok := err.(*crypto.ValidationError); ok {
			return &x402types.VerifyResponse{
				IsValid:       false,
				InvalidReason: validationErr.Code,
				Payer:         authorization.From,
			}, nil
		}
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidExactEVMPayloadSignature,
			Payer:         authorization.From,
		}, nil
	}

	// Step 2: Balance Verification
	erc20, err := blockchain.NewERC20(v.client, assetAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create ERC20 contract: %w", err)
	}

	balance, err := erc20.BalanceOf(ctx, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	if balance.Cmp(value) < 0 {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInsufficientFunds,
			Payer:         authorization.From,
		}, nil
	}

	// Step 3: Amount Validation
	err = crypto.ValidateAmount(value, requiredAmount)
	if err != nil {
		if validationErr, ok := err.(*crypto.ValidationError); ok {
			return &x402types.VerifyResponse{
				IsValid:       false,
				InvalidReason: validationErr.Code,
				Payer:         authorization.From,
			}, nil
		}
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidExactEVMPayloadAuthorizationValue,
			Payer:         authorization.From,
		}, nil
	}

	// Step 4: Time Window Check
	currentTime := time.Now()
	err = crypto.ValidateTimeWindow(validAfter, validBefore, currentTime)
	if err != nil {
		if validationErr, ok := err.(*crypto.ValidationError); ok {
			return &x402types.VerifyResponse{
				IsValid:       false,
				InvalidReason: validationErr.Code,
				Payer:         authorization.From,
			}, nil
		}
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorUnexpectedVerifyError,
			Payer:         authorization.From,
		}, nil
	}

	// Step 5: Parameter Matching - Recipient validation
	expectedTo, err := crypto.ParseAddress(requirements.PayTo)
	if err != nil {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidPaymentRequirements,
			Payer:         authorization.From,
		}, nil
	}

	err = crypto.ValidateRecipient(toAddr, expectedTo)
	if err != nil {
		if validationErr, ok := err.(*crypto.ValidationError); ok {
			return &x402types.VerifyResponse{
				IsValid:       false,
				InvalidReason: validationErr.Code,
				Payer:         authorization.From,
			}, nil
		}
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidExactEVMPayloadRecipientMismatch,
			Payer:         authorization.From,
		}, nil
	}

	// Step 6: Transaction Simulation
	// Decode signature for simulation
	signatureBytes, err := decodeSignature(signature)
	if err != nil {
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorInvalidExactEVMPayloadSignature,
			Payer:         authorization.From,
		}, nil
	}

	err = erc20.SimulateTransferWithAuthorization(
		ctx,
		fromAddr,
		toAddr,
		value,
		validAfter,
		validBefore,
		nonce,
		signatureBytes,
	)
	if err != nil {
		// Check if it's a balance issue or other contract error
		// For now, return a generic verification error
		return &x402types.VerifyResponse{
			IsValid:       false,
			InvalidReason: x402types.ErrorUnexpectedVerifyError,
			Payer:         authorization.From,
		}, nil
	}

	// All verifications passed
	return &x402types.VerifyResponse{
		IsValid: true,
		Payer:   authorization.From,
	}, nil
}

// decodeSignature converts a hex signature string to bytes
func decodeSignature(signatureHex string) ([]byte, error) {
	return common.FromHex(signatureHex), nil
}
