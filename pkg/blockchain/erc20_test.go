package blockchain

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/seanmcgary/x402-go/pkg/blockchain/mocks"
	"github.com/stretchr/testify/mock"
)

func TestNewERC20(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	contractAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")

	erc20, err := NewERC20(mockClient, contractAddr)
	if err != nil {
		t.Fatalf("NewERC20 failed: %v", err)
	}

	if erc20.contractAddress != contractAddr {
		t.Errorf("Expected contract address %s, got %s", contractAddr.Hex(), erc20.contractAddress.Hex())
	}
}

func TestBalanceOf(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	contractAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")
	account := common.HexToAddress("0x857b06519E91e3A54538791bDbb0E22373e36b66")

	erc20, err := NewERC20(mockClient, contractAddr)
	if err != nil {
		t.Fatalf("NewERC20 failed: %v", err)
	}

	// Parse ABI to encode the expected balance
	contractABI, _ := abi.JSON(strings.NewReader(ERC20EIP3009ABI))
	expectedBalance := big.NewInt(100000)

	// Encode the balance as bytes (ABI encoding)
	encodedBalance, err := contractABI.Pack("", expectedBalance)
	if err != nil {
		// If pack with empty name fails, try manual encoding
		encodedBalance = common.LeftPadBytes(expectedBalance.Bytes(), 32)
	}

	mockClient.EXPECT().
		CallContract(context.Background(), mock.Anything, (*big.Int)(nil)).
		Return(encodedBalance, nil)

	balance, err := erc20.BalanceOf(context.Background(), account)
	if err != nil {
		t.Fatalf("BalanceOf failed: %v", err)
	}

	if balance.Cmp(expectedBalance) != 0 {
		t.Errorf("Expected balance %s, got %s", expectedBalance.String(), balance.String())
	}
}

func TestSimulateTransferWithAuthorization(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	contractAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")

	erc20, err := NewERC20(mockClient, contractAddr)
	if err != nil {
		t.Fatalf("NewERC20 failed: %v", err)
	}

	from := common.HexToAddress("0x857b06519E91e3A54538791bDbb0E22373e36b66")
	to := common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C")
	value := big.NewInt(10000)
	validAfter := big.NewInt(1740672089)
	validBefore := big.NewInt(1740672154)
	nonce := [32]byte{0xf3, 0x74, 0x66, 0x13}
	signature := make([]byte, 65)
	signature[64] = 0 // v value

	mockClient.EXPECT().
		CallContract(context.Background(), mock.Anything, (*big.Int)(nil)).
		Return([]byte{}, nil)

	err = erc20.SimulateTransferWithAuthorization(
		context.Background(),
		from,
		to,
		value,
		validAfter,
		validBefore,
		nonce,
		signature,
	)

	if err != nil {
		t.Fatalf("SimulateTransferWithAuthorization failed: %v", err)
	}
}

func TestGetTransactionStatus(t *testing.T) {
	mockClient := mocks.NewMockEVMClient(t)
	contractAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")

	erc20, err := NewERC20(mockClient, contractAddr)
	if err != nil {
		t.Fatalf("NewERC20 failed: %v", err)
	}

	txHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	// Test successful transaction
	t.Run("successful transaction", func(t *testing.T) {
		successReceipt := &types.Receipt{
			Status: types.ReceiptStatusSuccessful,
			TxHash: txHash,
		}

		mockClient.EXPECT().
			TransactionReceipt(context.Background(), txHash).
			Return(successReceipt, nil).
			Once()

		success, err := erc20.GetTransactionStatus(context.Background(), txHash)
		if err != nil {
			t.Fatalf("GetTransactionStatus failed: %v", err)
		}

		if !success {
			t.Error("Expected successful transaction status")
		}
	})

	// Test failed transaction
	t.Run("failed transaction", func(t *testing.T) {
		failedReceipt := &types.Receipt{
			Status: types.ReceiptStatusFailed,
			TxHash: txHash,
		}

		mockClient.EXPECT().
			TransactionReceipt(context.Background(), txHash).
			Return(failedReceipt, nil).
			Once()

		success, err := erc20.GetTransactionStatus(context.Background(), txHash)
		if err != nil {
			t.Fatalf("GetTransactionStatus failed: %v", err)
		}

		if success {
			t.Error("Expected failed transaction status")
		}
	})
}

func TestExecuteTransferWithAuthorization(t *testing.T) {
	// This test is complex and requires mocking many dependencies
	// We'll test the core logic in integration tests
	t.Skip("ExecuteTransferWithAuthorization requires complex mocking - tested in integration tests")
}

func TestERC20EIP3009ABI(t *testing.T) {
	// Test that the ABI is valid and can be parsed
	_, err := abi.JSON(strings.NewReader(ERC20EIP3009ABI))
	if err != nil {
		t.Fatalf("Failed to parse ERC20EIP3009ABI: %v", err)
	}
}
