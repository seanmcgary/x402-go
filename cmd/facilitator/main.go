package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/seanmcgary/x402-go/internal/version"
	"github.com/seanmcgary/x402-go/pkg/blockchain"
	"github.com/seanmcgary/x402-go/pkg/chains"
	"github.com/seanmcgary/x402-go/pkg/config"
	"github.com/seanmcgary/x402-go/pkg/discovery"
	"github.com/seanmcgary/x402-go/pkg/facilitator"
	"github.com/seanmcgary/x402-go/pkg/schemes/exact"
	httpTransport "github.com/seanmcgary/x402-go/pkg/transport/http"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "facilitator",
		Usage:   "x402 facilitator for payment verification and settlement",
		Version: fmt.Sprintf("%s (commit: %s)", version.GetVersion(), version.GetCommit()),
		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "Print version information",
				Action: func(c *cli.Context) error {
					fmt.Printf("Version: %s\n", version.GetVersion())
					fmt.Printf("Commit:  %s\n", version.GetCommit())
					return nil
				},
			},
			{
				Name:  "serve",
				Usage: "Start the facilitator HTTP server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "host",
						Value:   "0.0.0.0",
						Usage:   "Host to bind to",
						EnvVars: []string{"X402_HOST"},
					},
					&cli.IntFlag{
						Name:    "port",
						Value:   8080,
						Usage:   "Port to listen on",
						EnvVars: []string{"X402_PORT"},
					},
					&cli.StringFlag{
						Name:    "base-sepolia-rpc",
						Usage:   "RPC endpoint for Base Sepolia",
						EnvVars: []string{"X402_BASE_SEPOLIA_RPC_URL"},
					},
					&cli.StringFlag{
						Name:    "base-rpc",
						Usage:   "RPC endpoint for Base",
						EnvVars: []string{"X402_BASE_RPC_URL"},
					},
					&cli.StringFlag{
						Name:    "ethereum-rpc",
						Usage:   "RPC endpoint for Ethereum",
						EnvVars: []string{"X402_ETHEREUM_RPC_URL"},
					},
					&cli.StringFlag{
						Name:    "executor-key",
						Usage:   "Private key for executing transactions (hex format without 0x prefix)",
						EnvVars: []string{"X402_EXECUTOR_KEY"},
					},
				},
				Action: serveAction,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serveAction(c *cli.Context) error {
	// Print version info
	log.Printf("x402 Facilitator v%s (commit: %s)", version.GetVersion(), version.GetCommit())

	// Build configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: c.String("host"),
			Port: c.Int("port"),
		},
		Chains: []config.ChainConfig{},
	}

	// Add chains based on flags
	if rpcURL := c.String("base-sepolia-rpc"); rpcURL != "" {
		cfg.Chains = append(cfg.Chains, config.ChainConfig{
			NetworkID: chains.NetworkBaseSepolia,
			RPCURL:    rpcURL,
			Enabled:   true,
		})
	}

	if rpcURL := c.String("base-rpc"); rpcURL != "" {
		cfg.Chains = append(cfg.Chains, config.ChainConfig{
			NetworkID: chains.NetworkBase,
			RPCURL:    rpcURL,
			Enabled:   true,
		})
	}

	if rpcURL := c.String("ethereum-rpc"); rpcURL != "" {
		cfg.Chains = append(cfg.Chains, config.ChainConfig{
			NetworkID: chains.NetworkEthereum,
			RPCURL:    rpcURL,
			Enabled:   true,
		})
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Build chain registry
	registry, err := cfg.BuildChainRegistry()
	if err != nil {
		return fmt.Errorf("failed to build chain registry: %w", err)
	}

	log.Printf("Listening on %s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Supported networks: %v", registry.List())

	// Parse executor private key
	executorKey, err := parseExecutorKey(c.String("executor-key"))
	if err != nil {
		return fmt.Errorf("failed to parse executor key: %w", err)
	}

	// Create blockchain clients and scheme handlers for each network
	verifiers := make(map[string]exact.Verifier)
	settlers := make(map[string]exact.Settler)

	for _, chain := range registry.GetAll() {
		// Create blockchain client
		client, err := blockchain.NewClient(chain.RPCURL())
		if err != nil {
			log.Printf("Warning: Failed to connect to %s: %v", chain.NetworkID(), err)
			continue
		}

		// Create verifier
		verifier := exact.NewVerifier(client)
		verifiers[chain.NetworkID()] = verifier

		// Create settler
		settler := exact.NewSettler(client, executorKey, 60*time.Second)
		settlers[chain.NetworkID()] = settler

		log.Printf("Initialized handlers for %s (%s)", chain.Name(), chain.NetworkID())
	}

	// Create facilitator service
	facilService := facilitator.NewService(verifiers, settlers, registry)

	// Create discovery service
	discoveryService := discovery.NewService()

	// Create HTTP server
	server := httpTransport.NewServer(
		cfg.Server.Host,
		cfg.Server.Port,
		facilService,
		discoveryService,
	)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for interrupt signal or error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}

// parseExecutorKey parses a hex-encoded private key
func parseExecutorKey(keyHex string) (*ecdsa.PrivateKey, error) {
	if keyHex == "" {
		// Generate a random key for development
		log.Println("Warning: No executor key provided, generating random key (for development only)")
		return crypto.GenerateKey()
	}

	// Parse hex key (with or without 0x prefix)
	if len(keyHex) >= 2 && keyHex[:2] == "0x" {
		keyHex = keyHex[2:]
	}

	key, err := crypto.HexToECDSA(keyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	return key, nil
}
