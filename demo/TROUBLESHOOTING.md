# Demo Troubleshooting Guide

## Common Issues

### "Failed to connect to facilitator"

**Problem:** Resource server can't reach the facilitator.

**Solutions:**
1. Verify facilitator is running: `curl http://localhost:8080/supported`
2. Check facilitator logs for errors
3. Ensure port 8080 is not blocked by firewall
4. Verify `FACILITATOR_URL` in `.env` is correct

### "Please switch to Base Sepolia network"

**Problem:** MetaMask is connected to wrong network.

**Solution:**
1. Open MetaMask
2. Click network dropdown
3. Select "Base Sepolia" (or add it if missing - see README)
4. Refresh the page

### "User rejected signature"

**Problem:** User cancelled the signature request in MetaMask.

**Solution:**
- Try clicking the payment button again
- When MetaMask popup appears, click "Sign"

### "Insufficient funds" error

**Problem:** Wallet doesn't have enough USDC.

**Solutions:**
1. Check USDC balance in MetaMask
2. Get testnet USDC:
   - Bridge from Ethereum Sepolia
   - Use a Base Sepolia USDC faucet
   - Ask in Base Discord for testnet tokens

### "Payment verification failed"

**Possible causes:**
- **Expired authorization:** Try again (authorizations expire after 5 minutes)
- **Insufficient balance:** Check USDC balance
- **Wrong network:** Ensure you're on Base Sepolia (Chain ID: 84532)
- **Invalid signature:** Make sure you signed with the correct wallet

### "Settlement failed"

**Possible causes:**
- **Nonce already used:** Can't reuse the same payment authorization
- **Insufficient gas:** Need Base Sepolia ETH for transaction fees
- **Contract error:** USDC contract may have issues

### Facilitator won't start

**Problem:** Facilitator fails to start or connect to RPC.

**Solutions:**
1. Check RPC URL is accessible: `curl https://sepolia.base.org`
2. Try a different RPC endpoint (Alchemy, Infura, QuickNode)
3. Check facilitator logs for specific errors

### Resource server errors

**Problem:** Resource server crashes or returns errors.

**Solutions:**
1. Check Node.js version (should be 18+): `node --version`
2. Reinstall dependencies: `rm -rf node_modules && npm install`
3. Check `.env` file exists and has correct values
4. Look at server console for error messages

## Testing the Demo

### Test Free Endpoints First

Before testing payments, verify the setup works:

```bash
# Health check
curl http://localhost:3000/api/health

# Payment info
curl http://localhost:3000/api/payment-info

# Free data
curl http://localhost:3000/api/free-items
```

### Test Facilitator

```bash
# Check facilitator is responding
curl http://localhost:8080/supported
```

Expected response:
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

### Debug Mode

Enable detailed logging:

**Facilitator:**
```bash
# Facilitator already logs all requests
```

**Resource Server:**
Add `console.log()` statements in `server.js` to see request flow.

**Browser:**
Open DevTools (F12) → Console to see client-side logs.

## Common Development Issues

### Port Already in Use

```bash
# Find what's using port 3000
lsof -i :3000

# Kill it
kill -9 <PID>

# Or use different port
PORT=3001 npm start
```

### CORS Errors

The server already has CORS enabled. If you still see CORS errors:
1. Make sure you're accessing via `http://localhost:3000` not `file://`
2. Check browser console for specific CORS error
3. Verify server is running

### MetaMask Not Detected

- Install MetaMask browser extension
- Refresh the page after installation
- Try a different browser (Chrome, Firefox, Brave)

## Getting Help

If you encounter issues:

1. Check the console logs (browser DevTools + server terminal)
2. Verify all prerequisites are met
3. Try the manual setup to isolate which component is failing
4. Check facilitator is actually running: `curl http://localhost:8080/supported`

## Example Test Payment

You can test the payment flow manually with curl:

```bash
# 1. Get payment requirements
curl http://localhost:3000/api/payment-info

# 2. Create payment authorization (requires wallet signing)
# This must be done in the browser with MetaMask

# 3. Make paid request (example payload)
curl http://localhost:3000/api/protected-data \
  -H "X-Payment: {\"x402Version\":1,\"scheme\":\"exact\",\"network\":\"base-sepolia\",\"payload\":{...}}"
```

The browser client handles the signing and payment flow automatically.
