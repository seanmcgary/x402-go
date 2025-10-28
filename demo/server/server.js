const express = require('express');
const cors = require('cors');
const axios = require('axios');
require('dotenv').config();

const app = express();
const PORT = process.env.PORT || 3000;
const FACILITATOR_URL = process.env.FACILITATOR_URL || 'http://localhost:8080';
const RESOURCE_WALLET = process.env.RESOURCE_WALLET || '0x9796984630b00D0E6199f046932E6e50e3c2cBD1';
const USDC_ADDRESS = process.env.USDC_ADDRESS || '0x036CbD53842c5426634e7929541eC2318f3dCF7e'; // Base Sepolia USDC

app.use(cors());
app.use(express.json());
app.use(express.static('../client'));

// Payment requirements for protected endpoint
const paymentRequirements = {
  scheme: 'exact',
  network: 'base-sepolia',
  maxAmountRequired: '10000', // 0.01 USDC (6 decimals)
  asset: USDC_ADDRESS,
  payTo: RESOURCE_WALLET,
  resource: 'http://localhost:3000/api/protected-data',
  description: 'Access to premium data API',
  maxTimeoutSeconds: 60,
  extra: {
    name: 'USDC',
    version: '2'
  }
};

// Middleware to check for payment
async function requirePayment(req, res, next) {
  const paymentHeader = req.headers['x-payment'];

  if (!paymentHeader) {
    // No payment provided - return 402 with payment requirements
    return res.status(402).json({
      x402Version: 1,
      error: 'Payment required to access this resource',
      accepts: [paymentRequirements]
    });
  }

  try {
    // Parse payment payload
    const paymentPayload = JSON.parse(paymentHeader);

    // Verify payment with facilitator
    const verifyResponse = await axios.post(`${FACILITATOR_URL}/verify`, {
      paymentPayload,
      paymentRequirements
    });

    if (!verifyResponse.data.isValid) {
      return res.status(402).json({
        x402Version: 1,
        error: `Payment verification failed: ${verifyResponse.data.invalidReason}`,
        accepts: [paymentRequirements]
      });
    }

    console.log(`✓ Payment verified for payer: ${verifyResponse.data.payer}`);

    // Settle payment with facilitator
    const settleResponse = await axios.post(`${FACILITATOR_URL}/settle`, {
      paymentPayload,
      paymentRequirements
    });

    if (!settleResponse.data.success) {
      return res.status(402).json({
        x402Version: 1,
        error: `Payment settlement failed: ${settleResponse.data.errorReason}`,
        accepts: [paymentRequirements]
      });
    }

    console.log(`✓ Payment settled! Transaction: ${settleResponse.data.transaction}`);

    // Store settlement info for response
    req.x402Settlement = settleResponse.data;
    next();

  } catch (error) {
    console.error('Payment processing error:', error.message);
    return res.status(500).json({
      error: 'Payment processing failed',
      details: error.message
    });
  }
}

// Public endpoint - no payment required
app.get('/api/health', (req, res) => {
  res.json({
    status: 'healthy',
    timestamp: Date.now()
  });
});

// Get payment requirements
app.get('/api/payment-info', (req, res) => {
  res.json({
    x402Version: 1,
    accepts: [paymentRequirements]
  });
});

// Protected endpoint - requires payment
app.get('/api/protected-data', requirePayment, (req, res) => {
  res.json({
    success: true,
    data: {
      message: 'Welcome to the premium data!',
      secret: 'This is protected content that required payment to access.',
      timestamp: Date.now(),
      premium: true,
      insights: [
        'Market prediction: Bullish trend expected',
        'Volume analysis: High trading activity detected',
        'Sentiment: Positive outlook from major holders'
      ]
    },
    x402: req.x402Settlement
  });
});

// List of free items
app.get('/api/free-items', (req, res) => {
  res.json({
    items: [
      { id: 1, name: 'Free Sample', description: 'No payment required' },
      { id: 2, name: 'Public Data', description: 'Available to everyone' }
    ]
  });
});

// Paid item - requires payment
app.get('/api/premium-item/:id', requirePayment, (req, res) => {
  const items = {
    '1': { id: 1, name: 'Premium Analysis', data: 'Detailed market analysis...' },
    '2': { id: 2, name: 'Exclusive Report', data: 'Insider insights...' },
    '3': { id: 3, name: 'Pro Features', data: 'Advanced functionality...' }
  };

  const item = items[req.params.id];
  if (!item) {
    return res.status(404).json({ error: 'Item not found' });
  }

  res.json({
    success: true,
    item,
    x402: req.x402Settlement
  });
});

app.listen(PORT, () => {
  console.log(`
╔════════════════════════════════════════════════════════╗
║  x402 Demo Resource Server                             ║
╚════════════════════════════════════════════════════════╝

🚀 Server running on http://localhost:${PORT}

📋 Configuration:
   Facilitator URL: ${FACILITATOR_URL}
   Resource Wallet: ${RESOURCE_WALLET}
   USDC Address:    ${USDC_ADDRESS}
   Payment Amount:  0.01 USDC (10000 units)

🔗 Endpoints:
   GET  /api/health           - Health check (free)
   GET  /api/payment-info     - Get payment requirements
   GET  /api/free-items       - Public data (free)
   GET  /api/protected-data   - Premium data (requires payment)
   GET  /api/premium-item/:id - Premium items (requires payment)

🌐 Client:
   http://localhost:${PORT}/

⚡ Ready to accept x402 payments!
  `);
});
