package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EVMClient defines the interface for interacting with EVM-compatible blockchains
type EVMClient interface {
	// BalanceAt returns the balance of an account at a specific block number
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)

	// CallContract executes a message call transaction (read-only, no state change)
	CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)

	// SendTransaction injects a signed transaction into the pending pool for execution
	SendTransaction(ctx context.Context, tx *types.Transaction) error

	// TransactionReceipt returns the receipt of a transaction by transaction hash
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)

	// ChainID retrieves the current chain ID for transaction replay protection
	ChainID(ctx context.Context) (*big.Int, error)

	// BlockNumber returns the most recent block number
	BlockNumber(ctx context.Context) (uint64, error)

	// PendingNonceAt returns the account nonce of the given account in the pending state
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)

	// EstimateGas tries to estimate the gas needed to execute a specific transaction
	EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error)

	// SuggestGasPrice retrieves the currently suggested gas price
	SuggestGasPrice(ctx context.Context) (*big.Int, error)

	// Close closes the client connection
	Close()
}

// Client is the concrete implementation of EVMClient using go-ethereum's ethclient
type Client struct {
	client *ethclient.Client
	rpcURL string
}

// NewClient creates a new EVM client connected to the specified RPC endpoint
func NewClient(rpcURL string) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC endpoint %s: %w", rpcURL, err)
	}

	return &Client{
		client: client,
		rpcURL: rpcURL,
	}, nil
}

// BalanceAt returns the balance of an account at a specific block number
func (c *Client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	balance, err := c.client.BalanceAt(ctx, account, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance for account %s: %w", account.Hex(), err)
	}
	return balance, nil
}

// CallContract executes a message call transaction (read-only, no state change)
func (c *Client) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	result, err := c.client.CallContract(ctx, call, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("contract call failed: %w", err)
	}
	return result, nil
}

// SendTransaction injects a signed transaction into the pending pool for execution
func (c *Client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if err := c.client.SendTransaction(ctx, tx); err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}
	return nil
}

// TransactionReceipt returns the receipt of a transaction by transaction hash
func (c *Client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	receipt, err := c.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt for %s: %w", txHash.Hex(), err)
	}
	return receipt, nil
}

// ChainID retrieves the current chain ID for transaction replay protection
func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}
	return chainID, nil
}

// BlockNumber returns the most recent block number
func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := c.client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get block number: %w", err)
	}
	return blockNumber, nil
}

// PendingNonceAt returns the account nonce of the given account in the pending state
func (c *Client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	nonce, err := c.client.PendingNonceAt(ctx, account)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending nonce for account %s: %w", account.Hex(), err)
	}
	return nonce, nil
}

// EstimateGas tries to estimate the gas needed to execute a specific transaction
func (c *Client) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	gas, err := c.client.EstimateGas(ctx, call)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}
	return gas, nil
}

// SuggestGasPrice retrieves the currently suggested gas price
func (c *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggested gas price: %w", err)
	}
	return gasPrice, nil
}

// Close closes the client connection
func (c *Client) Close() {
	c.client.Close()
}

// WaitForReceipt waits for a transaction receipt with timeout and polling
func (c *Client) WaitForReceipt(ctx context.Context, txHash common.Hash, pollInterval time.Duration) (*types.Receipt, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while waiting for receipt: %w", ctx.Err())
		case <-ticker.C:
			receipt, err := c.TransactionReceipt(ctx, txHash)
			if err == nil {
				return receipt, nil
			}
			// If error is "not found", continue polling
			// Otherwise, return the error
			if err.Error() != "not found" && err.Error() != ethereum.NotFound.Error() {
				return nil, err
			}
		}
	}
}
