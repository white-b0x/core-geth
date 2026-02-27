//go:build live

package live

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestMordorChainID verifies the node is on the Mordor testnet (chain ID 63).
func TestMordorChainID(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	chainID := getChainID(t, client)
	if chainID != MordorChainID {
		t.Fatalf("expected chain ID %d, got %d", MordorChainID, chainID)
	}
}

// TestMordorGenesisHash verifies the genesis block hash matches the known Mordor genesis.
func TestMordorGenesisHash(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	block := getBlockByNumber(t, client, 0, false)
	if block.Hash != MordorGenesisHash {
		t.Fatalf("expected genesis hash %s, got %s", MordorGenesisHash.Hex(), block.Hash.Hex())
	}
}

// TestMordorRecentBlockStructure verifies the latest block has valid structure.
func TestMordorRecentBlockStructure(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	block := getLatestBlock(t, client)
	if block.Number == nil {
		t.Fatal("latest block number is nil")
	}
	if block.Number.ToInt().Uint64() == 0 {
		t.Fatal("latest block number is 0 — node may not be synced")
	}
	if block.Hash == (MordorGenesisHash) && block.Number.ToInt().Uint64() > 0 {
		t.Fatal("non-genesis block has genesis hash")
	}
	if block.GasLimit == nil || block.GasLimit.ToInt().Sign() <= 0 {
		t.Fatal("block gas limit is zero or nil")
	}
}

// TestMordorPreOlympiaNoBaseFee verifies baseFee is absent in pre-Olympia blocks.
func TestMordorPreOlympiaNoBaseFee(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	// Check a block well before the fork
	block := getBlockByNumber(t, client, MordorForkBlock-1, false)
	if block.BaseFee != nil {
		t.Fatalf("expected no baseFee at block %d (pre-Olympia), got %s",
			MordorForkBlock-1, block.BaseFee.ToInt().String())
	}
}

// TestETCMainnetChainID verifies the node is on ETC mainnet (chain ID 61).
func TestETCMainnetChainID(t *testing.T) {
	client := dialRPC(t, getETCRPC())
	defer client.Close()

	chainID := getChainID(t, client)
	if chainID != ETCMainnetChainID {
		t.Fatalf("expected chain ID %d, got %d", ETCMainnetChainID, chainID)
	}
}

// TestETCMainnetGenesisHash verifies the genesis block hash matches known ETC genesis.
func TestETCMainnetGenesisHash(t *testing.T) {
	client := dialRPC(t, getETCRPC())
	defer client.Close()

	block := getBlockByNumber(t, client, 0, false)
	if block.Hash != ETCGenesisHash {
		t.Fatalf("expected genesis hash %s, got %s", ETCGenesisHash.Hex(), block.Hash.Hex())
	}
}

// TestETCMainnetRecentNoBaseFee verifies ETC mainnet blocks don't have baseFee (pre-Olympia).
func TestETCMainnetRecentNoBaseFee(t *testing.T) {
	client := dialRPC(t, getETCRPC())
	defer client.Close()

	block := getLatestBlock(t, client)
	if block.BaseFee != nil {
		t.Fatalf("expected no baseFee on ETC mainnet (pre-Olympia), got %s at block %s",
			block.BaseFee.ToInt().String(), block.Number.ToInt().String())
	}
}

// TestETCMainnetTreasuryBalance verifies the treasury address balance exists on mainnet.
func TestETCMainnetTreasuryBalance(t *testing.T) {
	client := dialRPC(t, getETCRPC())
	defer client.Close()

	treasury := common.HexToAddress(TreasuryAddressHex)
	balance := getBalance(t, client, treasury, "latest")
	// Pre-Olympia: treasury may or may not have a balance depending on whether
	// it's been used. Just verify the query works without error.
	t.Logf("ETC mainnet treasury balance: %s wei", balance.String())
}
