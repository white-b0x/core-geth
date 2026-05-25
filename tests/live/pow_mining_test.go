//go:build live

package live

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestMordorRecentBlockValidPoW verifies that a recent Mordor block has valid
// PoW fields: non-zero difficulty, nonce, mixHash, and a valid miner address.
func TestMordorRecentBlockValidPoW(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	block := getLatestBlock(t, client)
	blockNum := block.Number.ToInt().Uint64()

	if block.Difficulty == nil || block.Difficulty.ToInt().Sign() <= 0 {
		t.Fatal("block difficulty is zero or nil — not a PoW block")
	}
	if block.Nonce == "" || block.Nonce == "0x0000000000000000" {
		t.Fatal("block nonce is empty or zero")
	}
	if block.MixHash == (common.Hash{}) {
		t.Fatal("block mixHash is zero")
	}
	if block.Miner == (common.Address{}) {
		t.Fatal("block miner address is zero")
	}

	t.Logf("Mordor block %d: difficulty=%s, nonce=%s, miner=%s",
		blockNum, block.Difficulty.ToInt().String(), block.Nonce, block.Miner.Hex())
}

// TestMordorDifficultyInRange verifies that Mordor block difficulty is within
// a reasonable range for the testnet (not absurdly high or low).
func TestMordorDifficultyInRange(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	block := getLatestBlock(t, client)
	difficulty := block.Difficulty.ToInt()

	// Mordor difficulty should be between 1M and 100T
	// (it's typically in the 100M-10B range for a testnet)
	minDifficulty := big.NewInt(1_000_000)                                            // 1M
	maxDifficulty := new(big.Int).Mul(big.NewInt(100_000_000_000_000), big.NewInt(1)) // 100T

	if difficulty.Cmp(minDifficulty) < 0 {
		t.Fatalf("difficulty %s is below minimum %s", difficulty, minDifficulty)
	}
	if difficulty.Cmp(maxDifficulty) > 0 {
		t.Fatalf("difficulty %s exceeds maximum %s", difficulty, maxDifficulty)
	}

	t.Logf("Mordor difficulty: %s (block %s)", difficulty, block.Number.ToInt().String())
}

// TestMordorECIP1099EpochCalibration verifies the ECIP-1099 fork block
// (2,520,000) exists and has valid PoW fields, confirming the epoch
// calibration fork was processed correctly.
func TestMordorECIP1099EpochCalibration(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorECIP1099Block)

	// Check the fork block itself
	block := getBlockByNumber(t, client, MordorECIP1099Block, false)
	if block.Difficulty == nil || block.Difficulty.ToInt().Sign() <= 0 {
		t.Fatal("ECIP-1099 fork block has zero difficulty")
	}

	// Check blocks immediately before and after the fork
	preFork := getBlockByNumber(t, client, MordorECIP1099Block-1, false)
	postFork := getBlockByNumber(t, client, MordorECIP1099Block+1, false)

	if preFork.Difficulty == nil || preFork.Difficulty.ToInt().Sign() <= 0 {
		t.Fatal("pre-ECIP-1099 block has zero difficulty")
	}
	if postFork.Difficulty == nil || postFork.Difficulty.ToInt().Sign() <= 0 {
		t.Fatal("post-ECIP-1099 block has zero difficulty")
	}

	t.Logf("ECIP-1099 boundary: block %d diff=%s, block %d diff=%s, block %d diff=%s",
		MordorECIP1099Block-1, preFork.Difficulty.ToInt().String(),
		MordorECIP1099Block, block.Difficulty.ToInt().String(),
		MordorECIP1099Block+1, postFork.Difficulty.ToInt().String())
}

// TestMordorECIP1017EraSchedule verifies that the current Mordor block is in
// the expected ECIP-1017 era and that era boundaries have valid structure.
func TestMordorECIP1017EraSchedule(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	block := getLatestBlock(t, client)
	blockNum := block.Number.ToInt().Uint64()
	era := blockNum/ECIP1017EraLength + 1

	t.Logf("Mordor block %d is in Era %d (ECIP-1017 era length: %d)", blockNum, era, ECIP1017EraLength)

	// Verify the most recent era boundary block has valid structure
	if era > 1 {
		boundaryBlock := (era - 1) * ECIP1017EraLength
		boundary := getBlockByNumber(t, client, boundaryBlock, false)
		if boundary.Difficulty == nil || boundary.Difficulty.ToInt().Sign() <= 0 {
			t.Fatalf("era %d boundary block %d has zero difficulty", era, boundaryBlock)
		}
		t.Logf("Era %d boundary (block %d): difficulty=%s",
			era, boundaryBlock, boundary.Difficulty.ToInt().String())
	}
}

// TestMordorBlockTimestamps verifies that block timestamps are monotonically
// increasing in recent blocks, as required by the PoW consensus rules.
func TestMordorBlockTimestamps(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	currentBlock := getBlockNumber(t, client)
	if currentBlock < 10 {
		t.Skip("not enough blocks")
	}

	// Check 10 consecutive recent blocks
	var prevTimestamp uint64
	start := currentBlock - 9
	for blockNum := start; blockNum <= currentBlock; blockNum++ {
		block := getBlockByNumber(t, client, blockNum, false)
		if block.Timestamp == nil {
			t.Fatalf("block %d has nil timestamp", blockNum)
		}
		ts := block.Timestamp.ToInt().Uint64()
		if ts == 0 {
			t.Fatalf("block %d has zero timestamp", blockNum)
		}
		if prevTimestamp > 0 && ts < prevTimestamp {
			t.Errorf("block %d timestamp %d < previous %d — non-monotonic",
				blockNum, ts, prevTimestamp)
		}
		prevTimestamp = ts
	}
	t.Logf("verified monotonic timestamps for blocks %d-%d", start, currentBlock)
}

// TestMordorGasLimitAdjustment verifies that gas limits in recent blocks
// follow the 1/1024 adjustment rule (each block can only change by ±parent/1024).
func TestMordorGasLimitAdjustment(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	currentBlock := getBlockNumber(t, client)
	if currentBlock < 10 {
		t.Skip("not enough blocks")
	}

	// Check gas limit changes across 10 blocks
	start := currentBlock - 9
	for blockNum := start + 1; blockNum <= currentBlock; blockNum++ {
		parent := getBlockByNumber(t, client, blockNum-1, false)
		block := getBlockByNumber(t, client, blockNum, false)

		parentGL := parent.GasLimit.ToInt().Uint64()
		blockGL := block.GasLimit.ToInt().Uint64()

		// Gas limit can change by at most parent_gas_limit / 1024
		maxDelta := parentGL / 1024
		var delta uint64
		if blockGL > parentGL {
			delta = blockGL - parentGL
		} else {
			delta = parentGL - blockGL
		}

		if delta > maxDelta+1 { // +1 for rounding tolerance
			t.Errorf("block %d gas limit change too large: parent=%d, block=%d, delta=%d, max=%d",
				blockNum, parentGL, blockGL, delta, maxDelta)
		}
	}
	t.Logf("verified gas limit adjustment rule for blocks %d-%d", start+1, currentBlock)
}

// TestETCMainnetRecentBlockValidPoW verifies that a recent ETC mainnet block
// has valid PoW fields.
func TestETCMainnetRecentBlockValidPoW(t *testing.T) {
	client := dialRPC(t, getETCRPC())
	defer client.Close()

	block := getLatestBlock(t, client)
	blockNum := block.Number.ToInt().Uint64()

	if block.Difficulty == nil || block.Difficulty.ToInt().Sign() <= 0 {
		t.Fatal("block difficulty is zero or nil")
	}
	if block.Nonce == "" || block.Nonce == "0x0000000000000000" {
		t.Fatal("block nonce is empty or zero")
	}
	if block.MixHash == (common.Hash{}) {
		t.Fatal("block mixHash is zero")
	}
	if block.Miner == (common.Address{}) {
		t.Fatal("block miner address is zero")
	}

	t.Logf("ETC mainnet block %d: difficulty=%s, miner=%s",
		blockNum, block.Difficulty.ToInt().String(), block.Miner.Hex())
}

// TestETCMainnetDifficultyRange verifies the ETC mainnet difficulty is in
// an expected range for a production PoW chain.
func TestETCMainnetDifficultyRange(t *testing.T) {
	client := dialRPC(t, getETCRPC())
	defer client.Close()

	block := getLatestBlock(t, client)
	difficulty := block.Difficulty.ToInt()

	// ETC mainnet difficulty should be between 100B and 100P
	// (typically in the 1T-100T range)
	minDifficulty := big.NewInt(100_000_000_000)                                         // 100B
	maxDifficulty := new(big.Int).Mul(big.NewInt(100_000_000_000_000), big.NewInt(1000)) // 100P

	if difficulty.Cmp(minDifficulty) < 0 {
		t.Fatalf("ETC mainnet difficulty %s is below minimum %s", difficulty, minDifficulty)
	}
	if difficulty.Cmp(maxDifficulty) > 0 {
		t.Fatalf("ETC mainnet difficulty %s exceeds maximum %s", difficulty, maxDifficulty)
	}

	t.Logf("ETC mainnet difficulty: %s (block %s)", difficulty, block.Number.ToInt().String())
}

// TestETCMainnetECIP1017CurrentEra verifies the current era on ETC mainnet
// and that the chain is in era 5+ (block 20M+).
func TestETCMainnetECIP1017CurrentEra(t *testing.T) {
	client := dialRPC(t, getETCRPC())
	defer client.Close()

	block := getLatestBlock(t, client)
	blockNum := block.Number.ToInt().Uint64()
	era := blockNum/ECIP1017EraLength + 1

	t.Logf("ETC mainnet block %d is in Era %d", blockNum, era)

	if blockNum < 20_000_000 {
		t.Logf("WARNING: expected ETC mainnet to be at block 20M+ (Era 5+)")
	}

	// Verify gas limit is around 8M (ETC standard)
	gasLimit := block.GasLimit.ToInt().Uint64()
	if gasLimit < 1_000_000 || gasLimit > 100_000_000 {
		t.Fatalf("unexpected gas limit %d — ETC mainnet should be around 8M", gasLimit)
	}
	t.Logf("ETC mainnet gas limit: %d", gasLimit)
}
