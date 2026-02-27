//go:build live

package live

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestMordorPreOlympiaNoType2Txs verifies no Type-2 or Type-4 transactions exist pre-Olympia.
func TestMordorPreOlympiaNoType2Txs(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	// Check several recent pre-fork blocks
	checkBlock := uint64(MordorForkBlock - 100)
	for i := uint64(0); i < 10; i++ {
		block := getBlockByNumber(t, client, checkBlock+i, true)
		for _, rawTx := range block.Transactions {
			var tx rpcTx
			if err := json.Unmarshal(rawTx, &tx); err != nil {
				t.Fatalf("failed to unmarshal tx in block %d: %v", checkBlock+i, err)
			}
			txType := uint64(tx.Type)
			if txType == 2 {
				t.Errorf("found Type-2 (EIP-1559) tx %s in pre-Olympia block %d",
					tx.Hash.Hex(), checkBlock+i)
			}
			if txType == 4 {
				t.Errorf("found Type-4 (EIP-7702) tx %s in pre-Olympia block %d",
					tx.Hash.Hex(), checkBlock+i)
			}
		}
	}
}

// TestMordorLegacyTxDecodes verifies legacy transactions in recent blocks decode correctly.
func TestMordorLegacyTxDecodes(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	// Find a block with transactions by scanning recent blocks
	currentBlock := getBlockNumber(t, client)
	if currentBlock < 100 {
		t.Skip("not enough blocks to scan")
	}

	for blockNum := currentBlock; blockNum > currentBlock-200 && blockNum > 0; blockNum-- {
		block := getBlockByNumber(t, client, blockNum, true)
		if len(block.Transactions) == 0 {
			continue
		}

		// Found a block with transactions
		var tx rpcTx
		if err := json.Unmarshal(block.Transactions[0], &tx); err != nil {
			t.Fatalf("failed to unmarshal tx: %v", err)
		}

		// Legacy or EIP-2930 transaction
		txType := uint64(tx.Type)
		if txType > 1 {
			// Skip non-legacy for this test
			continue
		}

		if tx.Hash == (common.Hash{}) {
			t.Fatal("transaction hash is zero")
		}
		if tx.GasPrice == nil {
			t.Fatal("legacy tx gasPrice is nil")
		}
		if tx.Gas == nil || tx.Gas.ToInt().Sign() <= 0 {
			t.Fatal("legacy tx gas is zero or nil")
		}

		t.Logf("verified legacy tx %s in block %d (type=%d, gas=%s)",
			tx.Hash.Hex(), blockNum, txType, tx.Gas.ToInt().String())
		return
	}

	t.Log("no transactions found in recent 200 blocks — Mordor may be quiet")
}
