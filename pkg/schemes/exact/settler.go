package exact

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/seanmcgary/x402-go/pkg/blockchain"
	"github.com/seanmcgary/x402-go/pkg/crypto"
	x402types "github.com/seanmcgary/x402-go/pkg/types"
)

// Settler defines the interface for settling payment authorizations
type Settler interface {
	// Settle executes a verified payment authorization on the blockchain
	Settle(ctx context.Context, req x402types.SettleRequest) (*x402types.SettlementResponse, error)
}

// ExactSchemeSettler implements payment settlement for the exact scheme
type ExactSchemeSettler struct {
	client         blockchain.EVMClient
	executorKey    *ecdsa.PrivateKey
	receiptTimeout time.Duration
}

// NewSettler creates a new ExactSchemeSettler
func NewSettler(client blockchain.EVMClient, executorKey *ecdsa.PrivateKey, receiptTimeout time.Duration) *ExactSchemeSettler {
	if receiptTimeout == 0 {
		receiptTimeout = 60 * time.Second // Default 60 seconds
	}

	return &ExactSchemeSettler{
		client:         client,
		executorKey:    executorKey,
		receiptTimeout: receiptTimeout,
	}
}

// Settle executes a verified payment authorization on the blockchain
func (s *ExactSchemeSettler) Settle(ctx context.Context, req x402types.SettleRequest) (*x402types.SettlementResponse, error) {
	payload := req.PaymentPayload
	requirements := req.PaymentRequirements

	// Parse authorization fields
	authorization := payload.Payload.Authorization
	signature := payload.Payload.Signature

	// Parse addresses
	fromAddr, err := crypto.ParseAddress(authorization.From)
	if err != nil {
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorInvalidPayload,
			Transaction: "",
			Network:     payload.Network,
			Payer:       authorization.From,
		}, nil
	}

	toAddr, err := crypto.ParseAddress(authorization.To)
	if err != nil {
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorInvalidPayload,
			Transaction: "",
			Network:     payload.Network,
			Payer:       authorization.From,
		}, nil
	}

	assetAddr, err := crypto.ParseAddress(requirements.Asset)
	if err != nil {
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorInvalidPaymentRequirements,
			Transaction: "",
			Network:     payload.Network,
			Payer:       authorization.From,
		}, nil
	}

	// Parse amounts
	value, err := crypto.ParseBigInt(authorization.Value)
	if err != nil {
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorInvalidPayload,
			Transaction: "",
			Network:     payload.Network,
			Payer:       authorization.From,
		}, nil
	}

	// Parse time values
	validAfter, err := crypto.ParseBigInt(authorization.ValidAfter)
	if err != nil {
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorInvalidPayload,
			Transaction: "",
			Network:     payload.Network,
			Payer:       authorization.From,
		}, nil
	}

	validBefore, err := crypto.ParseBigInt(authorization.ValidBefore)
	if err != nil {
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorInvalidPayload,
			Transaction: "",
			Network:     payload.Network,
			Payer:       authorization.From,
		}, nil
	}

	// Parse nonce
	nonce, err := crypto.ValidateNonce(authorization.Nonce)
	if err != nil {
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorInvalidPayload,
			Transaction: "",
			Network:     payload.Network,
			Payer:       authorization.From,
		}, nil
	}

	// Decode signature
	signatureBytes, err := decodeSignature(signature)
	if err != nil {
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorInvalidExactEVMPayloadSignature,
			Transaction: "",
			Network:     payload.Network,
			Payer:       authorization.From,
		}, nil
	}

	// Create ERC20 contract instance
	erc20, err := blockchain.NewERC20(s.client, assetAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create ERC20 contract: %w", err)
	}

	// Execute the transfer with authorization
	// The facilitator (using executor key) posts the transaction
	// The contract will validate the user's signature and transfer tokens
	executorAddress := ethcrypto.PubkeyToAddress(s.executorKey.PublicKey)
	fmt.Printf("Settlement: Executor %s posting transaction for payer %s\n", executorAddress.Hex(), fromAddr.Hex())

	tx, err := erc20.ExecuteTransferWithAuthorization(
		ctx,
		fromAddr,
		toAddr,
		value,
		validAfter,
		validBefore,
		nonce,
		signatureBytes,
		s.executorKey,
	)
	if err != nil {
		// Handle different types of errors
		fmt.Printf("Settlement failed: %v\n", err)
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorUnexpectedSettleError,
			Transaction: "",
			Network:     payload.Network,
			Payer:       authorization.From,
		}, nil
	}

	// Wait for transaction receipt
	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.receiptTimeout)
	defer cancel()

	receipt, err := s.client.(*blockchain.Client).WaitForReceipt(ctxWithTimeout, tx.Hash(), 2*time.Second)
	if err != nil {
		// Transaction was sent but we couldn't get receipt in time
		// Still return the transaction hash
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorUnexpectedSettleError,
			Transaction: tx.Hash().Hex(),
			Network:     payload.Network,
			Payer:       authorization.From,
		}, nil
	}

	// Check transaction status
	success := receipt.Status == 1
	if !success {
		return &x402types.SettlementResponse{
			Success:     false,
			ErrorReason: x402types.ErrorInvalidTransactionState,
			Transaction: tx.Hash().Hex(),
			Network:     payload.Network,
			Payer:       authorization.From,
		}, nil
	}

	// Settlement successful
	return &x402types.SettlementResponse{
		Success:     true,
		Transaction: tx.Hash().Hex(),
		Network:     payload.Network,
		Payer:       authorization.From,
	}, nil
}
