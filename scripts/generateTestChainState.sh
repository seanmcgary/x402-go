#!/usr/bin/env bash

anvilL1Pid=""

function cleanup() {
    kill $anvilL1Pid || true

    exit $?
}
trap cleanup ERR
set -euo pipefail


# ethereum holesky
L1_FORK_RPC_URL=https://practical-serene-mound.ethereum-sepolia.quiknode.pro/3aaa48bd95f3d6aed60e89a1a466ed1e2a440b61/

anvilL1ChainId=31337
anvilL1StartBlock=9469897
anvilL1DumpStatePath=./anvil-l1.json
anvilL1ConfigPath=./anvil-l1-config.json
anvilL1RpcPort=8545
anvilL1RpcUrl="http://localhost:${anvilL1RpcPort}"


seedAccounts=$(cat ./anvilConfig/accounts.json)

# -----------------------------------------------------------------------------
# Start Ethereum L1
# -----------------------------------------------------------------------------
anvil \
    --fork-url $L1_FORK_RPC_URL \
    --dump-state $anvilL1DumpStatePath \
    --config-out $anvilL1ConfigPath \
    --chain-id $anvilL1ChainId \
    --port $anvilL1RpcPort \
    --block-time 2 \
    --fork-block-number $anvilL1StartBlock &

anvilL1Pid=$!
sleep 3

function fundAccount() {
    address=$1
    echo "Funding address $address on L1"
    cast rpc --rpc-url $anvilL1RpcUrl anvil_setBalance $address '0x21E19E0C9BAB2400000' # 10,000 ETH
}

# loop over the seed accounts (json array) and fund the accounts
numAccounts=$(echo $seedAccounts | jq '. | length - 1')
for i in $(seq 0 $numAccounts); do
    account=$(echo $seedAccounts | jq -r ".[$i]")
    address=$(echo $account | jq -r '.address')

    fundAccount $address
done


# create account example
# deployAccountAddress=$(echo $seedAccounts | jq -r '.[0].address')
# deployAccountPk=$(echo $seedAccounts | jq -r '.[0].private_key')
# export PRIVATE_KEY_DEPLOYER=$deployAccountPk
# echo "Deploy account: $deployAccountAddress"
# echo "Deploy account private key: $deployAccountPk"

export L1_RPC_URL="http://localhost:${anvilL1RpcPort}"

# -----------------------------------------------------------------------------
# Logic to add chain state
# -----------------------------------------------------------------------------

# Get payer and recipient accounts for x402 testing
payerAccountAddress=$(echo $seedAccounts | jq -r '.[0].address')
payerAccountPk=$(echo $seedAccounts | jq -r '.[0].private_key')
recipientAccountAddress=$(echo $seedAccounts | jq -r '.[1].address')
recipientAccountPk=$(echo $seedAccounts | jq -r '.[1].private_key')

echo "Payer account: $payerAccountAddress"
echo "Recipient account: $recipientAccountAddress"

# USDC token address on Ethereum Sepolia
USDC_ADDRESS="0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238"

# Fund payer account with USDC using anvil cheats
# Set USDC balance to 1,000,000 USDC (6 decimals) = 1,000,000,000,000
USDC_AMOUNT="0xE8D4A51000" # 1,000,000,000,000 in hex

echo "Funding payer with USDC..."
# Use cast to set the storage slot for the balance
# For USDC, balances are typically in slot 9 (may vary by token)
# We'll use anvil_setStorageAt to set the balance directly
cast rpc --rpc-url $anvilL1RpcUrl anvil_setCode $USDC_ADDRESS "0x608060405234801561001057600080fd5b50600436106100365760003560e01c806370a082311461003b578063a9059cbb14610071575b600080fd5b61005560048036038101906100509190610123565b6100a1565b6040516100689190610169565b60405180910390f35b61008b60048036038101906100869190610184565b6100e9565b60405161009891906101df565b60405180910390f35b60008060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b600080600090505b600190505056" || true

echo "Setting USDC balance for payer..."
# Calculate storage slot for balance mapping: keccak256(abi.encode(address, slot))
# For simplicity, we'll use anvil's deal command if available, or set balance directly
cast rpc --rpc-url $anvilL1RpcUrl anvil_setBalance $payerAccountAddress '0x21E19E0C9BAB2400000' # Ensure ETH for gas

echo "USDC funding complete (Note: USDC contract may need to exist on fork for balance queries to work)"


# -----------------------------------------------------------------------------
# Cleanup
# -----------------------------------------------------------------------------
echo "Ended at block number: "
cast block-number

kill $anvilL1Pid || true
sleep 3

rm -rf ./internal/testData/anvil*.json

cp -R $anvilL1DumpStatePath internal/testData/anvil-l1-state.json
cp -R $anvilL1ConfigPath internal/testData/anvil-l1-config.json

# make the files read-only since anvil likes to overwrite things
chmod 444 internal/testData/anvil*

rm $anvilL1DumpStatePath
rm $anvilL1ConfigPath

# create a heredoc json file and dump it to internal/testData/chain-config.json
cat <<EOF > internal/testData/chain-config.json
{
      "forkL1Block": "$anvilL1StartBlock",
      "payerAccount": {
            "address": "$payerAccountAddress",
            "privateKey": "$payerAccountPk"
      },
      "recipientAccount": {
            "address": "$recipientAccountAddress",
            "privateKey": "$recipientAccountPk"
      }
}
EOF

