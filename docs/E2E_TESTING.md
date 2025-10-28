# End-to-End Testing with Anvil

This document describes how to run end-to-end tests against a local Anvil instance to validate the x402-go facilitator with real blockchain interactions.

## Prerequisites

1. **Install Foundry** (includes Anvil):
   ```bash
   curl -L https://foundry.paradigm.xyz | bash
   foundryup
   ```

2. **Verify Anvil is installed**:
   ```bash
   anvil --version
   ```

## Running E2E Tests

### Quick Start

Use the provided script to automatically start Anvil and run tests:

```bash
./scripts/run-e2e-test.sh
```

### Manual Setup

#### Option 1: Local Anvil (Fresh Chain)

Start Anvil with default settings:

```bash
# Start Anvil on port 8545
anvil --port 8545
```

Run the e2e tests:

```bash
go test -v -tags=e2e ./pkg/integration/...
```

#### Option 2: Forked Base Sepolia (Recommended)

Fork Base Sepolia to test with real USDC contract:

```bash
# Fork Base Sepolia testnet
anvil --fork-url https://sepolia.base.org --port 8545
```

Run the e2e tests:

```bash
go test -v -tags=e2e ./pkg/integration/...
```

## What the Tests Validate

### TestE2EConnectionAndSetup
- ✓ Connects to Anvil successfully
- ✓ Retrieves chain ID and block number
- ✓ Checks Anvil default account has ETH balance
- ✓ Validates basic blockchain operations

### TestE2EVerifySignatureFlow
- ✓ Creates valid EIP-712 signatures
- ✓ Computes domain separators correctly
- ✓ Recovers signer address from signature
- ✓ Validates complete authorization structure
- ✓ Tests time window validation
- ✓ Tests amount validation
- ✓ Runs full verification flow
- ⚠ May show insufficient_funds if account has no USDC (expected)

### TestE2EVerifyWithAnvil
- ✓ Tests payment verification against live blockchain
- ✓ Validates cryptographic operations
- ✓ Checks ERC-20 balance queries
- ✓ Tests all 6 verification steps
- ✓ Validates error code handling

### TestE2EBlockchainConnection
- ✓ Tests all blockchain client methods
- ✓ Validates ChainID, BlockNumber
- ✓ Tests BalanceAt, SuggestGasPrice
- ✓ Tests PendingNonceAt

## Test Accounts

Anvil provides default test accounts with private keys:

```
Account 0:
  Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
  Private Key: ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
  Balance: 10000 ETH

Account 1:
  Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
  Private Key: 59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d
  Balance: 10000 ETH
```

## Expected Behavior

### Fresh Anvil Instance (Local Chain)

When running against a fresh Anvil instance (chain ID 31337):

- Blockchain connection tests: ✅ PASS
- Cryptographic validation tests: ✅ PASS
- Balance checks: ⚠ Will fail (no USDC contract deployed)
- Verification flow: ⚠ Returns `insufficient_funds` or contract error (expected)

### Forked Base Sepolia

When forking Base Sepolia (chain ID 84532):

- Blockchain connection tests: ✅ PASS
- Cryptographic validation tests: ✅ PASS
- Balance checks: ✅ PASS (USDC contract exists)
- Verification flow: ⚠ Returns `insufficient_funds` (Anvil accounts have no USDC)

## Testing Real Settlement (Advanced)

To test actual settlement with USDC transfers, you would need to:

1. Fork Base Sepolia mainnet (has real USDC with balances)
2. Impersonate an account with USDC balance:
   ```bash
   anvil --fork-url https://mainnet.base.org \
         --fork-block-number <recent_block> \
         --port 8545
   ```

3. Use Anvil's `anvil_impersonateAccount` RPC to impersonate a wealthy USDC holder
4. Create and execute real settlements

This is advanced usage and requires additional test setup.

## What Gets Validated

Even without USDC balances, these tests validate:

✅ **All cryptographic operations work correctly**:
- EIP-712 domain separator computation
- EIP-712 typed data hashing
- Signature creation and recovery
- Address extraction from signatures

✅ **All validation logic works correctly**:
- Time window validation
- Amount validation
- Recipient validation
- Nonce format validation

✅ **Blockchain integration works**:
- RPC connection and communication
- Contract ABI encoding/decoding
- Balance queries
- Gas estimation calls
- Nonce management

✅ **Error handling is correct**:
- Proper error codes returned
- Graceful handling of missing balances
- Network error handling

## Troubleshooting

### "Failed to connect to Anvil"

Make sure Anvil is running:
```bash
# Check if port 8545 is in use
lsof -i :8545

# Start Anvil
anvil --port 8545
```

### "Contract may not exist on this fork"

This is expected when testing against a fresh Anvil instance. The tests will still validate cryptographic operations.

To test with real contracts, fork Base Sepolia:
```bash
anvil --fork-url https://sepolia.base.org --port 8545
```

### "Insufficient funds"

This is expected behavior when Anvil test accounts don't have USDC. The test validates that the error is correctly detected and reported.

## CI/CD Integration

To run e2e tests in CI:

```yaml
- name: Start Anvil
  run: |
    anvil --port 8545 &
    sleep 5

- name: Run E2E Tests
  run: go test -v -tags=e2e ./pkg/integration/...
```

## Summary

These e2e tests validate that the x402-go facilitator:
- Correctly implements EIP-712 and EIP-3009 standards
- Properly integrates with Ethereum/EVM blockchains
- Handles all verification steps accurately
- Returns correct error codes
- Works with real blockchain infrastructure

The tests are designed to pass regardless of USDC balance, focusing on validating the correctness of the implementation rather than requiring complex test setup.
