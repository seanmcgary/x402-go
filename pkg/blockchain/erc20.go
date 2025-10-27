package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// ERC20 represents an ERC-20 token contract with EIP-3009 support
type ERC20 struct {
	client          EVMClient
	contractAddress common.Address
	abi             abi.ABI
}

// NewERC20 creates a new ERC-20 contract interface
func NewERC20(client EVMClient, contractAddress common.Address) (*ERC20, error) {
	// Parse the ERC-20 + EIP-3009 ABI
	contractABI, err := abi.JSON(strings.NewReader(ERC20EIP3009ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ERC-20 ABI: %w", err)
	}

	return &ERC20{
		client:          client,
		contractAddress: contractAddress,
		abi:             contractABI,
	}, nil
}

// BalanceOf returns the token balance of an account
func (e *ERC20) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	// Pack the balanceOf call
	data, err := e.abi.Pack("balanceOf", account)
	if err != nil {
		return nil, fmt.Errorf("failed to pack balanceOf call: %w", err)
	}

	// Make the call
	result, err := e.client.CallContract(ctx, ethereum.CallMsg{
		To:   &e.contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("balanceOf call failed: %w", err)
	}

	// Unpack the result
	var balance *big.Int
	err = e.abi.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack balanceOf result: %w", err)
	}

	return balance, nil
}

// SimulateTransferWithAuthorization simulates a transferWithAuthorization call without executing it
func (e *ERC20) SimulateTransferWithAuthorization(
	ctx context.Context,
	from common.Address,
	to common.Address,
	value *big.Int,
	validAfter *big.Int,
	validBefore *big.Int,
	nonce [32]byte,
	signature []byte,
) error {
	// Pack the transferWithAuthorization call
	// The signature is passed as a single bytes parameter, not split into r,s,v
	data, err := e.abi.Pack(
		"transferWithAuthorization",
		from,
		to,
		value,
		validAfter,
		validBefore,
		nonce,
		signature, // Pass signature as bytes
	)
	if err != nil {
		return fmt.Errorf("failed to pack transferWithAuthorization call: %w", err)
	}

	// Simulate the call
	// Note: The From address doesn't matter for transferWithAuthorization
	// because the contract validates the signature, not msg.sender
	_, err = e.client.CallContract(ctx, ethereum.CallMsg{
		From: from,
		To:   &e.contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return fmt.Errorf("transferWithAuthorization simulation failed: %w", err)
	}

	return nil
}

// ExecuteTransferWithAuthorization executes a transferWithAuthorization transaction
func (e *ERC20) ExecuteTransferWithAuthorization(
	ctx context.Context,
	from common.Address,
	to common.Address,
	value *big.Int,
	validAfter *big.Int,
	validBefore *big.Int,
	nonce [32]byte,
	signature []byte,
	executorPrivateKey *ecdsa.PrivateKey,
) (*types.Transaction, error) {
	// Pack the transferWithAuthorization call
	// The signature is passed as a single bytes parameter
	data, err := e.abi.Pack(
		"transferWithAuthorization",
		from,
		to,
		value,
		validAfter,
		validBefore,
		nonce,
		signature, // Pass signature as bytes
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transferWithAuthorization call: %w", err)
	}

	// Get chain ID
	chainID, err := e.client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Get executor address
	executorAddress := crypto.PubkeyToAddress(executorPrivateKey.PublicKey)

	// Get nonce for executor
	txNonce, err := e.client.PendingNonceAt(ctx, executorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Estimate gas
	gasLimit, err := e.client.EstimateGas(ctx, ethereum.CallMsg{
		From: executorAddress,
		To:   &e.contractAddress,
		Data: data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Get gas price
	gasPrice, err := e.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(
		txNonce,
		e.contractAddress,
		big.NewInt(0),
		gasLimit,
		gasPrice,
		data,
	)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), executorPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = e.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx, nil
}

// GetTransactionStatus checks if a transaction was successful
func (e *ERC20) GetTransactionStatus(ctx context.Context, txHash common.Hash) (bool, error) {
	receipt, err := e.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return false, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Status == 1 means success, 0 means failure
	return receipt.Status == types.ReceiptStatusSuccessful, nil
}

// ERC20EIP3009ABI is the ABI for ERC-20 tokens with EIP-3009 support
const ERC20EIP3009ABI = `[
	{
		"constant": true,
		"inputs": [{"name": "account", "type": "address"}],
		"name": "balanceOf",
		"outputs": [{"name": "", "type": "uint256"}],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "from", "type": "address"},
			{"name": "to", "type": "address"},
			{"name": "value", "type": "uint256"},
			{"name": "validAfter", "type": "uint256"},
			{"name": "validBefore", "type": "uint256"},
			{"name": "nonce", "type": "bytes32"},
			{"name": "signature", "type": "bytes"}
		],
		"name": "transferWithAuthorization",
		"outputs": [],
		"type": "function"
	}
]`
