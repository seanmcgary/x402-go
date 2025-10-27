package chains

import (
	"math/big"
	"testing"
)

func TestNewChain(t *testing.T) {
	name := "Test Chain"
	networkID := "test-network"
	chainID := big.NewInt(12345)
	rpcURL := "https://test.rpc.url"

	chain := NewChain(name, networkID, chainID, rpcURL)

	if chain.Name() != name {
		t.Errorf("Expected name %s, got %s", name, chain.Name())
	}
	if chain.NetworkID() != networkID {
		t.Errorf("Expected networkID %s, got %s", networkID, chain.NetworkID())
	}
	if chain.ChainID().Cmp(chainID) != 0 {
		t.Errorf("Expected chainID %s, got %s", chainID.String(), chain.ChainID().String())
	}
	if chain.RPCURL() != rpcURL {
		t.Errorf("Expected rpcURL %s, got %s", rpcURL, chain.RPCURL())
	}
}

func TestNewBaseSepolia(t *testing.T) {
	rpcURL := "https://sepolia.base.org"
	chain := NewBaseSepolia(rpcURL)

	if chain.Name() != "Base Sepolia" {
		t.Errorf("Expected name 'Base Sepolia', got %s", chain.Name())
	}
	if chain.NetworkID() != NetworkBaseSepolia {
		t.Errorf("Expected networkID %s, got %s", NetworkBaseSepolia, chain.NetworkID())
	}
	if chain.ChainID().Cmp(ChainIDBaseSepolia) != 0 {
		t.Errorf("Expected chainID %s, got %s", ChainIDBaseSepolia.String(), chain.ChainID().String())
	}
	if chain.RPCURL() != rpcURL {
		t.Errorf("Expected rpcURL %s, got %s", rpcURL, chain.RPCURL())
	}
}

func TestNewBase(t *testing.T) {
	rpcURL := "https://mainnet.base.org"
	chain := NewBase(rpcURL)

	if chain.Name() != "Base" {
		t.Errorf("Expected name 'Base', got %s", chain.Name())
	}
	if chain.NetworkID() != NetworkBase {
		t.Errorf("Expected networkID %s, got %s", NetworkBase, chain.NetworkID())
	}
	if chain.ChainID().Cmp(ChainIDBase) != 0 {
		t.Errorf("Expected chainID %s, got %s", ChainIDBase.String(), chain.ChainID().String())
	}
	if chain.RPCURL() != rpcURL {
		t.Errorf("Expected rpcURL %s, got %s", rpcURL, chain.RPCURL())
	}
}

func TestNewEthereumHolesky(t *testing.T) {
	rpcURL := "https://ethereum-holesky.publicnode.com"
	chain := NewEthereumHolesky(rpcURL)

	if chain.Name() != "Ethereum Holesky" {
		t.Errorf("Expected name 'Ethereum Holesky', got %s", chain.Name())
	}
	if chain.NetworkID() != NetworkEthereumHolesky {
		t.Errorf("Expected networkID %s, got %s", NetworkEthereumHolesky, chain.NetworkID())
	}
	if chain.ChainID().Cmp(ChainIDEthereumHolesky) != 0 {
		t.Errorf("Expected chainID %s, got %s", ChainIDEthereumHolesky.String(), chain.ChainID().String())
	}
	if chain.RPCURL() != rpcURL {
		t.Errorf("Expected rpcURL %s, got %s", rpcURL, chain.RPCURL())
	}
}

func TestNewEthereum(t *testing.T) {
	rpcURL := "https://eth.llamarpc.com"
	chain := NewEthereum(rpcURL)

	if chain.Name() != "Ethereum" {
		t.Errorf("Expected name 'Ethereum', got %s", chain.Name())
	}
	if chain.NetworkID() != NetworkEthereum {
		t.Errorf("Expected networkID %s, got %s", NetworkEthereum, chain.NetworkID())
	}
	if chain.ChainID().Cmp(ChainIDEthereum) != 0 {
		t.Errorf("Expected chainID %s, got %s", ChainIDEthereum.String(), chain.ChainID().String())
	}
	if chain.RPCURL() != rpcURL {
		t.Errorf("Expected rpcURL %s, got %s", rpcURL, chain.RPCURL())
	}
}

func TestNewEthereumSepolia(t *testing.T) {
	rpcURL := "https://eth-sepolia.public.blastapi.io"
	chain := NewEthereumSepolia(rpcURL)

	if chain.Name() != "Ethereum Sepolia" {
		t.Errorf("Expected name 'Ethereum Sepolia', got %s", chain.Name())
	}
	if chain.NetworkID() != NetworkEthereumSepolia {
		t.Errorf("Expected networkID %s, got %s", NetworkEthereumSepolia, chain.NetworkID())
	}
	if chain.ChainID().Cmp(ChainIDEthereumSepolia) != 0 {
		t.Errorf("Expected chainID %s, got %s", ChainIDEthereumSepolia.String(), chain.ChainID().String())
	}
	if chain.RPCURL() != rpcURL {
		t.Errorf("Expected rpcURL %s, got %s", rpcURL, chain.RPCURL())
	}
}

func TestChainIDConstants(t *testing.T) {
	tests := []struct {
		name     string
		chainID  *big.Int
		expected int64
	}{
		{"Base Sepolia", ChainIDBaseSepolia, 84532},
		{"Base", ChainIDBase, 8453},
		{"Ethereum", ChainIDEthereum, 1},
		{"Ethereum Sepolia", ChainIDEthereumSepolia, 11155111},
		{"Ethereum Holesky", ChainIDEthereumHolesky, 17000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.chainID.Int64() != tt.expected {
				t.Errorf("Expected chainID %d for %s, got %d", tt.expected, tt.name, tt.chainID.Int64())
			}
		})
	}
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}

	if len(registry.List()) != 0 {
		t.Errorf("Expected empty registry, got %d chains", len(registry.List()))
	}
}

func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry()
	chain := NewBaseSepolia("https://sepolia.base.org")

	registry.Register(chain)

	if !registry.Has(NetworkBaseSepolia) {
		t.Error("Expected registry to have base-sepolia chain")
	}

	retrieved, err := registry.Get(NetworkBaseSepolia)
	if err != nil {
		t.Fatalf("Failed to get chain: %v", err)
	}

	if retrieved.NetworkID() != chain.NetworkID() {
		t.Errorf("Expected networkID %s, got %s", chain.NetworkID(), retrieved.NetworkID())
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("non-existent-network")
	if err == nil {
		t.Error("Expected error when getting non-existent chain")
	}
}

func TestRegistryHas(t *testing.T) {
	registry := NewRegistry()
	chain := NewBase("https://mainnet.base.org")

	if registry.Has(NetworkBase) {
		t.Error("Expected registry to not have base chain before registration")
	}

	registry.Register(chain)

	if !registry.Has(NetworkBase) {
		t.Error("Expected registry to have base chain after registration")
	}
}

func TestRegistryList(t *testing.T) {
	registry := NewRegistry()

	chain1 := NewBaseSepolia("https://sepolia.base.org")
	chain2 := NewBase("https://mainnet.base.org")

	registry.Register(chain1)
	registry.Register(chain2)

	networkIDs := registry.List()

	if len(networkIDs) != 2 {
		t.Errorf("Expected 2 network IDs, got %d", len(networkIDs))
	}

	// Check that both network IDs are present
	hasBaseSepolia := false
	hasBase := false
	for _, id := range networkIDs {
		if id == NetworkBaseSepolia {
			hasBaseSepolia = true
		}
		if id == NetworkBase {
			hasBase = true
		}
	}

	if !hasBaseSepolia {
		t.Error("Expected base-sepolia in network IDs list")
	}
	if !hasBase {
		t.Error("Expected base in network IDs list")
	}
}

func TestRegistryGetAll(t *testing.T) {
	registry := NewRegistry()

	chain1 := NewBaseSepolia("https://sepolia.base.org")
	chain2 := NewBase("https://mainnet.base.org")

	registry.Register(chain1)
	registry.Register(chain2)

	allChains := registry.GetAll()

	if len(allChains) != 2 {
		t.Errorf("Expected 2 chains, got %d", len(allChains))
	}

	// Check that both chains are present
	hasBaseSepolia := false
	hasBase := false
	for _, chain := range allChains {
		if chain.NetworkID() == NetworkBaseSepolia {
			hasBaseSepolia = true
		}
		if chain.NetworkID() == NetworkBase {
			hasBase = true
		}
	}

	if !hasBaseSepolia {
		t.Error("Expected base-sepolia chain in GetAll result")
	}
	if !hasBase {
		t.Error("Expected base chain in GetAll result")
	}
}

func TestRegistryMultipleRegistrations(t *testing.T) {
	registry := NewRegistry()

	// Register multiple chains
	chains := []Chain{
		NewBaseSepolia("https://sepolia.base.org"),
		NewBase("https://mainnet.base.org"),
		NewEthereum("https://eth.llamarpc.com"),
		NewEthereumSepolia("https://eth-sepolia.public.blastapi.io"),
		NewEthereumHolesky("https://ethereum-holesky.publicnode.com"),
	}

	for _, chain := range chains {
		registry.Register(chain)
	}

	if len(registry.List()) != 5 {
		t.Errorf("Expected 5 chains registered, got %d", len(registry.List()))
	}

	// Verify each chain can be retrieved
	for _, chain := range chains {
		retrieved, err := registry.Get(chain.NetworkID())
		if err != nil {
			t.Errorf("Failed to get chain %s: %v", chain.NetworkID(), err)
		}
		if retrieved.NetworkID() != chain.NetworkID() {
			t.Errorf("Expected networkID %s, got %s", chain.NetworkID(), retrieved.NetworkID())
		}
	}
}
