package chains

import (
	"fmt"
	"math/big"
)

// Chain represents a blockchain network configuration
type Chain interface {
	// Name returns the human-readable name of the chain
	Name() string

	// NetworkID returns the network identifier (e.g., "base-sepolia", "base")
	NetworkID() string

	// ChainID returns the numeric chain ID for EVM networks
	ChainID() *big.Int

	// RPCURL returns the RPC endpoint URL for the chain
	RPCURL() string
}

// chain is the concrete implementation of the Chain interface
type chain struct {
	name      string
	networkID string
	chainID   *big.Int
	rpcURL    string
}

func (c *chain) Name() string {
	return c.name
}

func (c *chain) NetworkID() string {
	return c.networkID
}

func (c *chain) ChainID() *big.Int {
	return c.chainID
}

func (c *chain) RPCURL() string {
	return c.rpcURL
}

// NewChain creates a new Chain instance with the provided configuration
func NewChain(name, networkID string, chainID *big.Int, rpcURL string) Chain {
	return &chain{
		name:      name,
		networkID: networkID,
		chainID:   chainID,
		rpcURL:    rpcURL,
	}
}

// Predefined chain IDs
var (
	// Base Sepolia testnet
	ChainIDBaseSepolia = big.NewInt(84532)

	// Base mainnet
	ChainIDBase = big.NewInt(8453)

	// Ethereum mainnet
	ChainIDEthereum = big.NewInt(1)

	// Ethereum Sepolia testnet
	ChainIDEthereumSepolia = big.NewInt(11155111)

	// Ethereum Holesky testnet
	ChainIDEthereumHolesky = big.NewInt(17000)
)

// Network identifier constants
const (
	NetworkBaseSepolia     = "base-sepolia"
	NetworkBase            = "base"
	NetworkEthereum        = "ethereum"
	NetworkEthereumSepolia = "ethereum-sepolia"
	NetworkEthereumHolesky = "ethereum-holesky"
)

// NewBaseSepolia creates a Base Sepolia testnet chain configuration
func NewBaseSepolia(rpcURL string) Chain {
	return NewChain("Base Sepolia", NetworkBaseSepolia, ChainIDBaseSepolia, rpcURL)
}

// NewBase creates a Base mainnet chain configuration
func NewBase(rpcURL string) Chain {
	return NewChain("Base", NetworkBase, ChainIDBase, rpcURL)
}

// NewEthereum creates an Ethereum mainnet chain configuration
func NewEthereum(rpcURL string) Chain {
	return NewChain("Ethereum", NetworkEthereum, ChainIDEthereum, rpcURL)
}

// NewEthereumSepolia creates an Ethereum Sepolia testnet chain configuration
func NewEthereumSepolia(rpcURL string) Chain {
	return NewChain("Ethereum Sepolia", NetworkEthereumSepolia, ChainIDEthereumSepolia, rpcURL)
}

// NewEthereumHolesky creates an Ethereum Holesky testnet chain configuration
func NewEthereumHolesky(rpcURL string) Chain {
	return NewChain("Ethereum Holesky", NetworkEthereumHolesky, ChainIDEthereumHolesky, rpcURL)
}

// Registry manages a collection of chain configurations
type Registry struct {
	chains map[string]Chain
}

// NewRegistry creates a new empty chain registry
func NewRegistry() *Registry {
	return &Registry{
		chains: make(map[string]Chain),
	}
}

// Register adds a chain to the registry
func (r *Registry) Register(chain Chain) {
	r.chains[chain.NetworkID()] = chain
}

// Get retrieves a chain by its network ID
func (r *Registry) Get(networkID string) (Chain, error) {
	chain, ok := r.chains[networkID]
	if !ok {
		return nil, fmt.Errorf("chain not found for network ID: %s", networkID)
	}
	return chain, nil
}

// Has checks if a chain is registered for the given network ID
func (r *Registry) Has(networkID string) bool {
	_, ok := r.chains[networkID]
	return ok
}

// List returns all registered network IDs
func (r *Registry) List() []string {
	networkIDs := make([]string, 0, len(r.chains))
	for networkID := range r.chains {
		networkIDs = append(networkIDs, networkID)
	}
	return networkIDs
}

// GetAll returns all registered chains
func (r *Registry) GetAll() []Chain {
	chains := make([]Chain, 0, len(r.chains))
	for _, chain := range r.chains {
		chains = append(chains, chain)
	}
	return chains
}
