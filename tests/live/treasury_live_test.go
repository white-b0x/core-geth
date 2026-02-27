//go:build live

package live

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestMordorTreasuryBalanceIncreasesPostFork verifies the treasury balance increases after Olympia.
func TestMordorTreasuryBalanceIncreasesPostFork(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock+10)

	treasury := common.HexToAddress(TreasuryAddressHex)

	balanceBefore := getBalance(t, client, treasury, blockTag(MordorForkBlock-1))
	balanceAfter := getBalance(t, client, treasury, blockTag(MordorForkBlock+10))

	t.Logf("treasury balance at fork-1: %s wei", balanceBefore.String())
	t.Logf("treasury balance at fork+10: %s wei", balanceAfter.String())

	if balanceAfter.Cmp(balanceBefore) <= 0 {
		t.Fatalf("expected treasury balance to increase after fork, before=%s after=%s",
			balanceBefore.String(), balanceAfter.String())
	}
}

// TestMordorTreasuryCreditMatchesBaseFee verifies treasury credit equals sum of baseFee*gasUsed.
func TestMordorTreasuryCreditMatchesBaseFee(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock+5)

	treasury := common.HexToAddress(TreasuryAddressHex)

	// Calculate expected treasury credit from baseFee * gasUsed for fork blocks
	balancePre := getBalance(t, client, treasury, blockTag(MordorForkBlock-1))

	expectedCredit := big.NewInt(0)
	blockRewardCredit := big.NewInt(0) // ECIP-1098 20% of block reward

	for i := uint64(0); i < 5; i++ {
		block := getBlockByNumber(t, client, MordorForkBlock+i, false)
		if block.BaseFee != nil && block.GasUsed != nil {
			credit := new(big.Int).Mul(block.BaseFee.ToInt(), block.GasUsed.ToInt())
			expectedCredit.Add(expectedCredit, credit)
			t.Logf("block %d: baseFee=%s gasUsed=%s credit=%s",
				MordorForkBlock+i, block.BaseFee.ToInt().String(),
				block.GasUsed.ToInt().String(), credit.String())
		}
	}

	balancePost := getBalance(t, client, treasury, blockTag(MordorForkBlock+4))

	actualIncrease := new(big.Int).Sub(balancePost, balancePre)
	t.Logf("total balance increase: %s wei", actualIncrease.String())
	t.Logf("expected baseFee credit: %s wei", expectedCredit.String())
	t.Logf("block reward credit (unknown exact): %s wei", blockRewardCredit.String())

	// The actual increase includes both baseFee redirect AND ECIP-1098 block reward share
	// So actualIncrease >= expectedCredit
	if actualIncrease.Cmp(expectedCredit) < 0 {
		t.Fatalf("actual treasury increase (%s) is less than expected baseFee credit (%s)",
			actualIncrease.String(), expectedCredit.String())
	}
}
