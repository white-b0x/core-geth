//go:build live

package live

import (
	"testing"
)

// TestMordorPostForkBlocksUnderSizeCap verifies post-fork blocks are under the 8MB RLP size cap.
func TestMordorPostForkBlocksUnderSizeCap(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock+5)

	for i := uint64(0); i < 5; i++ {
		block := getBlockByNumber(t, client, MordorForkBlock+i, false)
		if block.Size != nil {
			size := block.Size.ToInt().Uint64()
			if size > BlockRLPSizeCap {
				t.Fatalf("block %d size %d exceeds EIP-7934 cap %d",
					MordorForkBlock+i, size, BlockRLPSizeCap)
			}
			t.Logf("block %d size: %d bytes (cap: %d)", MordorForkBlock+i, size, BlockRLPSizeCap)
		}
	}
}
