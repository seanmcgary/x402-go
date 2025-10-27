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
- [ ] Create pkg/blockchain package:
  - [ ] EVM client interface
  - [ ] ERC-20 contract interface (EIP-3009 compliant)
  - [ ] Balance checking functionality
  - [ ] Transaction simulation (eth_call)
  - [ ] Transaction execution (transferWithAuthorization)
  - [ ] Transaction receipt polling
- [ ] Implement connection pooling for RPC endpoints
- [ ] Add error handling for network failures
- [ ] Write unit tests with mocked blockchain interactions:
  - [ ] Generate mocks with mockery
  - [ ] Test balance checks
  - [ ] Test transaction simulation
  - [ ] Test transaction execution
- [ ] Verify tests pass with `make test`

---

## Milestone 5: Payment Verification Logic

**Objective**: Implement the exact scheme verification without blockchain execution

### Tasks:
- [ ] Create pkg/schemes/exact package:
  - [ ] Verifier interface
  - [ ] ExactSchemeVerifier implementation
  - [ ] Implement all verification steps from spec section 6.1.2:
    1. Signature validation
    2. Balance verification
    3. Amount validation
    4. Time window check
    5. Parameter matching
    6. Transaction simulation
- [ ] Implement error code mapping (section 9 of spec):
  - [ ] insufficient_funds
  - [ ] invalid_exact_evm_payload_signature
  - [ ] invalid_exact_evm_payload_authorization_valid_before/after
  - [ ] invalid_exact_evm_payload_authorization_value
  - [ ] invalid_exact_evm_payload_recipient_mismatch
  - [ ] unexpected_verify_error
- [ ] Write comprehensive unit tests:
  - [ ] Test each verification step independently
  - [ ] Test all error conditions
  - [ ] Test with testnet and mainnet configurations
- [ ] Verify tests pass with `make test`

---

## Milestone 6: Payment Settlement Logic

**Objective**: Implement blockchain transaction execution

### Tasks:
- [ ] Extend pkg/schemes/exact package:
  - [ ] Settler interface
  - [ ] ExactSchemeSettler implementation
  - [ ] Transaction submission to blockchain
  - [ ] Transaction receipt retrieval
  - [ ] Success/failure determination
- [ ] Implement settlement error handling:
  - [ ] Network errors
  - [ ] Transaction revert errors
  - [ ] Gas estimation failures
  - [ ] Nonce conflicts
- [ ] Write unit tests with mocked blockchain:
  - [ ] Test successful settlement
  - [ ] Test failed settlement scenarios
  - [ ] Test transaction receipt parsing
- [ ] Verify tests pass with `make test`

---

## Milestone 7: HTTP API Endpoints

**Objective**: Implement the three facilitator REST endpoints

### Tasks:
- [ ] Create pkg/facilitator package:
  - [ ] FacilitatorService struct
  - [ ] Service constructor with dependencies
  - [ ] Business logic for verify/settle/supported operations
- [ ] Create pkg/transport/http package:
  - [ ] HTTP handler implementations
  - [ ] Request validation and parsing
  - [ ] Response formatting
  - [ ] Error response handling
- [ ] Implement POST /verify endpoint (spec section 7.1):
  - [ ] Parse VerifyRequest
  - [ ] Validate request structure
  - [ ] Call verification logic
  - [ ] Return VerifyResponse
- [ ] Implement POST /settle endpoint (spec section 7.2):
  - [ ] Parse SettleRequest
  - [ ] Validate request structure
  - [ ] Call settlement logic
  - [ ] Return SettlementResponse
- [ ] Implement GET /supported endpoint (spec section 7.3):
  - [ ] Return list of supported schemes and networks
  - [ ] Format response per spec
- [ ] Add middleware for:
  - [ ] Request logging
  - [ ] Error recovery
  - [ ] Content-type validation
- [ ] Write HTTP integration tests:
  - [ ] Test each endpoint with valid requests
  - [ ] Test error cases
  - [ ] Test malformed requests
- [ ] Verify tests pass with `make test`

---

## Milestone 8: Discovery API

**Objective**: Implement the Bazaar discovery endpoint

### Tasks:
- [ ] Create pkg/discovery package:
  - [ ] DiscoveryService interface
  - [ ] Resource storage (in-memory for MVP, extensible for DB)
  - [ ] Query filtering (type, limit, offset)
  - [ ] Pagination logic
- [ ] Create discovery types:
  - [ ] DiscoveredResource struct
  - [ ] DiscoveryResponse struct
  - [ ] Pagination struct
- [ ] Implement GET /discovery/resources endpoint (spec section 8.1):
  - [ ] Parse query parameters
  - [ ] Apply filters
  - [ ] Implement pagination
  - [ ] Return formatted response
- [ ] Write unit tests:
  - [ ] Test filtering by type
  - [ ] Test pagination
  - [ ] Test empty results
- [ ] Verify tests pass with `make test`

---

## Milestone 9: CLI Implementation

**Objective**: Implement the facilitator binary with urfave/cli

### Tasks:
- [ ] Implement cmd/facilitator/main.go:
  - [ ] CLI app initialization with urfave/cli
  - [ ] Version command displaying version and commit
  - [ ] Serve command with flags:
    - [ ] --port (default 8080)
    - [ ] --host (default 0.0.0.0)
    - [ ] --ethereum-rpc (RPC endpoint for Ethereum)
    - [ ] --base-rpc (RPC endpoint for Base)
    - [ ] --base-sepolia-rpc (RPC endpoint for Base Sepolia)
  - [ ] Configuration loading from flags and env vars
  - [ ] Server initialization
  - [ ] Graceful shutdown handling
- [ ] Add server startup logging:
  - [ ] Log version and commit
  - [ ] Log listening address
  - [ ] Log supported networks
- [ ] Test binary execution:
  - [ ] Build with `make build/cmd/facilitator`
  - [ ] Test --help output
  - [ ] Test --version output
  - [ ] Test serve command starts server

---

## Milestone 10: Integration & End-to-End Testing

**Objective**: Comprehensive integration testing and validation

### Tasks:
- [ ] Write integration tests:
  - [ ] Full verify flow with mock blockchain
  - [ ] Full settle flow with mock blockchain
  - [ ] Test supported endpoint
  - [ ] Test discovery endpoint
- [ ] Write end-to-end tests (optional, with testnet):
  - [ ] Deploy test server
  - [ ] Test actual verification with testnet data
  - [ ] Test actual settlement with testnet data
- [ ] Verify all tests pass with `make test`
- [ ] Verify code formatting with `make fmtcheck`
- [ ] Verify linting passes with `make lint`

---

## Milestone 11: Documentation & Polish

**Objective**: Final touches and documentation

### Tasks:
- [ ] Update README.md with:
  - [ ] Project overview
  - [ ] Installation instructions
  - [ ] Usage examples
  - [ ] Configuration options
  - [ ] API documentation links
- [ ] Add code comments for public APIs
- [ ] Add example configurations
- [ ] Add example requests/responses
- [ ] Final verification:
  - [ ] All tests pass (`make test`)
  - [ ] Code is formatted (`make fmt && make fmtcheck`)
  - [ ] Linting passes (`make lint`)
  - [ ] Binary builds successfully (`make build/cmd`)
  - [ ] Binary runs and responds to requests

---

## Progress Tracking

**Current Status**: In Progress

**Completed Milestones**: 3/11

**Latest Completion**: Milestone 3 - Cryptography & Signature Validation ✓

**Next Actions**: Begin Milestone 4 - Blockchain Interaction Layer
