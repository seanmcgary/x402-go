package crypto

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/seanmcgary/x402-go/pkg/types"
)

// ValidationError represents a validation error with an associated error code
type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ValidateTimeWindow validates that the current time is within the authorization's valid time window
func ValidateTimeWindow(validAfter, validBefore *big.Int, currentTime time.Time) error {
	currentTimestamp := big.NewInt(currentTime.Unix())

	// Check if authorization is not yet valid
	if currentTimestamp.Cmp(validAfter) < 0 {
		return &ValidationError{
			Code:    types.ErrorInvalidExactEVMPayloadAuthorizationValidAfter,
			Message: fmt.Sprintf("authorization not yet valid: current time %d is before validAfter %d", currentTimestamp.Int64(), validAfter.Int64()),
		}
	}

	// Check if authorization has expired
	if currentTimestamp.Cmp(validBefore) >= 0 {
		return &ValidationError{
			Code:    types.ErrorInvalidExactEVMPayloadAuthorizationValidBefore,
			Message: fmt.Sprintf("authorization expired: current time %d is at or after validBefore %d", currentTimestamp.Int64(), validBefore.Int64()),
		}
	}

	return nil
}

// ValidateAmount validates that the payment amount meets or exceeds the required amount
func ValidateAmount(paymentAmount, requiredAmount *big.Int) error {
	if paymentAmount.Cmp(requiredAmount) < 0 {
		return &ValidationError{
			Code:    types.ErrorInvalidExactEVMPayloadAuthorizationValue,
			Message: fmt.Sprintf("insufficient payment amount: got %s, required %s", paymentAmount.String(), requiredAmount.String()),
		}
	}
	return nil
}

// ValidateRecipient validates that the payment recipient matches the expected recipient
func ValidateRecipient(paymentRecipient, expectedRecipient common.Address) error {
	if paymentRecipient != expectedRecipient {
		return &ValidationError{
			Code:    types.ErrorInvalidExactEVMPayloadRecipientMismatch,
			Message: fmt.Sprintf("recipient mismatch: got %s, expected %s", paymentRecipient.Hex(), expectedRecipient.Hex()),
		}
	}
	return nil
}

// ValidateNonce validates that the nonce is a valid 32-byte hex string
func ValidateNonce(nonceStr string) ([32]byte, error) {
	var nonce [32]byte

	// Remove 0x prefix if present
	nonceStr = strings.TrimPrefix(nonceStr, "0x")

	// Decode hex string
	decoded, err := hex.DecodeString(nonceStr)
	if err != nil {
		return nonce, fmt.Errorf("invalid nonce hex encoding: %w", err)
	}

	// Check length
	if len(decoded) != 32 {
		return nonce, fmt.Errorf("invalid nonce length: expected 32 bytes, got %d", len(decoded))
	}

	copy(nonce[:], decoded)
	return nonce, nil
}

// ValidateSignature validates the EIP-712 signature and returns the recovered signer address
func ValidateSignature(domain EIP712Domain, params TransferWithAuthorizationParams, signatureHex string, expectedSigner common.Address) error {
	// Remove 0x prefix if present
	signatureHex = strings.TrimPrefix(signatureHex, "0x")

	// Decode signature
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return &ValidationError{
			Code:    types.ErrorInvalidExactEVMPayloadSignature,
			Message: fmt.Sprintf("invalid signature hex encoding: %v", err),
		}
	}

	// Verify signature and recover signer
	recoveredSigner, err := VerifySignature(domain, params, signature)
	if err != nil {
		return &ValidationError{
			Code:    types.ErrorInvalidExactEVMPayloadSignature,
			Message: fmt.Sprintf("signature verification failed: %v", err),
		}
	}

	// Check if recovered signer matches expected signer
	if recoveredSigner != expectedSigner {
		return &ValidationError{
			Code:    types.ErrorInvalidExactEVMPayloadSignature,
			Message: fmt.Sprintf("signer mismatch: recovered %s, expected %s", recoveredSigner.Hex(), expectedSigner.Hex()),
		}
	}

	return nil
}

// ParseAddress parses an Ethereum address string
func ParseAddress(addressStr string) (common.Address, error) {
	if !common.IsHexAddress(addressStr) {
		return common.Address{}, fmt.Errorf("invalid address format: %s", addressStr)
	}
	return common.HexToAddress(addressStr), nil
}

// ParseBigInt parses a string into a big.Int
func ParseBigInt(s string) (*big.Int, error) {
	value := new(big.Int)
	value, ok := value.SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("invalid big int format: %s", s)
	}
	return value, nil
}

// ValidateAuthorization performs comprehensive validation of an EIP-3009 authorization
func ValidateAuthorization(
	domain EIP712Domain,
	authorization types.Authorization,
	paymentRequirements types.PaymentRequirements,
	signature string,
	currentTime time.Time,
) error {
	// Parse addresses
	fromAddr, err := ParseAddress(authorization.From)
	if err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}

	toAddr, err := ParseAddress(authorization.To)
	if err != nil {
		return fmt.Errorf("invalid to address: %w", err)
	}

	expectedToAddr, err := ParseAddress(paymentRequirements.PayTo)
	if err != nil {
		return fmt.Errorf("invalid payTo address: %w", err)
	}

	// Parse amounts
	value, err := ParseBigInt(authorization.Value)
	if err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}

	requiredAmount, err := ParseBigInt(paymentRequirements.MaxAmountRequired)
	if err != nil {
		return fmt.Errorf("invalid maxAmountRequired: %w", err)
	}

	// Parse time values
	validAfter, err := ParseBigInt(authorization.ValidAfter)
	if err != nil {
		return fmt.Errorf("invalid validAfter: %w", err)
	}

	validBefore, err := ParseBigInt(authorization.ValidBefore)
	if err != nil {
		return fmt.Errorf("invalid validBefore: %w", err)
	}

	// Validate nonce format
	nonce, err := ValidateNonce(authorization.Nonce)
	if err != nil {
		return fmt.Errorf("invalid nonce: %w", err)
	}

	// Validate recipient matches
	if err := ValidateRecipient(toAddr, expectedToAddr); err != nil {
		return err
	}

	// Validate amount
	if err := ValidateAmount(value, requiredAmount); err != nil {
		return err
	}

	// Validate time window
	if err := ValidateTimeWindow(validAfter, validBefore, currentTime); err != nil {
		return err
	}

	// Build params for signature verification
	params := TransferWithAuthorizationParams{
		From:        fromAddr,
		To:          toAddr,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	// Validate signature
	if err := ValidateSignature(domain, params, signature, fromAddr); err != nil {
		return err
	}

	return nil
}
