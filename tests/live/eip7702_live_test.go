//go:build live

package live

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestMordorType4TxPostFork scans for Type-4 (EIP-7702) transactions after the fork.
func TestMordorType4TxPostFork(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock)

	currentBlock := getBlockNumber(t, client)
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
			if uint64(tx.Type) == 4 {
				t.Logf("found Type-4 tx %s in block %d", tx.Hash.Hex(), blockNum)
				if tx.AuthorizationList == nil || string(tx.AuthorizationList) == "null" {
					t.Error("Type-4 tx missing authorizationList")
				}
				return
			}
		}
	}
	t.Log("no Type-4 transactions found — may need to send one for full coverage")
}

// TestMordorDelegationCode verifies delegation code prefix at delegated addresses.
func TestMordorDelegationCode(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock)

	// Scan for Type-4 transactions to find delegated addresses
	currentBlock := getBlockNumber(t, client)
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
			if uint64(tx.Type) == 4 {
				// Parse authorization list to find delegated addresses
				type authTuple struct {
					ChainId *string        `json:"chainId"`
					Address common.Address `json:"address"`
				}
				var authList []authTuple
				if err := json.Unmarshal(tx.AuthorizationList, &authList); err != nil {
					t.Logf("failed to parse authorizationList: %v", err)
					continue
				}

				for _, auth := range authList {
					code := getCode(t, client, auth.Address, "latest")
					if len(code) >= 3 && code[0] == 0xef && code[1] == 0x01 && code[2] == 0x00 {
						t.Logf("verified delegation prefix 0xef0100 at %s (code length: %d)",
							auth.Address.Hex(), len(code))
						return
					}
				}
			}
		}
	}
	t.Log("no EIP-7702 delegated addresses found — may need to send a Type-4 tx")
}
