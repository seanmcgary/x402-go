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
      "forkL1Block": "$anvilL1StartBlock"
}
EOF

