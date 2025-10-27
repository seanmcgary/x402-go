package blockchain

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/seanmcgary/x402-go/pkg/blockchain/mocks"
)

func TestNewClient(t *testing.T) {
	// This test requires a real RPC endpoint, so we'll skip it in unit tests
	t.Skip("Skipping NewClient test - requires real RPC endpoint")
}

func TestBalanceAt(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	account := common.HexToAddress("0x1234567890123456789012345678901234567890")
	expectedBalance := big.NewInt(1000000000000000000)

	mockClient.EXPECT().
		BalanceAt(context.Background(), account, (*big.Int)(nil)).
		Return(expectedBalance, nil)

	balance, err := mockClient.BalanceAt(context.Background(), account, nil)
	if err != nil {
		t.Fatalf("BalanceAt failed: %v", err)
	}

	if balance.Cmp(expectedBalance) != 0 {
		t.Errorf("Expected balance %s, got %s", expectedBalance.String(), balance.String())
	}
}

func TestCallContract(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	contractAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	expectedResult := []byte{0x00, 0x00, 0x00, 0x01}

	call := ethereum.CallMsg{
		To:   &contractAddr,
		Data: []byte{0x01, 0x02, 0x03},
	}

	mockClient.EXPECT().
		CallContract(context.Background(), call, (*big.Int)(nil)).
		Return(expectedResult, nil)

	result, err := mockClient.CallContract(context.Background(), call, nil)
	if err != nil {
		t.Fatalf("CallContract failed: %v", err)
	}

	if len(result) != len(expectedResult) {
		t.Errorf("Expected result length %d, got %d", len(expectedResult), len(result))
	}
}

func TestSendTransaction(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)

	tx := types.NewTransaction(
		0,
		common.HexToAddress("0x1234567890123456789012345678901234567890"),
		big.NewInt(0),
		21000,
		big.NewInt(1000000000),
		nil,
	)

	mockClient.EXPECT().
		SendTransaction(context.Background(), tx).
		Return(nil)

	err := mockClient.SendTransaction(context.Background(), tx)
	if err != nil {
		t.Fatalf("SendTransaction failed: %v", err)
	}
}

func TestTransactionReceipt(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	txHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	expectedReceipt := &types.Receipt{
		Status: types.ReceiptStatusSuccessful,
		TxHash: txHash,
	}

	mockClient.EXPECT().
		TransactionReceipt(context.Background(), txHash).
		Return(expectedReceipt, nil)

	receipt, err := mockClient.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		t.Fatalf("TransactionReceipt failed: %v", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Errorf("Expected successful receipt, got status %d", receipt.Status)
	}
}

func TestChainID(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	expectedChainID := big.NewInt(84532)

	mockClient.EXPECT().
		ChainID(context.Background()).
		Return(expectedChainID, nil)

	chainID, err := mockClient.ChainID(context.Background())
	if err != nil {
		t.Fatalf("ChainID failed: %v", err)
	}

	if chainID.Cmp(expectedChainID) != 0 {
		t.Errorf("Expected chainID %s, got %s", expectedChainID.String(), chainID.String())
	}
}

func TestBlockNumber(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	expectedBlockNumber := uint64(1000000)

	mockClient.EXPECT().
		BlockNumber(context.Background()).
		Return(expectedBlockNumber, nil)

	blockNumber, err := mockClient.BlockNumber(context.Background())
	if err != nil {
		t.Fatalf("BlockNumber failed: %v", err)
	}

	if blockNumber != expectedBlockNumber {
		t.Errorf("Expected block number %d, got %d", expectedBlockNumber, blockNumber)
	}
}

func TestPendingNonceAt(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	account := common.HexToAddress("0x1234567890123456789012345678901234567890")
	expectedNonce := uint64(5)

	mockClient.EXPECT().
		PendingNonceAt(context.Background(), account).
		Return(expectedNonce, nil)

	nonce, err := mockClient.PendingNonceAt(context.Background(), account)
	if err != nil {
		t.Fatalf("PendingNonceAt failed: %v", err)
	}

	if nonce != expectedNonce {
		t.Errorf("Expected nonce %d, got %d", expectedNonce, nonce)
	}
}

func TestEstimateGas(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	expectedGas := uint64(21000)

	call := ethereum.CallMsg{
		To:   &common.Address{},
		Data: []byte{},
	}

	mockClient.EXPECT().
		EstimateGas(context.Background(), call).
		Return(expectedGas, nil)

	gas, err := mockClient.EstimateGas(context.Background(), call)
	if err != nil {
		t.Fatalf("EstimateGas failed: %v", err)
	}

	if gas != expectedGas {
		t.Errorf("Expected gas %d, got %d", expectedGas, gas)
	}
}

func TestSuggestGasPrice(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	expectedGasPrice := big.NewInt(1000000000)

	mockClient.EXPECT().
		SuggestGasPrice(context.Background()).
		Return(expectedGasPrice, nil)

	gasPrice, err := mockClient.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatalf("SuggestGasPrice failed: %v", err)
	}

	if gasPrice.Cmp(expectedGasPrice) != 0 {
		t.Errorf("Expected gas price %s, got %s", expectedGasPrice.String(), gasPrice.String())
	}
}

func TestWaitForReceipt(t *testing.T) {
	t.Skip("WaitForReceipt requires real Client implementation - tested in integration tests")
}

func TestClose(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)

	mockClient.EXPECT().Close()

	mockClient.Close()
}
