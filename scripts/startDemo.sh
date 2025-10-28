#!/bin/bash

set -e

echo "🚀 Starting x402 Demo"
echo "===================="
echo ""

# Check if node_modules exists
if [ ! -d "demo/server/node_modules" ]; then
    echo "📦 Installing dependencies..."
    cd demo/server
    npm install
    cd ../..
    echo "✅ Dependencies installed"
fi

# Create .env if it doesn't exist
if [ ! -f "demo/server/.env" ]; then
    echo "📝 Creating .env file from .env.example..."
    cp demo/server/.env.example demo/server/.env
    echo "✅ .env created - you may want to customize it"
fi

echo ""
echo "Starting services..."
echo ""

# Get executor key from environment or use default test key
# NOTE: This test key should have ETH on Base Sepolia for gas
EXECUTOR_KEY="${X402_EXECUTOR_KEY:-}"

if [[ -z "$EXECUTOR_KEY" ]]; then
    echo "⚠️  No executor key found in environment variable X402_EXECUTOR_KEY"
    echo "   Using default test executor key (DO NOT USE IN PRODUCTION)"
    exit 1
fi

# Start facilitator in background using go run (always runs latest code)
echo "🔧 Starting facilitator on port 8080..."
echo "   Executor key: ${EXECUTOR_KEY:0:10}... (first 10 chars)"
go run ./cmd/facilitator serve \
  --base-sepolia-rpc https://sepolia.base.org \
  --executor-key "$EXECUTOR_KEY" \
  --port 8080 &
FACILITATOR_PID=$!

# Wait for facilitator to start
sleep 2

# Cleanup function
cleanup() {
    echo ""
    echo "🛑 Shutting down services..."
    kill $FACILITATOR_PID 2>/dev/null || true
    exit
}
trap cleanup EXIT INT TERM

# Start resource server
echo "🌐 Starting resource server on port 3000..."
echo ""
cd demo/server
npm start

# This will block until Ctrl+C
