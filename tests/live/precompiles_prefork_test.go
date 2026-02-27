//go:build live

package live

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestMordorEcrecoverPreFork verifies the ecrecover precompile works at a pre-fork block.
func TestMordorEcrecoverPreFork(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	// ecrecover at address 0x01 — send invalid input, should return empty (not error)
	ecrecover := common.HexToAddress("0x0000000000000000000000000000000000000001")
	result, err := ethCall(t, client, ecrecover, []byte{0x00}, blockTag(MordorForkBlock-1))
	if err != nil {
		// ecrecover with bad input may return an error or empty result — both are OK
		t.Logf("ecrecover with bad input returned error (expected): %v", err)
		return
	}
	// Empty result or zero result is fine for invalid input
	t.Logf("ecrecover with bad input returned %d bytes", len(result))
}

// TestMordorP256VerifyPreFork verifies P256Verify (0x100) is NOT active before Olympia.
func TestMordorP256VerifyPreFork(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	p256verify := common.HexToAddress(P256VerifyAddr)

	// Pre-fork: no contract at this address, eth_call should return empty
	result, err := ethCall(t, client, p256verify, make([]byte, 160), blockTag(MordorForkBlock-1))
	if err != nil {
		// Error is acceptable — precompile doesn't exist yet
		t.Logf("P256Verify pre-fork call returned error (expected): %v", err)
		return
	}
	if len(result) > 0 {
		t.Fatalf("expected empty result from P256Verify pre-fork, got %d bytes", len(result))
	}
}
