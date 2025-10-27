package config

import (
	"fmt"
	"os"

	"github.com/seanmcgary/x402-go/pkg/chains"
)

// ChainConfig holds the configuration for a specific blockchain network
type ChainConfig struct {
	// NetworkID is the network identifier (e.g., "base-sepolia", "base")
	NetworkID string

	// RPCURL is the RPC endpoint URL for the chain
	RPCURL string

	// Enabled indicates whether this chain is enabled
	Enabled bool
}

// Config holds the runtime configuration for the facilitator
type Config struct {
	// Server configuration
	Server ServerConfig

	// Chains contains configuration for all supported blockchain networks
	Chains []ChainConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	// Host is the host to bind to (default: 0.0.0.0)
	Host string

	// Port is the port to listen on (default: 8080)
	Port int
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port number: %d (must be between 1 and 65535)", c.Server.Port)
	}

	// Ensure at least one chain is enabled
	hasEnabledChain := false
	for _, chain := range c.Chains {
		if chain.Enabled {
			hasEnabledChain = true
			break
		}
	}

	if !hasEnabledChain {
		return fmt.Errorf("at least one chain must be enabled")
	}

	// Validate each enabled chain
	for _, chain := range c.Chains {
		if !chain.Enabled {
			continue
		}

		if chain.RPCURL == "" {
			return fmt.Errorf("RPC URL is required for enabled chain: %s", chain.NetworkID)
		}

		if chain.NetworkID == "" {
			return fmt.Errorf("network ID is required for chain")
		}
	}

	return nil
}

// GetEnabledChains returns only the enabled chain configurations
func (c *Config) GetEnabledChains() []ChainConfig {
	enabled := make([]ChainConfig, 0)
	for _, chain := range c.Chains {
		if chain.Enabled {
			enabled = append(enabled, chain)
		}
	}
	return enabled
}

// BuildChainRegistry creates a chain registry from the configuration
func (c *Config) BuildChainRegistry() (*chains.Registry, error) {
	registry := chains.NewRegistry()

	for _, chainCfg := range c.GetEnabledChains() {
		var chain chains.Chain

		switch chainCfg.NetworkID {
		case chains.NetworkBaseSepolia:
			chain = chains.NewBaseSepolia(chainCfg.RPCURL)
		case chains.NetworkBase:
			chain = chains.NewBase(chainCfg.RPCURL)
		case chains.NetworkAvalancheFuji:
			chain = chains.NewAvalancheFuji(chainCfg.RPCURL)
		case chains.NetworkAvalanche:
			chain = chains.NewAvalanche(chainCfg.RPCURL)
		case chains.NetworkEthereum:
			chain = chains.NewEthereum(chainCfg.RPCURL)
		case chains.NetworkEthereumSepolia:
			chain = chains.NewEthereumSepolia(chainCfg.RPCURL)
		default:
			return nil, fmt.Errorf("unsupported network ID: %s", chainCfg.NetworkID)
		}

		registry.Register(chain)
	}

	return registry, nil
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host: getEnvOrDefault("X402_HOST", "0.0.0.0"),
			Port: getEnvAsIntOrDefault("X402_PORT", 8080),
		},
		Chains: []ChainConfig{},
	}

	// Load chain configurations from environment variables
	// Each chain can be configured with X402_<NETWORK>_RPC_URL and X402_<NETWORK>_ENABLED

	// Base Sepolia
	if rpcURL := os.Getenv("X402_BASE_SEPOLIA_RPC_URL"); rpcURL != "" {
		config.Chains = append(config.Chains, ChainConfig{
			NetworkID: chains.NetworkBaseSepolia,
			RPCURL:    rpcURL,
			Enabled:   getEnvAsBoolOrDefault("X402_BASE_SEPOLIA_ENABLED", true),
		})
	}

	// Base
	if rpcURL := os.Getenv("X402_BASE_RPC_URL"); rpcURL != "" {
		config.Chains = append(config.Chains, ChainConfig{
			NetworkID: chains.NetworkBase,
			RPCURL:    rpcURL,
			Enabled:   getEnvAsBoolOrDefault("X402_BASE_ENABLED", true),
		})
	}

	// Avalanche Fuji
	if rpcURL := os.Getenv("X402_AVALANCHE_FUJI_RPC_URL"); rpcURL != "" {
		config.Chains = append(config.Chains, ChainConfig{
			NetworkID: chains.NetworkAvalancheFuji,
			RPCURL:    rpcURL,
			Enabled:   getEnvAsBoolOrDefault("X402_AVALANCHE_FUJI_ENABLED", true),
		})
	}

	// Avalanche
	if rpcURL := os.Getenv("X402_AVALANCHE_RPC_URL"); rpcURL != "" {
		config.Chains = append(config.Chains, ChainConfig{
			NetworkID: chains.NetworkAvalanche,
			RPCURL:    rpcURL,
			Enabled:   getEnvAsBoolOrDefault("X402_AVALANCHE_ENABLED", true),
		})
	}

	// Ethereum
	if rpcURL := os.Getenv("X402_ETHEREUM_RPC_URL"); rpcURL != "" {
		config.Chains = append(config.Chains, ChainConfig{
			NetworkID: chains.NetworkEthereum,
			RPCURL:    rpcURL,
			Enabled:   getEnvAsBoolOrDefault("X402_ETHEREUM_ENABLED", true),
		})
	}

	// Ethereum Sepolia
	if rpcURL := os.Getenv("X402_ETHEREUM_SEPOLIA_RPC_URL"); rpcURL != "" {
		config.Chains = append(config.Chains, ChainConfig{
			NetworkID: chains.NetworkEthereumSepolia,
			RPCURL:    rpcURL,
			Enabled:   getEnvAsBoolOrDefault("X402_ETHEREUM_SEPOLIA_ENABLED", true),
		})
	}

	return config, nil
}

// NewDefaultConfig returns a default configuration with common public RPC endpoints
func NewDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Chains: []ChainConfig{
			{
				NetworkID: chains.NetworkBaseSepolia,
				RPCURL:    "https://sepolia.base.org",
				Enabled:   true,
			},
			{
				NetworkID: chains.NetworkBase,
				RPCURL:    "https://mainnet.base.org",
				Enabled:   true,
			},
		},
	}
}

// Helper functions for environment variable parsing

func getEnvOrDefault(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsIntOrDefault(key string, defaultVal int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}

	var value int
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		return defaultVal
	}

	return value
}

func getEnvAsBoolOrDefault(key string, defaultVal bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}

	switch valueStr {
	case "true", "TRUE", "True", "1", "yes", "YES", "Yes":
		return true
	case "false", "FALSE", "False", "0", "no", "NO", "No":
		return false
	default:
		return defaultVal
	}
}
