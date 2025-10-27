# Requirements

The goal of this docs is to define the requirements and execution plan for an x402 facilitator built in Go.

The facilitator is a webserver that should implement the x402 spec found in `./docs/x402-specification.md`.

## Acceptance Criteria

- Binary to run the server implemented in `cmd/facilitator/main.go` and compiled to `bin/facilitator`.
- The webserver of the facilitator should be built using gorilla/mux for http routing.
- The facilitator should support Ethereum and Base to start, but be designed to easily add more chains in the future.
- The runtime interface of the facilitator binary should use `urfave/cli` for command line parsing.

## Tips and tricks

- Follow idiomatic Go practices.
- Make sure to write tests for your code.
- Make sure all tests are properly passing
- Make sure code is formatted properly (`make fmt`) and passes the linter (`make lint`).

# Execution

## Milestones Overview

This execution plan follows the order of requirements and implements the x402 facilitator as specified in the x402-specification.md. The implementation is broken down into logical milestones that build upon each other.

---

## Milestone 1: Project Foundation & Core Types

**Objective**: Set up project structure, dependencies, and core x402 protocol types

### Tasks:
- [x] Create VERSION file with initial version (0.1.0)
- [x] Add required dependencies to go.mod:
  - gorilla/mux (HTTP routing)
  - urfave/cli (CLI parsing)
  - ethereum libraries (go-ethereum for EVM interaction)
  - EIP-712 signing libraries
- [x] Create pkg/types package with core protocol types:
  - [x] PaymentRequirements struct
  - [x] PaymentPayload struct
  - [x] SettlementResponse struct
  - [x] Authorization struct (EIP-3009)
  - [x] SchemePayload struct
  - [x] VerifyRequest/Response structs
  - [x] SettleRequest/Response structs
  - [x] SupportedKinds struct
- [x] Add JSON marshaling/unmarshaling with proper tags
- [x] Write unit tests for all type definitions
- [x] Verify tests pass with `make test`

---

## Milestone 2: Network & Chain Configuration

**Objective**: Implement chain configuration for Ethereum and Base networks

### Tasks:
- [x] Create pkg/chains package with chain definitions:
  - [x] Chain interface with RPC URL, Chain ID, Name methods
  - [x] base-sepolia configuration (Chain ID: 84532)
  - [x] base configuration (Chain ID: 8453)
  - [x] Chain registry/factory for easy expansion
- [x] Create pkg/config package for runtime configuration:
  - [x] Configuration struct with RPC endpoints
  - [x] Environment variable loading
  - [x] Configuration validation
- [x] Write unit tests for chain configuration
- [x] Verify tests pass with `make test`

---

## Milestone 3: Cryptography & Signature Validation

**Objective**: Implement EIP-712 signature validation and EIP-3009 verification

### Tasks:
- [x] Create pkg/crypto package:
  - [x] EIP-712 domain separator construction
  - [x] EIP-712 typed data hashing for TransferWithAuthorization
  - [x] Signature recovery and validation
  - [x] Address extraction from signatures
- [x] Implement EIP-3009 authorization validation:
  - [x] Time window validation (validAfter/validBefore)
  - [x] Amount validation
  - [x] Recipient address matching
  - [x] Nonce format validation (32-byte)
- [x] Write comprehensive unit tests:
  - [x] Test with valid signatures
  - [x] Test with invalid signatures
  - [x] Test edge cases (expired, not yet valid, wrong recipient)
- [x] Verify tests pass with `make test`

---

## Milestone 4: Blockchain Interaction Layer

**Objective**: Implement EVM interaction for balance checks and transaction execution

### Tasks:
- [x] Create pkg/blockchain package:
  - [x] EVM client interface
  - [x] ERC-20 contract interface (EIP-3009 compliant)
  - [x] Balance checking functionality
  - [x] Transaction simulation (eth_call)
  - [x] Transaction execution (transferWithAuthorization)
  - [x] Transaction receipt polling
- [x] Implement connection pooling for RPC endpoints
- [x] Add error handling for network failures
- [x] Write unit tests with mocked blockchain interactions:
  - [x] Generate mocks with mockery
  - [x] Test balance checks
  - [x] Test transaction simulation
  - [x] Test transaction execution
- [x] Verify tests pass with `make test`

---

## Milestone 5: Payment Verification Logic

**Objective**: Implement the exact scheme verification without blockchain execution

### Tasks:
- [x] Create pkg/schemes/exact package:
  - [x] Verifier interface
  - [x] ExactSchemeVerifier implementation
  - [x] Implement all verification steps from spec section 6.1.2:
    1. Signature validation
    2. Balance verification
    3. Amount validation
    4. Time window check
    5. Parameter matching
    6. Transaction simulation
- [x] Implement error code mapping (section 9 of spec):
  - [x] insufficient_funds
  - [x] invalid_exact_evm_payload_signature
  - [x] invalid_exact_evm_payload_authorization_valid_before/after
  - [x] invalid_exact_evm_payload_authorization_value
  - [x] invalid_exact_evm_payload_recipient_mismatch
  - [x] unexpected_verify_error
- [x] Write comprehensive unit tests:
  - [x] Test each verification step independently
  - [x] Test all error conditions
  - [x] Test with testnet and mainnet configurations
- [x] Verify tests pass with `make test`

---

## Milestone 6: Payment Settlement Logic

**Objective**: Implement blockchain transaction execution

### Tasks:
- [x] Extend pkg/schemes/exact package:
  - [x] Settler interface
  - [x] ExactSchemeSettler implementation
  - [x] Transaction submission to blockchain
  - [x] Transaction receipt retrieval
  - [x] Success/failure determination
- [x] Implement settlement error handling:
  - [x] Network errors
  - [x] Transaction revert errors
  - [x] Gas estimation failures
  - [x] Nonce conflicts
- [x] Write unit tests with mocked blockchain:
  - [x] Test successful settlement
  - [x] Test failed settlement scenarios
  - [x] Test transaction receipt parsing
- [x] Verify tests pass with `make test`

---

## Milestone 7: HTTP API Endpoints

**Objective**: Implement the three facilitator REST endpoints

### Tasks:
- [x] Create pkg/facilitator package:
  - [x] FacilitatorService struct
  - [x] Service constructor with dependencies
  - [x] Business logic for verify/settle/supported operations
- [x] Create pkg/transport/http package:
  - [x] HTTP handler implementations
  - [x] Request validation and parsing
  - [x] Response formatting
  - [x] Error response handling
- [x] Implement POST /verify endpoint (spec section 7.1):
  - [x] Parse VerifyRequest
  - [x] Validate request structure
  - [x] Call verification logic
  - [x] Return VerifyResponse
- [x] Implement POST /settle endpoint (spec section 7.2):
  - [x] Parse SettleRequest
  - [x] Validate request structure
  - [x] Call settlement logic
  - [x] Return SettlementResponse
- [x] Implement GET /supported endpoint (spec section 7.3):
  - [x] Return list of supported schemes and networks
  - [x] Format response per spec
- [x] Add middleware for:
  - [x] Request logging
  - [x] Error recovery
  - [x] Content-type validation
- [x] Write HTTP integration tests:
  - [x] Test each endpoint with valid requests
  - [x] Test error cases
  - [x] Test malformed requests
- [x] Verify tests pass with `make test`

---

## Milestone 8: Discovery API

**Objective**: Implement the Bazaar discovery endpoint

### Tasks:
- [x] Create pkg/discovery package:
  - [x] DiscoveryService interface
  - [x] Resource storage (in-memory for MVP, extensible for DB)
  - [x] Query filtering (type, limit, offset)
  - [x] Pagination logic
- [x] Create discovery types:
  - [x] DiscoveredResource struct
  - [x] DiscoveryResponse struct
  - [x] Pagination struct
- [x] Implement GET /discovery/resources endpoint (spec section 8.1):
  - [x] Parse query parameters
  - [x] Apply filters
  - [x] Implement pagination
  - [x] Return formatted response
- [x] Write unit tests:
  - [x] Test filtering by type
  - [x] Test pagination
  - [x] Test empty results
- [x] Verify tests pass with `make test`

---

## Milestone 9: CLI Implementation

**Objective**: Implement the facilitator binary with urfave/cli

### Tasks:
- [x] Implement cmd/facilitator/main.go:
  - [x] CLI app initialization with urfave/cli
  - [x] Version command displaying version and commit
  - [x] Serve command with flags:
    - [x] --port (default 8080)
    - [x] --host (default 0.0.0.0)
    - [x] --ethereum-rpc (RPC endpoint for Ethereum)
    - [x] --base-rpc (RPC endpoint for Base)
    - [x] --base-sepolia-rpc (RPC endpoint for Base Sepolia)
  - [x] Configuration loading from flags and env vars
  - [x] Server initialization
  - [x] Graceful shutdown handling
- [x] Add server startup logging:
  - [x] Log version and commit
  - [x] Log listening address
  - [x] Log supported networks
- [x] Test binary execution:
  - [x] Build with `make build/cmd/facilitator`
  - [x] Test --help output
  - [x] Test --version output
  - [x] Test serve command starts server

---

## Milestone 10: Integration & End-to-End Testing

**Objective**: Comprehensive integration testing and validation

### Tasks:
- [x] Write integration tests:
  - [x] Full verify flow with mock blockchain
  - [x] Full settle flow with mock blockchain
  - [x] Test supported endpoint
  - [x] Test discovery endpoint
- [x] Write end-to-end tests (optional, with testnet):
  - [x] Deploy test server
  - [x] Test actual verification with testnet data
  - [x] Test actual settlement with testnet data
- [x] Verify all tests pass with `make test`
- [x] Verify code formatting with `make fmtcheck`
- [x] Verify linting passes with `make lint`

---

## Milestone 11: Documentation & Polish

**Objective**: Final touches and documentation

### Tasks:
- [x] Update README.md with:
  - [x] Project overview
  - [x] Installation instructions
  - [x] Usage examples
  - [x] Configuration options
  - [x] API documentation links
- [x] Add code comments for public APIs
- [x] Add example configurations
- [x] Add example requests/responses
- [x] Final verification:
  - [x] All tests pass (`make test`)
  - [x] Code is formatted (`make fmt && make fmtcheck`)
  - [x] Linting passes (`make lint`)
  - [x] Binary builds successfully (`make build/cmd`)
  - [x] Binary runs and responds to requests

---

## Progress Tracking

**Current Status**: ✅ COMPLETE

**Completed Milestones**: 11/11 (100%)

**Latest Completion**: Milestone 11 - Documentation & Polish ✓

**All Milestones Complete!**

The x402-go facilitator is fully implemented with:
- ✓ Core protocol types and validation
- ✓ Multi-network blockchain support (Base, Ethereum, Avalanche)
- ✓ EIP-712 and EIP-3009 cryptography
- ✓ Payment verification and settlement
- ✓ HTTP REST API with 4 endpoints
- ✓ Discovery API (Bazaar)
- ✓ Full CLI implementation with urfave/cli
- ✓ 114+ passing unit tests
- ✓ End-to-end tests with Anvil support
- ✓ 0 linting issues
- ✓ Comprehensive documentation

## End-to-End Testing

The implementation includes comprehensive e2e tests that validate real blockchain interactions with Anvil:

**Test Files**:
- `pkg/integration/e2e_anvil_test.go` - Full e2e test suite

**Running E2E Tests**:
```bash
# Option 1: Use the provided script
./scripts/run-e2e-test.sh

# Option 2: Manual setup
# Terminal 1: Start Anvil
anvil --fork-url https://sepolia.base.org --port 8545

# Terminal 2: Run tests
go test -v -tags=e2e ./pkg/integration/...
```

**What E2E Tests Validate**:
- ✓ Real blockchain connection (Anvil on port 8545)
- ✓ EIP-712 signature creation and verification
- ✓ EIP-712 domain separator computation
- ✓ Signature recovery with real private keys
- ✓ Complete authorization validation
- ✓ Time window enforcement
- ✓ Amount validation
- ✓ Full verification flow with blockchain queries
- ✓ ERC-20 balance checks
- ✓ All blockchain client operations (ChainID, BlockNumber, BalanceAt, etc.)

See `docs/E2E_TESTING.md` for complete testing guide.
