# x402-go Demo

A complete demonstration of the x402 payment protocol showing how to monetize API endpoints with blockchain micropayments.

## Overview

This demo shows a real-world example of x402 in action:
- **Resource Server** (Node.js/Express) - API that charges for premium endpoints
- **Web Client** (HTML/JavaScript) - Browser app that pays to access protected data
- **Go Facilitator** - Verifies and settles payments on blockchain

## Architecture

```
┌─────────────┐         ┌─────────────────┐         ┌──────────────┐
│ Web Client  │ ──────> │ Resource Server │ ──────> │  Facilitator │
│  (Browser)  │ <────── │   (Express)     │ <────── │  (Go Service)│
└─────────────┘         └─────────────────┘         └──────────────┘
                                                            │
                                                            ▼
                                                      ┌──────────┐
                                                      │Blockchain│
                                                      └──────────┘
```

## Quick Start

### Automated (Recommended)

```bash
# From project root
./scripts/startDemo.sh
```

This will:
1. Build the facilitator if needed
2. Install Node.js dependencies
3. Start the facilitator on port 8080
4. Start the resource server on port 3000
5. Open http://localhost:3000 in your browser

### Manual Setup

**Terminal 1 - Facilitator:**
```bash
./bin/facilitator serve \
  --base-sepolia-rpc https://sepolia.base.org \
  --port 8080
```

**Terminal 2 - Resource Server:**
```bash
cd demo/server
npm install
npm start
```

**Browser:**
```
http://localhost:3000
```

## How It Works

1. **Client requests protected data** from the resource server
2. **Server responds with 402 Payment Required** and payment requirements
3. **Client creates payment authorization** (signs with wallet)
4. **Client submits payment** with the request
5. **Server verifies payment** with the facilitator
6. **Facilitator validates** signature, balance, and executes settlement
7. **Server returns protected data** to the client

## Payment Flow

```
Client: GET /api/protected-data
Server: 402 Payment Required
        {
          "x402Version": 1,
          "error": "Payment required",
          "accepts": [{
            "scheme": "exact",
            "network": "base-sepolia",
            "maxAmountRequired": "10000",
            "asset": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
            "payTo": "0x...",
            ...
          }]
        }

Client: Signs payment authorization with wallet

Client: GET /api/protected-data
        X-Payment: {payment payload with signature}

Server: Calls facilitator to verify payment

Facilitator: Validates and settles payment

Server: 200 OK
        {
          "data": "Protected content",
          "x402": {settlement response}
        }
```

## Status

✅ **This demo is fully functional!** It demonstrates the complete x402 payment flow including verification and settlement on Base Sepolia.

## Prerequisites

### Required

1. **Node.js** 18+ and npm
2. **MetaMask** browser extension
3. **Base Sepolia network** added to MetaMask
4. **Testnet tokens**:
   - Base Sepolia ETH (for gas) - Get from [faucet](https://faucet.quicknode.com/base/sepolia)
   - USDC on Base Sepolia (for payments) - Bridge from Sepolia or use faucet (needed for verification)

### Add Base Sepolia to MetaMask

- **Network Name:** Base Sepolia
- **RPC URL:** https://sepolia.base.org
- **Chain ID:** 84532
- **Currency Symbol:** ETH
- **Block Explorer:** https://sepolia.basescan.org

## Configuration

The resource server can be configured via `demo/server/.env`:

```bash
FACILITATOR_URL=http://localhost:8080  # Facilitator service URL
RESOURCE_WALLET=0x9796984630b00D0E6199f046932E6e50e3c2cBD1  # Wallet to receive payments
USDC_ADDRESS=0x036CbD53842c5426634e7929541eC2318f3dCF7e     # Base Sepolia USDC
PORT=3000                                                    # Server port
```

Copy `.env.example` to `.env` and customize as needed.

## Example Protected Resource

The demo server provides a simple protected endpoint:

**GET /api/protected-data**
- Requires payment: 0.01 USDC (10000 in atomic units)
- Returns: `{ "data": "Secret data available only with payment!" }`

## Development

```bash
# Start all services
npm run demo  # Starts facilitator + server + opens client
```

## Learn More

- [x402 Specification](../docs/x402-specification.md)
- [Facilitator Documentation](../README.md)
