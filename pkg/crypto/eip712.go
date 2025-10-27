package crypto

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// EIP712Domain represents the EIP-712 domain separator fields
type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           *big.Int
	VerifyingContract common.Address
}

// TransferWithAuthorizationParams contains the parameters for EIP-3009 transferWithAuthorization
type TransferWithAuthorizationParams struct {
	From        common.Address
	To          common.Address
	Value       *big.Int
	ValidAfter  *big.Int
	ValidBefore *big.Int
	Nonce       [32]byte
}

// EIP712DomainSeparator computes the EIP-712 domain separator
func EIP712DomainSeparator(domain EIP712Domain) (common.Hash, error) {
	// EIP-712 domain type hash: keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")
	domainTypeHash := crypto.Keccak256Hash([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))

	// Hash the domain fields
	nameHash := crypto.Keccak256Hash([]byte(domain.Name))
	versionHash := crypto.Keccak256Hash([]byte(domain.Version))

	// Encode and hash: domainTypeHash || nameHash || versionHash || chainId || verifyingContract
	encoded := append(domainTypeHash.Bytes(), nameHash.Bytes()...)
	encoded = append(encoded, versionHash.Bytes()...)
	encoded = append(encoded, common.LeftPadBytes(domain.ChainID.Bytes(), 32)...)
	encoded = append(encoded, common.LeftPadBytes(domain.VerifyingContract.Bytes(), 32)...)

	return crypto.Keccak256Hash(encoded), nil
}

// TransferWithAuthorizationTypeHash returns the EIP-712 type hash for TransferWithAuthorization
func TransferWithAuthorizationTypeHash() common.Hash {
	// keccak256("TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)")
	return crypto.Keccak256Hash([]byte("TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)"))
}

// HashTransferWithAuthorization computes the EIP-712 hash for TransferWithAuthorization
func HashTransferWithAuthorization(params TransferWithAuthorizationParams) common.Hash {
	typeHash := TransferWithAuthorizationTypeHash()

	// Encode: typeHash || from || to || value || validAfter || validBefore || nonce
	encoded := append(typeHash.Bytes(), common.LeftPadBytes(params.From.Bytes(), 32)...)
	encoded = append(encoded, common.LeftPadBytes(params.To.Bytes(), 32)...)
	encoded = append(encoded, common.LeftPadBytes(params.Value.Bytes(), 32)...)
	encoded = append(encoded, common.LeftPadBytes(params.ValidAfter.Bytes(), 32)...)
	encoded = append(encoded, common.LeftPadBytes(params.ValidBefore.Bytes(), 32)...)
	encoded = append(encoded, params.Nonce[:]...)

	return crypto.Keccak256Hash(encoded)
}

// EIP712Hash computes the final EIP-712 hash for signing
func EIP712Hash(domainSeparator common.Hash, structHash common.Hash) common.Hash {
	// EIP-712 encoding: "\x19\x01" || domainSeparator || structHash
	encoded := append([]byte{0x19, 0x01}, domainSeparator.Bytes()...)
	encoded = append(encoded, structHash.Bytes()...)
	return crypto.Keccak256Hash(encoded)
}

// VerifySignature verifies an EIP-712 signature and returns the recovered address
func VerifySignature(domain EIP712Domain, params TransferWithAuthorizationParams, signature []byte) (common.Address, error) {
	if len(signature) != 65 {
		return common.Address{}, fmt.Errorf("invalid signature length: expected 65, got %d", len(signature))
	}

	// Compute domain separator
	domainSeparator, err := EIP712DomainSeparator(domain)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to compute domain separator: %w", err)
	}

	// Hash the struct
	structHash := HashTransferWithAuthorization(params)

	// Compute final EIP-712 hash
	hash := EIP712Hash(domainSeparator, structHash)

	fmt.Printf("EIP-712 Recovery Debug:\n")
	fmt.Printf("  Domain separator: %s\n", domainSeparator.Hex())
	fmt.Printf("  Struct hash: %s\n", structHash.Hex())
	fmt.Printf("  Final hash: %s\n", hash.Hex())
	fmt.Printf("  Signature v (raw): %d\n", signature[64])

	// Adjust v value if needed (some implementations use 27/28, we need 0/1)
	sig := make([]byte, 65)
	copy(sig, signature)
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	fmt.Printf("  Signature v (adjusted): %d\n", sig[64])

	// Recover public key from signature
	pubKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get address from public key
	address := crypto.PubkeyToAddress(*pubKey)
	fmt.Printf("  Recovered address: %s\n", address.Hex())
	return address, nil
}

// RecoverSigner recovers the signer address from an EIP-712 signature
func RecoverSigner(domain EIP712Domain, params TransferWithAuthorizationParams, signature []byte) (common.Address, error) {
	return VerifySignature(domain, params, signature)
}

// BuildTypedData constructs the EIP-712 typed data structure for use with signing libraries
func BuildTypedData(domain EIP712Domain, params TransferWithAuthorizationParams) apitypes.TypedData {
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"TransferWithAuthorization": []apitypes.Type{
				{Name: "from", Type: "address"},
				{Name: "to", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "validAfter", Type: "uint256"},
				{Name: "validBefore", Type: "uint256"},
				{Name: "nonce", Type: "bytes32"},
			},
		},
		PrimaryType: "TransferWithAuthorization",
		Domain: apitypes.TypedDataDomain{
			Name:              domain.Name,
			Version:           domain.Version,
			ChainId:           (*math.HexOrDecimal256)(domain.ChainID),
			VerifyingContract: domain.VerifyingContract.Hex(),
		},
		Message: apitypes.TypedDataMessage{
			"from":        params.From.Hex(),
			"to":          params.To.Hex(),
			"value":       params.Value.String(),
			"validAfter":  params.ValidAfter.String(),
			"validBefore": params.ValidBefore.String(),
			"nonce":       common.BytesToHash(params.Nonce[:]).Hex(),
		},
	}
}
