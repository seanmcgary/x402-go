package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	// USDC token address on Ethereum Sepolia
	USDCTokenAddress = "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238"
)

func GetProjectRootPath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	startingPath := ""
	iterations := 0
	for {
		if iterations > 10 {
			panic("Could not find project root path")
		}
		iterations++
		p, err := filepath.Abs(fmt.Sprintf("%s/%s", wd, startingPath))
		if err != nil {
			panic(err)
		}

		match := regexp.MustCompile(`\/x402-go([A-Za-z0-9_-]+)?\/?$`)
		if match.MatchString(p) {
			fmt.Printf("Found project root path: %s\n", p)
			return p
		}
		startingPath = startingPath + "/.."
	}
}

// TestAccount represents a test account with address and private key
type TestAccount struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

// ChainConfig contains test chain configuration and account data
type ChainConfig struct {
	ForkL1Block      string      `json:"forkL1Block"`
	PayerAccount     TestAccount `json:"payerAccount"`
	RecipientAccount TestAccount `json:"recipientAccount"`
}

func ReadChainConfig(projectRoot string) (*ChainConfig, error) {
	filePath := fmt.Sprintf("%s/internal/testData/chain-config.json", projectRoot)

	// read the file into bytes
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cf *ChainConfig
	if err := json.Unmarshal(file, &cf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}
	return cf, nil
}
