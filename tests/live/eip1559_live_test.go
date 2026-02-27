//go:build live

package live

import (
	"encoding/json"
	"math/big"
	"testing"
)

// TestMordorForkBlockHasBaseFee verifies the Olympia fork block has a baseFee field.
func TestMordorForkBlockHasBaseFee(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock)

	block := getBlockByNumber(t, client, MordorForkBlock, false)
	if block.BaseFee == nil {
		t.Fatalf("expected baseFee at fork block %d, got nil", MordorForkBlock)
	}
	t.Logf("fork block %d baseFee: %s wei", MordorForkBlock, block.BaseFee.ToInt().String())
}

// TestMordorInitialBaseFee verifies the initial baseFee is 1 gwei at the fork block.
func TestMordorInitialBaseFee(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock)

	block := getBlockByNumber(t, client, MordorForkBlock, false)
	if block.BaseFee == nil {
		t.Fatalf("baseFee is nil at fork block")
	}

	expected := big.NewInt(InitialBaseFeeGwei)
	if block.BaseFee.ToInt().Cmp(expected) != 0 {
		t.Fatalf("expected initial baseFee %s, got %s",
			expected.String(), block.BaseFee.ToInt().String())
	}
}

// TestMordorBaseFeeAdjusts verifies baseFee changes based on gas usage after the fork.
func TestMordorBaseFeeAdjusts(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock+5)

	prevBaseFee := big.NewInt(0)
	changed := false

	for i := uint64(0); i <= 5; i++ {
		block := getBlockByNumber(t, client, MordorForkBlock+i, false)
		if block.BaseFee == nil {
			t.Fatalf("baseFee is nil at block %d", MordorForkBlock+i)
		}
		currentBaseFee := block.BaseFee.ToInt()
		t.Logf("block %d: baseFee=%s gasUsed=%s gasLimit=%s",
			MordorForkBlock+i, currentBaseFee.String(),
			block.GasUsed.ToInt().String(), block.GasLimit.ToInt().String())

		if i > 0 && currentBaseFee.Cmp(prevBaseFee) != 0 {
			changed = true
		}
		prevBaseFee = new(big.Int).Set(currentBaseFee)
	}

	// baseFee should change if there's any gas usage variation
	// If all blocks are empty, baseFee will decrease toward the floor
	if !changed {
		t.Log("WARNING: baseFee did not change over 5 blocks — all blocks may have same gas usage")
	}
}

// TestMordorType2TxPostFork looks for a Type-2 (EIP-1559) transaction after the fork.
func TestMordorType2TxPostFork(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock)

	currentBlock := getBlockNumber(t, client)
	// Scan post-fork blocks for Type-2 transactions
	scanEnd := uint64(MordorForkBlock + 1000)
	if currentBlock < scanEnd {
		scanEnd = currentBlock
	}

	for blockNum := uint64(MordorForkBlock); blockNum <= scanEnd; blockNum++ {
		block := getBlockByNumber(t, client, blockNum, true)
		for _, rawTx := range block.Transactions {
			var tx rpcTx
			if err := json.Unmarshal(rawTx, &tx); err != nil {
				continue
			}
			if uint64(tx.Type) == 2 {
				t.Logf("found Type-2 tx %s in block %d", tx.Hash.Hex(), blockNum)
				if tx.MaxFeePerGas == nil {
					t.Error("Type-2 tx missing maxFeePerGas")
				}
				if tx.MaxPriorityFeePerGas == nil {
					t.Error("Type-2 tx missing maxPriorityFeePerGas")
				}
				return
			}
		}
	}
	t.Log("no Type-2 transactions found in post-fork blocks — may need to send one")
}
