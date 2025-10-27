# x402-go

Go implementation of the [x402 protocol](https://github.com/coinbase/x402) for internet-native payments.

## Overview

x402-go is a facilitator service that handles payment verification and settlement for the x402 protocol. It provides HTTP REST APIs for verifying and executing blockchain-based micropayments using EIP-3009 (Transfer with Authorization) for gasless ERC-20 token transfers.

## Quick Start

```bash
# Build
make build/cmd/facilitator

# Run with Base Sepolia testnet
./bin/facilitator serve --base-sepolia-rpc https://sepolia.base.org

# Run with multiple networks
./bin/facilitator serve \
  --base-sepolia-rpc https://sepolia.base.org \
  --base-rpc https://mainnet.base.org \
  --ethereum-rpc https://eth.llamarpc.com \
  --executor-key YOUR_PRIVATE_KEY_HEX
```

## API Endpoints

### POST /verify

Verifies a payment authorization without executing it on the blockchain. Performs signature validation, balance checks, amount verification, time window validation, and transaction simulation.

### POST /settle

Executes a verified payment by broadcasting the transaction to the blockchain. Returns the transaction hash upon successful settlement.

### GET /supported

Returns the list of payment schemes and networks supported by this facilitator instance.

### GET /discovery/resources

Lists discoverable x402-enabled resources (the "Bazaar" marketplace) with support for type filtering and pagination.

See [API documentation](#api-documentation) below for detailed request/response schemas.

## Configuration

### Command Line Flags

```
--host value                 Host to bind to (default: "0.0.0.0")
--port value                 Port to listen on (default: 8080)
--base-sepolia-rpc value     RPC endpoint for Base Sepolia
--base-rpc value             RPC endpoint for Base mainnet
--ethereum-rpc value         RPC endpoint for Ethereum mainnet
--ethereum-sepolia-rpc value RPC endpoint for Ethereum Sepolia
--ethereum-holesky-rpc value RPC endpoint for Ethereum Holesky
--executor-key value         Private key for executing transactions (hex format)
```

### Environment Variables

All flags can be set via environment variables:

```bash
export X402_HOST=0.0.0.0
export X402_PORT=8080
export X402_BASE_SEPOLIA_RPC_URL=https://sepolia.base.org
export X402_BASE_RPC_URL=https://mainnet.base.org
export X402_ETHEREUM_RPC_URL=https://eth.llamarpc.com
export X402_ETHEREUM_SEPOLIA_RPC_URL=https://ethereum-sepolia.publicnode.com
export X402_ETHEREUM_HOLESKY_RPC_URL=https://ethereum-holesky.publicnode.com
export X402_EXECUTOR_KEY=your_private_key_hex

./bin/facilitator serve
```

## Deployment

### Production Deployment

```bash
# Using systemd
./bin/facilitator serve \
  --base-rpc https://mainnet.base.org \
  --ethereum-rpc https://eth.llamarpc.com \
  --executor-key $EXECUTOR_PRIVATE_KEY \
  --port 8080
```

### Docker Deployment

```dockerfile
FROM golang:1.23.6-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache make git && make build/cmd/facilitator

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/bin/facilitator /usr/local/bin/facilitator
EXPOSE 8080
ENTRYPOINT ["facilitator"]
CMD ["serve"]
```

```bash
# Build and run
docker build -t x402-facilitator .
docker run -p 8080:8080 \
  -e X402_BASE_RPC_URL=https://mainnet.base.org \
  -e X402_EXECUTOR_KEY=$EXECUTOR_PRIVATE_KEY \
  x402-facilitator
```

## Supported Networks

- **base-sepolia**: Base Sepolia testnet (Chain ID: 84532)
- **base**: Base mainnet (Chain ID: 8453)
- **ethereum**: Ethereum mainnet (Chain ID: 1)
- **ethereum-sepolia**: Ethereum Sepolia testnet (Chain ID: 11155111)
- **ethereum-holesky**: Ethereum Holesky testnet (Chain ID: 17000)

The chain registry system makes it easy to add additional networks.

## Security Features

- **EIP-712 Signature Validation**: Cryptographic verification of payment authorizations
- **Replay Attack Prevention**: Unique 32-byte nonces enforced at blockchain level
- **Time Window Enforcement**: Authorizations expire after validBefore timestamp
- **Balance Verification**: Checks payer has sufficient tokens before settlement
- **Transaction Simulation**: Validates transactions will succeed before execution
- **Amount Validation**: Ensures payment amount meets requirements

## Development

### Prerequisites

- Go 1.23.6 or higher
- Make
- Foundry (for E2E tests)

### Build

```bash
make deps                  # Install dependencies
make build/cmd/facilitator # Build binary
```

### Testing

```bash
make test     # Run all tests (includes E2E with Anvil)
make fmt      # Format code
make fmtcheck # Check formatting
make lint     # Run linter
```

The test suite includes:
- 114+ unit tests across all packages
- Integration tests for full payment flows
- End-to-end tests with Anvil (automatic blockchain testing)

### Project Structure

```
x402-go/
├── cmd/facilitator/        # CLI binary
├── pkg/
│   ├── blockchain/         # EVM blockchain client
│   ├── chains/             # Network configurations
│   ├── crypto/             # EIP-712 and EIP-3009
│   ├── facilitator/        # Business logic
│   ├── schemes/exact/      # Payment scheme implementation
│   ├── transport/http/     # REST API
│   └── types/              # Protocol types
└── docs/                   # Specification documents
```

## API Documentation

### POST /verify

**Request:**
```json
{
  "paymentPayload": {
    "x402Version": 1,
    "scheme": "exact",
    "network": "base-sepolia",
    "payload": {
      "signature": "0x...",
      "authorization": {
        "from": "0x857b06519E91e3A54538791bDbb0E22373e36b66",
        "to": "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
        "value": "10000",
        "validAfter": "1740672089",
        "validBefore": "1740672154",
        "nonce": "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480"
      }
    }
  },
  "paymentRequirements": {
    "scheme": "exact",
    "network": "base-sepolia",
    "maxAmountRequired": "10000",
    "asset": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
    "payTo": "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
    "resource": "https://api.example.com/data",
    "description": "Premium data access",
    "maxTimeoutSeconds": 60
  }
}
```

**Success Response:**
```json
{
  "isValid": true,
  "payer": "0x857b06519E91e3A54538791bDbb0E22373e36b66"
}
```

**Error Response:**
```json
{
  "isValid": false,
  "invalidReason": "insufficient_funds",
  "payer": "0x857b06519E91e3A54538791bDbb0E22373e36b66"
}
```

### POST /settle

Executes a verified payment on the blockchain.

**Request:** Same as `/verify`

**Success Response:**
```json
{
  "success": true,
  "payer": "0x857b06519E91e3A54538791bDbb0E22373e36b66",
  "transaction": "0x1234567890abcdef...",
  "network": "base-sepolia"
}
```

### GET /supported

**Response:**
```json
{
  "kinds": [
    {
      "x402Version": 1,
      "scheme": "exact",
      "network": "base-sepolia"
    }
  ]
}
```

### GET /discovery/resources

Query parameters: `type`, `limit` (1-100), `offset`

**Response:**
```json
{
  "x402Version": 1,
  "items": [...],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 100
  }
}
```

## Error Codes

Standard error codes returned by the facilitator:

- `insufficient_funds` - Payer lacks sufficient token balance
- `invalid_exact_evm_payload_signature` - Invalid signature
- `invalid_exact_evm_payload_authorization_valid_after` - Not yet valid
- `invalid_exact_evm_payload_authorization_valid_before` - Expired
- `invalid_exact_evm_payload_authorization_value` - Insufficient amount
- `invalid_exact_evm_payload_recipient_mismatch` - Wrong recipient
- `invalid_network` - Unsupported network
- `invalid_scheme` - Unsupported payment scheme
- `unexpected_verify_error` - Verification error
- `unexpected_settle_error` - Settlement error

## References

- [x402 Specification](docs/x402-specification.md)
- [EIP-3009 Specification](https://eips.ethereum.org/EIPS/eip-3009)
- [EIP-712 Specification](https://eips.ethereum.org/EIPS/eip-712)
- [Original x402 Implementation](https://github.com/coinbase/x402)

## License

See LICENSE file for details.

## Version

Current version: 0.1.0

Version information is injected at build time from the VERSION file and git commit hash.
