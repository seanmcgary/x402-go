package types

// PaymentRequirements defines acceptable payment methods for a resource
// as specified in section 5.1 of the x402 specification
type PaymentRequirements struct {
	// Scheme is the payment scheme identifier (e.g., "exact")
	Scheme string `json:"scheme"`

	// Network is the blockchain network identifier (e.g., "base-sepolia", "ethereum-mainnet")
	Network string `json:"network"`

	// MaxAmountRequired is the required payment amount in atomic token units
	MaxAmountRequired string `json:"maxAmountRequired"`

	// Asset is the token contract address
	Asset string `json:"asset"`

	// PayTo is the recipient wallet address for the payment
	PayTo string `json:"payTo"`

	// Resource is the URL of the protected resource
	Resource string `json:"resource"`

	// Description is a human-readable description of the resource
	Description string `json:"description"`

	// MimeType is the MIME type of the expected response (optional)
	MimeType string `json:"mimeType,omitempty"`

	// OutputSchema is a JSON schema describing the response format (optional)
	OutputSchema interface{} `json:"outputSchema,omitempty"`

	// MaxTimeoutSeconds is the maximum time allowed for payment completion
	MaxTimeoutSeconds int `json:"maxTimeoutSeconds"`

	// Extra contains scheme-specific additional information (optional)
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// PaymentRequirementsResponse is the response when payment is required
// as specified in section 5.1 of the x402 specification
type PaymentRequirementsResponse struct {
	// X402Version is the protocol version identifier
	X402Version int `json:"x402Version"`

	// Error is a human-readable error message explaining why payment is required
	Error string `json:"error"`

	// Accepts is an array of payment requirement objects defining acceptable payment methods
	Accepts []PaymentRequirements `json:"accepts"`
}

// Authorization contains EIP-3009 authorization parameters
// as specified in section 5.2 of the x402 specification
type Authorization struct {
	// From is the payer's wallet address
	From string `json:"from"`

	// To is the recipient's wallet address
	To string `json:"to"`

	// Value is the payment amount in atomic units
	Value string `json:"value"`

	// ValidAfter is the Unix timestamp when authorization becomes valid
	ValidAfter string `json:"validAfter"`

	// ValidBefore is the Unix timestamp when authorization expires
	ValidBefore string `json:"validBefore"`

	// Nonce is a 32-byte random nonce to prevent replay attacks
	Nonce string `json:"nonce"`
}

// SchemePayload contains scheme-specific payment data
// as specified in section 5.2 of the x402 specification
type SchemePayload struct {
	// Signature is the EIP-712 signature for authorization
	Signature string `json:"signature"`

	// Authorization contains the EIP-3009 authorization parameters
	Authorization Authorization `json:"authorization"`
}

// PaymentPayload contains the client's payment authorization
// as specified in section 5.2 of the x402 specification
type PaymentPayload struct {
	// X402Version is the protocol version identifier (must be 1)
	X402Version int `json:"x402Version"`

	// Scheme is the payment scheme identifier (e.g., "exact")
	Scheme string `json:"scheme"`

	// Network is the blockchain network identifier (e.g., "base-sepolia", "ethereum-mainnet")
	Network string `json:"network"`

	// Payload contains the payment data object
	Payload SchemePayload `json:"payload"`
}

// SettlementResponse contains transaction details after payment settlement
// as specified in section 5.3 of the x402 specification
type SettlementResponse struct {
	// Success indicates whether the payment settlement was successful
	Success bool `json:"success"`

	// ErrorReason is the error reason if settlement failed (omitted if successful)
	ErrorReason string `json:"errorReason,omitempty"`

	// Transaction is the blockchain transaction hash (empty string if settlement failed)
	Transaction string `json:"transaction"`

	// Network is the blockchain network identifier
	Network string `json:"network"`

	// Payer is the address of the payer's wallet
	Payer string `json:"payer"`
}

// VerifyRequest is the request to verify a payment authorization
// as specified in section 7.1 of the x402 specification
type VerifyRequest struct {
	// PaymentPayload contains the payment authorization from the client
	PaymentPayload PaymentPayload `json:"paymentPayload"`

	// PaymentRequirements contains the original payment requirements
	PaymentRequirements PaymentRequirements `json:"paymentRequirements"`
}

// VerifyResponse is the response from verifying a payment authorization
// as specified in section 7.1 of the x402 specification
type VerifyResponse struct {
	// IsValid indicates whether the payment authorization is valid
	IsValid bool `json:"isValid"`

	// InvalidReason is the reason the payment is invalid (only present if IsValid is false)
	InvalidReason string `json:"invalidReason,omitempty"`

	// Payer is the address of the payer's wallet
	Payer string `json:"payer"`
}

// SettleRequest is the request to execute a verified payment
// as specified in section 7.2 of the x402 specification
type SettleRequest struct {
	// PaymentPayload contains the payment authorization from the client
	PaymentPayload PaymentPayload `json:"paymentPayload"`

	// PaymentRequirements contains the original payment requirements
	PaymentRequirements PaymentRequirements `json:"paymentRequirements"`
}

// SupportedKind represents a supported payment scheme and network combination
// as specified in section 7.3 of the x402 specification
type SupportedKind struct {
	// X402Version is the protocol version identifier
	X402Version int `json:"x402Version"`

	// Scheme is the payment scheme identifier (e.g., "exact")
	Scheme string `json:"scheme"`

	// Network is the blockchain network identifier
	Network string `json:"network"`
}

// SupportedResponse is the response containing supported payment schemes and networks
// as specified in section 7.3 of the x402 specification
type SupportedResponse struct {
	// Kinds is the list of supported payment scheme and network combinations
	Kinds []SupportedKind `json:"kinds"`
}

// DiscoveredResource represents a discoverable x402 resource
// as specified in section 8.2 of the x402 specification
type DiscoveredResource struct {
	// Resource is the resource URL or identifier being monetized
	Resource string `json:"resource"`

	// Type is the resource type (currently "http" for HTTP endpoints)
	Type string `json:"type"`

	// X402Version is the protocol version supported by the resource
	X402Version int `json:"x402Version"`

	// Accepts is an array of PaymentRequirements objects specifying payment methods
	Accepts []PaymentRequirements `json:"accepts"`

	// LastUpdated is the Unix timestamp of when the resource was last updated
	LastUpdated int64 `json:"lastUpdated"`

	// Metadata contains additional metadata (category, provider, etc.)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Pagination contains pagination information for discovery responses
// as specified in section 8.1 of the x402 specification
type Pagination struct {
	// Limit is the maximum number of results returned
	Limit int `json:"limit"`

	// Offset is the number of results skipped
	Offset int `json:"offset"`

	// Total is the total number of available results
	Total int `json:"total"`
}

// DiscoveryResponse is the response from the discovery API
// as specified in section 8.1 of the x402 specification
type DiscoveryResponse struct {
	// X402Version is the protocol version identifier
	X402Version int `json:"x402Version"`

	// Items is the list of discovered resources
	Items []DiscoveredResource `json:"items"`

	// Pagination contains pagination information
	Pagination Pagination `json:"pagination"`
}

// ErrorCode constants as specified in section 9 of the x402 specification
const (
	ErrorInsufficientFunds                              = "insufficient_funds"
	ErrorInvalidExactEVMPayloadAuthorizationValidAfter  = "invalid_exact_evm_payload_authorization_valid_after"
	ErrorInvalidExactEVMPayloadAuthorizationValidBefore = "invalid_exact_evm_payload_authorization_valid_before"
	ErrorInvalidExactEVMPayloadAuthorizationValue       = "invalid_exact_evm_payload_authorization_value"
	ErrorInvalidExactEVMPayloadSignature                = "invalid_exact_evm_payload_signature"
	ErrorInvalidExactEVMPayloadRecipientMismatch        = "invalid_exact_evm_payload_recipient_mismatch"
	ErrorInvalidNetwork                                 = "invalid_network"
	ErrorInvalidPayload                                 = "invalid_payload"
	ErrorInvalidPaymentRequirements                     = "invalid_payment_requirements"
	ErrorInvalidScheme                                  = "invalid_scheme"
	ErrorUnsupportedScheme                              = "unsupported_scheme"
	ErrorInvalidX402Version                             = "invalid_x402_version"
	ErrorInvalidTransactionState                        = "invalid_transaction_state"
	ErrorUnexpectedVerifyError                          = "unexpected_verify_error"
	ErrorUnexpectedSettleError                          = "unexpected_settle_error"
)
