package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestEIP712DomainSeparator(t *testing.T) {
	domain := EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(84532),
		VerifyingContract: common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
	}

	separator, err := EIP712DomainSeparator(domain)
	if err != nil {
		t.Fatalf("Failed to compute domain separator: %v", err)
	}

	if separator == (common.Hash{}) {
		t.Error("Expected non-zero domain separator")
	}

	// Compute again to ensure determinism
	separator2, err := EIP712DomainSeparator(domain)
	if err != nil {
		t.Fatalf("Failed to compute domain separator: %v", err)
	}

	if separator != separator2 {
		t.Error("Domain separator should be deterministic")
	}
}

func TestTransferWithAuthorizationTypeHash(t *testing.T) {
	typeHash := TransferWithAuthorizationTypeHash()

	if typeHash == (common.Hash{}) {
		t.Error("Expected non-zero type hash")
	}

	// The type hash should be constant
	expectedTypeHash := crypto.Keccak256Hash([]byte("TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)"))

	if typeHash != expectedTypeHash {
		t.Errorf("Type hash mismatch: expected %s, got %s", expectedTypeHash.Hex(), typeHash.Hex())
	}
}

func TestHashTransferWithAuthorization(t *testing.T) {
	params := TransferWithAuthorizationParams{
		From:        common.HexToAddress("0x857b06519E91e3A54538791bDbb0E22373e36b66"),
		To:          common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C"),
		Value:       big.NewInt(10000),
		ValidAfter:  big.NewInt(1740672089),
		ValidBefore: big.NewInt(1740672154),
		Nonce:       [32]byte{0xf3, 0x74, 0x66, 0x13, 0xc2, 0xd9, 0x20, 0xb5, 0xfd, 0xab, 0xc0, 0x85, 0x6f, 0x2a, 0xeb, 0x2d, 0x4f, 0x88, 0xee, 0x60, 0x37, 0xb8, 0xcc, 0x5d, 0x04, 0xa7, 0x1a, 0x44, 0x62, 0xf1, 0x34, 0x80},
	}

	hash := HashTransferWithAuthorization(params)

	if hash == (common.Hash{}) {
		t.Error("Expected non-zero hash")
	}

	// Hash should be deterministic
	hash2 := HashTransferWithAuthorization(params)
	if hash != hash2 {
		t.Error("Hash should be deterministic")
	}
}

func TestEIP712Hash(t *testing.T) {
	domainSeparator := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	structHash := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

	hash := EIP712Hash(domainSeparator, structHash)

	if hash == (common.Hash{}) {
		t.Error("Expected non-zero hash")
	}

	// Verify it starts with 0x1901 prefix in the encoded form
	encoded := append([]byte{0x19, 0x01}, domainSeparator.Bytes()...)
	encoded = append(encoded, structHash.Bytes()...)
	expectedHash := crypto.Keccak256Hash(encoded)

	if hash != expectedHash {
		t.Errorf("EIP712Hash mismatch: expected %s, got %s", expectedHash.Hex(), hash.Hex())
	}
}

func TestVerifySignatureWithValidSignature(t *testing.T) {
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
		Nonce:       [32]byte{0xf3, 0x74, 0x66, 0x13, 0xc2, 0xd9, 0x20, 0xb5, 0xfd, 0xab, 0xc0, 0x85, 0x6f, 0x2a, 0xeb, 0x2d, 0x4f, 0x88, 0xee, 0x60, 0x37, 0xb8, 0xcc, 0x5d, 0x04, 0xa7, 0x1a, 0x44, 0x62, 0xf1, 0x34, 0x80},
	}

	// Sign the data
	signature, err := signEIP712(privateKey, domain, params)
	if err != nil {
		t.Fatalf("Failed to sign: %v", err)
	}

	// Verify signature
	recoveredAddr, err := VerifySignature(domain, params, signature)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}

	if recoveredAddr != address {
		t.Errorf("Address mismatch: expected %s, got %s", address.Hex(), recoveredAddr.Hex())
	}
}

func TestVerifySignatureWithInvalidSignature(t *testing.T) {
	domain := EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(84532),
		VerifyingContract: common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
	}

	params := TransferWithAuthorizationParams{
		From:        common.HexToAddress("0x857b06519E91e3A54538791bDbb0E22373e36b66"),
		To:          common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C"),
		Value:       big.NewInt(10000),
		ValidAfter:  big.NewInt(1740672089),
		ValidBefore: big.NewInt(1740672154),
		Nonce:       [32]byte{0xf3, 0x74, 0x66, 0x13, 0xc2, 0xd9, 0x20, 0xb5, 0xfd, 0xab, 0xc0, 0x85, 0x6f, 0x2a, 0xeb, 0x2d, 0x4f, 0x88, 0xee, 0x60, 0x37, 0xb8, 0xcc, 0x5d, 0x04, 0xa7, 0x1a, 0x44, 0x62, 0xf1, 0x34, 0x80},
	}

	// Create an invalid signature (all zeros)
	invalidSignature := make([]byte, 65)

	_, err := VerifySignature(domain, params, invalidSignature)
	if err == nil {
		t.Error("Expected error for invalid signature")
	}
}

func TestVerifySignatureWithWrongLength(t *testing.T) {
	domain := EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(84532),
		VerifyingContract: common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
	}

	params := TransferWithAuthorizationParams{
		From:        common.HexToAddress("0x857b06519E91e3A54538791bDbb0E22373e36b66"),
		To:          common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C"),
		Value:       big.NewInt(10000),
		ValidAfter:  big.NewInt(1740672089),
		ValidBefore: big.NewInt(1740672154),
		Nonce:       [32]byte{},
	}

	// Invalid signature length
	invalidSignature := make([]byte, 32)

	_, err := VerifySignature(domain, params, invalidSignature)
	if err == nil {
		t.Error("Expected error for invalid signature length")
	}
}

func TestRecoverSigner(t *testing.T) {
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
		Nonce:       [32]byte{0xf3, 0x74, 0x66, 0x13, 0xc2, 0xd9, 0x20, 0xb5},
	}

	// Sign the data
	signature, err := signEIP712(privateKey, domain, params)
	if err != nil {
		t.Fatalf("Failed to sign: %v", err)
	}

	// Recover signer
	recoveredAddr, err := RecoverSigner(domain, params, signature)
	if err != nil {
		t.Fatalf("Failed to recover signer: %v", err)
	}

	if recoveredAddr != address {
		t.Errorf("Signer mismatch: expected %s, got %s", address.Hex(), recoveredAddr.Hex())
	}
}

func TestBuildTypedData(t *testing.T) {
	domain := EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(84532),
		VerifyingContract: common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
	}

	params := TransferWithAuthorizationParams{
		From:        common.HexToAddress("0x857b06519E91e3A54538791bDbb0E22373e36b66"),
		To:          common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C"),
		Value:       big.NewInt(10000),
		ValidAfter:  big.NewInt(1740672089),
		ValidBefore: big.NewInt(1740672154),
		Nonce:       [32]byte{0xf3, 0x74, 0x66, 0x13, 0xc2, 0xd9, 0x20, 0xb5},
	}

	typedData := BuildTypedData(domain, params)

	if typedData.PrimaryType != "TransferWithAuthorization" {
		t.Errorf("Expected primary type TransferWithAuthorization, got %s", typedData.PrimaryType)
	}

	if typedData.Domain.Name != domain.Name {
		t.Errorf("Expected domain name %s, got %s", domain.Name, typedData.Domain.Name)
	}

	if typedData.Domain.Version != domain.Version {
		t.Errorf("Expected domain version %s, got %s", domain.Version, typedData.Domain.Version)
	}
}

// Helper function to sign EIP-712 data
func signEIP712(privateKey *ecdsa.PrivateKey, domain EIP712Domain, params TransferWithAuthorizationParams) ([]byte, error) {
	// Compute domain separator
	domainSeparator, err := EIP712DomainSeparator(domain)
	if err != nil {
		return nil, err
	}

	// Hash the struct
	structHash := HashTransferWithAuthorization(params)

	// Compute final EIP-712 hash
	hash := EIP712Hash(domainSeparator, structHash)

	// Sign
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func TestSignatureRoundTrip(t *testing.T) {
	// Test with multiple different signers
	for i := 0; i < 5; i++ {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}

		address := crypto.PubkeyToAddress(privateKey.PublicKey)

		domain := EIP712Domain{
			Name:              "Test Token",
			Version:           "1",
			ChainID:           big.NewInt(int64(1 + i)),
			VerifyingContract: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		}

		params := TransferWithAuthorizationParams{
			From:        address,
			To:          common.HexToAddress("0x0987654321098765432109876543210987654321"),
			Value:       big.NewInt(int64(1000 * (i + 1))),
			ValidAfter:  big.NewInt(int64(1000000 + i)),
			ValidBefore: big.NewInt(int64(2000000 + i)),
			Nonce:       [32]byte{byte(i), byte(i + 1), byte(i + 2)},
		}

		// Sign
		signature, err := signEIP712(privateKey, domain, params)
		if err != nil {
			t.Fatalf("Failed to sign (iteration %d): %v", i, err)
		}

		// Verify
		recoveredAddr, err := VerifySignature(domain, params, signature)
		if err != nil {
			t.Fatalf("Failed to verify signature (iteration %d): %v", i, err)
		}

		if recoveredAddr != address {
			t.Errorf("Iteration %d: Address mismatch: expected %s, got %s", i, address.Hex(), recoveredAddr.Hex())
		}
	}
}

func TestDifferentChainsProduceDifferentSignatures(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	params := TransferWithAuthorizationParams{
		From:        address,
		To:          common.HexToAddress("0x209693Bc6afc0C5328bA36FaF03C514EF312287C"),
		Value:       big.NewInt(10000),
		ValidAfter:  big.NewInt(1740672089),
		ValidBefore: big.NewInt(1740672154),
		Nonce:       [32]byte{0xf3, 0x74, 0x66, 0x13},
	}

	// Sign for chain 1
	domain1 := EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(1),
		VerifyingContract: common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
	}
	signature1, err := signEIP712(privateKey, domain1, params)
	if err != nil {
		t.Fatalf("Failed to sign for chain 1: %v", err)
	}

	// Sign for chain 2
	domain2 := EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(2),
		VerifyingContract: common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
	}
	signature2, err := signEIP712(privateKey, domain2, params)
	if err != nil {
		t.Fatalf("Failed to sign for chain 2: %v", err)
	}

	// Signatures should be different
	if hex.EncodeToString(signature1) == hex.EncodeToString(signature2) {
		t.Error("Expected different signatures for different chains")
	}

	// Both should verify correctly with their respective domains
	addr1, err := VerifySignature(domain1, params, signature1)
	if err != nil || addr1 != address {
		t.Errorf("Failed to verify signature for chain 1")
	}

	addr2, err := VerifySignature(domain2, params, signature2)
	if err != nil || addr2 != address {
		t.Errorf("Failed to verify signature for chain 2")
	}

	// Cross-verification should recover a different address (not matching the signer)
	addr1Wrong, err := VerifySignature(domain1, params, signature2)
	if err != nil {
		t.Errorf("Unexpected error when verifying chain 2 signature with chain 1 domain: %v", err)
	}
	if addr1Wrong == address {
		t.Error("Expected different address when verifying chain 2 signature with chain 1 domain")
	}

	addr2Wrong, err := VerifySignature(domain2, params, signature1)
	if err != nil {
		t.Errorf("Unexpected error when verifying chain 1 signature with chain 2 domain: %v", err)
	}
	if addr2Wrong == address {
		t.Error("Expected different address when verifying chain 1 signature with chain 2 domain")
	}
}
