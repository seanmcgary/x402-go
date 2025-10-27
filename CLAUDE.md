# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

x402-go is a Go implementation of the x402 protocol for internet-native payments. The x402 protocol enables clients to pay for external resources using blockchain-based micropayments with standardized message formats and payment flows.

## Protocol Architecture

x402 is composed of three core components:

1. **Types**: Core data structures (`PaymentRequirements`, `PaymentPayload`, `SettlementResponse`) that are independent of transport mechanism and payment scheme
2. **Logic**: Payment formation and verification logic that depends on the payment scheme (exact, deferred) and network (EVM, Solana, etc.)
3. **Representation**: How payment data is transmitted and signaled, depending on transport mechanism (HTTP, MCP, A2A)

## Core Payment Flow

1. **Client Request**: Client makes a request to a resource server
2. **Payment Required Response**: Server responds with payment requirements if no valid payment attached
3. **Payment Authorization Request**: Client submits signed payment authorization
4. **Settlement Response**: Server verifies payment and initiates blockchain settlement

## Key Protocol Components

### Payment Schemes

- **Exact Scheme**: Uses EIP-3009 (Transfer with Authorization) for gasless ERC-20 token transfers
  - Requires EIP-712 signature validation
  - Balance and amount verification
  - Time window checks (validAfter/validBefore)
  - Nonce-based replay attack prevention

### Facilitator Interface

The facilitator provides HTTP REST APIs:

- `POST /verify`: Verify payment authorization without blockchain execution
- `POST /settle`: Execute verified payment on blockchain
- `GET /supported`: List supported payment schemes and networks

### Discovery API

- `GET /discovery/resources`: List discoverable x402 resources ("Bazaar" marketplace)
  - Supports filtering by type, pagination
  - Returns payment requirements, metadata, and resource details

## Supported Networks

- `base-sepolia`: Base Sepolia testnet (Chain ID: 84532)
- `base`: Base mainnet (Chain ID: 8453)
- `avalanche-fuji`: Avalanche Fuji testnet (Chain ID: 43113)
- `avalanche`: Avalanche mainnet (Chain ID: 43114)
- `iotex`: IoTeX mainnet

## Supported Assets

- **USDC**: USD Coin (EIP-3009 compliant ERC-20)
- Additional ERC-20 tokens that implement EIP-3009

## Common Development Commands

The project uses a Makefile for common tasks:

```bash
# Install dependencies and tools (mockery, golangci-lint)
make deps

# Build the facilitator binary (outputs to ./bin/facilitator)
make build/cmd/facilitator

# Build all binaries
make build/cmd

# Run all tests (with -count=1 to disable caching, -p 1 for serial execution)
make test

# Format code
make fmt

# Check code formatting (CI-friendly, exits 1 if unformatted)
make fmtcheck

# Lint code (5 minute timeout)
make lint

# Generate mocks
make mocks

# Run specific test
go test -run TestName ./path/to/package

# Run tests with coverage
go test -cover ./...
```

### Binary Execution

```bash
# Run the facilitator server
./bin/facilitator [options]
```

The facilitator binary uses `urfave/cli` for command-line parsing.

### Version Management

Version information is injected at build time via ldflags:
- Version is read from a `VERSION` file (currently not present, will need to be created)
- Commit hash is retrieved from git
- Access via `internal/version` package: `version.GetVersion()` and `version.GetCommit()`

## Implementation Notes

### Security Considerations

- **Replay Attack Prevention**: Each authorization includes unique 32-byte nonce, blockchain-level nonce reuse prevention, time constraints, and signature verification
- **EIP-3009 Compliance**: All exact scheme implementations must follow EIP-3009 standard
- **Signature Validation**: Use EIP-712 typed data signing for all payment authorizations

### Error Codes

Key error codes to handle:
- `insufficient_funds`: Not enough tokens
- `invalid_exact_evm_payload_signature`: Invalid signature
- `invalid_exact_evm_payload_authorization_valid_before/after`: Time window issues
- `invalid_exact_evm_payload_authorization_value`: Insufficient payment amount
- `invalid_network`, `invalid_scheme`: Unsupported configuration
- `unexpected_verify_error`, `unexpected_settle_error`: System errors

### Data Types

All JSON schemas are strictly typed with required fields:

**PaymentRequirements**: `scheme`, `network`, `maxAmountRequired`, `asset`, `payTo`, `resource`, `description`, `maxTimeoutSeconds` (all required)

**PaymentPayload**: `x402Version`, `scheme`, `network`, `payload` (all required)
- Payload contains `signature` and `authorization` object
- Authorization: `from`, `to`, `value`, `validAfter`, `validBefore`, `nonce` (all required)

**SettlementResponse**: `success`, `transaction`, `network`, `payer` (all required), optional `errorReason`

## Project Structure

```
x402-go/
├── cmd/
│   └── facilitator/       # Main facilitator binary entrypoint
├── internal/
│   └── version/          # Version and build information (set via ldflags)
├── pkg/                  # Public packages (to be implemented)
├── docs/
│   ├── x402-specification.md  # Full protocol specification
│   └── 001_facilitator.md     # Facilitator implementation requirements
└── bin/                  # Compiled binaries (gitignored)
```

### Package Organization Guidelines

When implementing, organize code following Go standards:

- `cmd/`: Command-line binaries (currently `facilitator`)
- `internal/`: Private application code not meant for import by other projects
- `pkg/`: Public libraries that can be imported by external projects
  - Suggested packages: `types/`, `schemes/`, `facilitator/`, `transport/`, `discovery/`, `crypto/`, `validation/`

### Implementation Requirements

From `docs/001_facilitator.md`:

- **HTTP Router**: Use `gorilla/mux` for HTTP routing
- **CLI Framework**: Use `urfave/cli` for command-line parsing
- **Initial Chain Support**: Ethereum and Base (designed for easy expansion)
- **Build System**: Makefile with version info injected via ldflags
- **Testing**: Follow idiomatic Go practices with comprehensive test coverage

## Testing Approach

Tests run with specific flags to ensure consistency:
- `-count=1`: Disable test caching to ensure fresh runs
- `-p 1 -parallel 1`: Run tests serially (important for integration tests)

Focus areas:
- Test against both testnet (base-sepolia, avalanche-fuji) and mainnet configurations
- Mock blockchain interactions for unit tests (use `mockery` for generating mocks)
- Verify EIP-3009 compliance with actual contracts
- Test time window edge cases (validAfter/validBefore)
- Validate signature verification thoroughly
- Test all error conditions defined in the spec

## References

- x402 specification: `docs/x402-specification.md`
- Original implementation: https://github.com/coinbase/x402
- EIP-3009: https://eips.ethereum.org/EIPS/eip-3009
- EIP-712: https://eips.ethereum.org/EIPS/eip-712
