package config

import (
	"os"
	"testing"

	"github.com/seanmcgary/x402-go/pkg/chains"
)

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host 0.0.0.0, got %s", config.Server.Host)
	}

	if config.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Server.Port)
	}

	if len(config.Chains) != 2 {
		t.Errorf("Expected 2 chains, got %d", len(config.Chains))
	}

	// Check for Base Sepolia
	hasBaseSepolia := false
	for _, chain := range config.Chains {
		if chain.NetworkID == chains.NetworkBaseSepolia {
			hasBaseSepolia = true
			if !chain.Enabled {
				t.Error("Expected Base Sepolia to be enabled")
			}
			if chain.RPCURL != "https://sepolia.base.org" {
				t.Errorf("Expected Base Sepolia RPC URL https://sepolia.base.org, got %s", chain.RPCURL)
			}
		}
	}
	if !hasBaseSepolia {
		t.Error("Expected Base Sepolia in default config")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: &Config{
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
				},
			},
			expectErr: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 0,
				},
				Chains: []ChainConfig{
					{
						NetworkID: chains.NetworkBaseSepolia,
						RPCURL:    "https://sepolia.base.org",
						Enabled:   true,
					},
				},
			},
			expectErr: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 70000,
				},
				Chains: []ChainConfig{
					{
						NetworkID: chains.NetworkBaseSepolia,
						RPCURL:    "https://sepolia.base.org",
						Enabled:   true,
					},
				},
			},
			expectErr: true,
		},
		{
			name: "no enabled chains",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
				Chains: []ChainConfig{
					{
						NetworkID: chains.NetworkBaseSepolia,
						RPCURL:    "https://sepolia.base.org",
						Enabled:   false,
					},
				},
			},
			expectErr: true,
		},
		{
			name: "enabled chain without RPC URL",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
				Chains: []ChainConfig{
					{
						NetworkID: chains.NetworkBaseSepolia,
						RPCURL:    "",
						Enabled:   true,
					},
				},
			},
			expectErr: true,
		},
		{
			name: "enabled chain without network ID",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
				Chains: []ChainConfig{
					{
						NetworkID: "",
						RPCURL:    "https://sepolia.base.org",
						Enabled:   true,
					},
				},
			},
			expectErr: true,
		},
		{
			name: "disabled chain without RPC URL - should be valid",
			config: &Config{
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
						RPCURL:    "",
						Enabled:   false,
					},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

func TestConfigGetEnabledChains(t *testing.T) {
	config := &Config{
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
				Enabled:   false,
			},
			{
				NetworkID: chains.NetworkEthereum,
				RPCURL:    "https://eth.llamarpc.com",
				Enabled:   true,
			},
		},
	}

	enabled := config.GetEnabledChains()

	if len(enabled) != 2 {
		t.Errorf("Expected 2 enabled chains, got %d", len(enabled))
	}

	// Verify only enabled chains are returned
	for _, chain := range enabled {
		if !chain.Enabled {
			t.Errorf("Got disabled chain in enabled chains: %s", chain.NetworkID)
		}
	}

	// Verify correct chains are enabled
	hasBaseSepolia := false
	hasEthereum := false
	hasBase := false

	for _, chain := range enabled {
		if chain.NetworkID == chains.NetworkBaseSepolia {
			hasBaseSepolia = true
		}
		if chain.NetworkID == chains.NetworkEthereum {
			hasEthereum = true
		}
		if chain.NetworkID == chains.NetworkBase {
			hasBase = true
		}
	}

	if !hasBaseSepolia {
		t.Error("Expected Base Sepolia in enabled chains")
	}
	if !hasEthereum {
		t.Error("Expected Ethereum in enabled chains")
	}
	if hasBase {
		t.Error("Did not expect Base in enabled chains (should be disabled)")
	}
}

func TestConfigBuildChainRegistry(t *testing.T) {
	config := &Config{
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

	registry, err := config.BuildChainRegistry()
	if err != nil {
		t.Fatalf("Failed to build chain registry: %v", err)
	}

	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}

	// Verify chains are registered
	if !registry.Has(chains.NetworkBaseSepolia) {
		t.Error("Expected Base Sepolia to be registered")
	}

	if !registry.Has(chains.NetworkBase) {
		t.Error("Expected Base to be registered")
	}

	// Verify chain details
	baseSepoliaChain, err := registry.Get(chains.NetworkBaseSepolia)
	if err != nil {
		t.Fatalf("Failed to get Base Sepolia chain: %v", err)
	}

	if baseSepoliaChain.RPCURL() != "https://sepolia.base.org" {
		t.Errorf("Expected RPC URL https://sepolia.base.org, got %s", baseSepoliaChain.RPCURL())
	}
}

func TestConfigBuildChainRegistryWithDisabledChains(t *testing.T) {
	config := &Config{
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
				Enabled:   false,
			},
		},
	}

	registry, err := config.BuildChainRegistry()
	if err != nil {
		t.Fatalf("Failed to build chain registry: %v", err)
	}

	// Verify only enabled chains are registered
	if !registry.Has(chains.NetworkBaseSepolia) {
		t.Error("Expected Base Sepolia to be registered")
	}

	if registry.Has(chains.NetworkBase) {
		t.Error("Did not expect Base to be registered (should be disabled)")
	}
}

func TestConfigBuildChainRegistryWithUnsupportedNetwork(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Chains: []ChainConfig{
			{
				NetworkID: "unsupported-network",
				RPCURL:    "https://unsupported.rpc.url",
				Enabled:   true,
			},
		},
	}

	_, err := config.BuildChainRegistry()
	if err == nil {
		t.Error("Expected error when building registry with unsupported network")
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	envVars := []string{
		"X402_HOST",
		"X402_PORT",
		"X402_BASE_SEPOLIA_RPC_URL",
		"X402_BASE_SEPOLIA_ENABLED",
		"X402_BASE_RPC_URL",
		"X402_BASE_ENABLED",
	}

	for _, key := range envVars {
		if val, ok := os.LookupEnv(key); ok {
			originalEnv[key] = val
		}
		_ = os.Unsetenv(key)
	}

	// Restore env vars at the end
	defer func() {
		for _, key := range envVars {
			_ = os.Unsetenv(key)
		}
		for key, val := range originalEnv {
			_ = os.Setenv(key, val)
		}
	}()

	// Test with custom env vars
	_ = os.Setenv("X402_HOST", "127.0.0.1")
	_ = os.Setenv("X402_PORT", "9090")
	_ = os.Setenv("X402_BASE_SEPOLIA_RPC_URL", "https://test.sepolia.base.org")
	_ = os.Setenv("X402_BASE_SEPOLIA_ENABLED", "true")
	_ = os.Setenv("X402_BASE_RPC_URL", "https://test.mainnet.base.org")
	_ = os.Setenv("X402_BASE_ENABLED", "false")

	config, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	if config.Server.Host != "127.0.0.1" {
		t.Errorf("Expected host 127.0.0.1, got %s", config.Server.Host)
	}

	if config.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Server.Port)
	}

	if len(config.Chains) != 2 {
		t.Errorf("Expected 2 chains, got %d", len(config.Chains))
	}

	// Check Base Sepolia config
	var baseSepolia *ChainConfig
	for i := range config.Chains {
		if config.Chains[i].NetworkID == chains.NetworkBaseSepolia {
			baseSepolia = &config.Chains[i]
			break
		}
	}

	if baseSepolia == nil {
		t.Fatal("Expected Base Sepolia chain in config")
	}

	if baseSepolia.RPCURL != "https://test.sepolia.base.org" {
		t.Errorf("Expected Base Sepolia RPC URL https://test.sepolia.base.org, got %s", baseSepolia.RPCURL)
	}

	if !baseSepolia.Enabled {
		t.Error("Expected Base Sepolia to be enabled")
	}

	// Check Base config
	var base *ChainConfig
	for i := range config.Chains {
		if config.Chains[i].NetworkID == chains.NetworkBase {
			base = &config.Chains[i]
			break
		}
	}

	if base == nil {
		t.Fatal("Expected Base chain in config")
	}

	if base.Enabled {
		t.Error("Expected Base to be disabled")
	}
}

func TestLoadFromEnvDefaults(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	envVars := []string{
		"X402_HOST",
		"X402_PORT",
		"X402_BASE_SEPOLIA_RPC_URL",
	}

	for _, key := range envVars {
		if val, ok := os.LookupEnv(key); ok {
			originalEnv[key] = val
		}
		_ = os.Unsetenv(key)
	}

	// Restore env vars at the end
	defer func() {
		for _, key := range envVars {
			_ = os.Unsetenv(key)
		}
		for key, val := range originalEnv {
			_ = os.Setenv(key, val)
		}
	}()

	config, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	// Check defaults
	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected default host 0.0.0.0, got %s", config.Server.Host)
	}

	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}
}

func TestGetEnvAsBoolOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true", "true", true},
		{"TRUE", "TRUE", true},
		{"True", "True", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"YES", "YES", true},
		{"false", "false", false},
		{"FALSE", "FALSE", false},
		{"False", "False", false},
		{"0", "0", false},
		{"no", "no", false},
		{"NO", "NO", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEnvAsBoolOrDefault("", true)
			if result != true {
				t.Errorf("Expected default true for empty string, got %v", result)
			}

			_ = os.Setenv("TEST_BOOL", tt.value)
			defer func() { _ = os.Unsetenv("TEST_BOOL") }()

			result = getEnvAsBoolOrDefault("TEST_BOOL", false)
			if result != tt.expected {
				t.Errorf("For value %s, expected %v, got %v", tt.value, tt.expected, result)
			}
		})
	}
}
