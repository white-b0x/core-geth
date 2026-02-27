//go:build live

package live

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestMordorTreasuryPreForkBalance snapshots the treasury balance before the Olympia fork.
func TestMordorTreasuryPreForkBalance(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	treasury := common.HexToAddress(TreasuryAddressHex)

	// Get balance at the block just before the fork
	balance := getBalance(t, client, treasury, blockTag(MordorForkBlock-1))
	t.Logf("treasury balance at block %d (pre-fork): %s wei", MordorForkBlock-1, balance.String())
	// This is a snapshot test — we record the value. Post-fork tests will verify it increases.
}
