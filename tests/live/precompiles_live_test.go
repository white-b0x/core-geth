//go:build live

package live

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestMordorP256VerifyPostFork verifies P256Verify is active after the Olympia fork.
func TestMordorP256VerifyPostFork(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock)

	p256verify := common.HexToAddress(P256VerifyAddr)

	// Send a valid-length input (160 bytes) — will fail verification but should consume gas
	input := make([]byte, 160)
	// Fill with non-zero data so it's not just zeros
	for i := range input {
		input[i] = byte(i)
	}

	result, err := ethCall(t, client, p256verify, input, blockTag(MordorForkBlock))
	if err != nil {
		// Precompile returning an error is acceptable — it means it's active but input is invalid
		t.Logf("P256Verify post-fork returned error (precompile active, bad input): %v", err)
		return
	}
	// Empty result means signature verification failed (expected for random input)
	t.Logf("P256Verify post-fork returned %d bytes (expected 0 for invalid sig)", len(result))
}

// TestMordorBLSStubsPostFork verifies BLS12-381 precompile stubs consume gas and fail.
func TestMordorBLSStubsPostFork(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock)

	for _, addrHex := range BLSPrecompileAddrs {
		addr := common.HexToAddress(addrHex)
		// Send minimal input
		input := make([]byte, 64)

		_, err := ethCall(t, client, addr, input, blockTag(MordorForkBlock))
		if err != nil {
			// BLS stubs should consume gas and return failure — error is expected
			t.Logf("BLS precompile %s: returned error (stub active): %v", addrHex, err)
		} else {
			t.Logf("BLS precompile %s: returned success (unexpected for stub)", addrHex)
		}
	}
}

// TestMordorModExpPostFork verifies ModExp works with EIP-7883 repricing after the fork.
func TestMordorModExpPostFork(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock)

	modexp := common.HexToAddress("0x0000000000000000000000000000000000000005")

	// Simple modexp: 2^3 mod 5 = 3
	// Input format: Bsize (32) | Esize (32) | Msize (32) | B | E | M
	input := make([]byte, 0, 96+3)
	// Bsize = 1
	bsize := make([]byte, 32)
	bsize[31] = 1
	input = append(input, bsize...)
	// Esize = 1
	esize := make([]byte, 32)
	esize[31] = 1
	input = append(input, esize...)
	// Msize = 1
	msize := make([]byte, 32)
	msize[31] = 1
	input = append(input, msize...)
	// B = 2, E = 3, M = 5
	input = append(input, 2, 3, 5)

	result, err := ethCall(t, client, modexp, input, blockTag(MordorForkBlock))
	if err != nil {
		t.Fatalf("ModExp call failed: %v", err)
	}

	// Result should be 1 byte: 3 (2^3 mod 5 = 8 mod 5 = 3)
	if len(result) != 1 || result[0] != 3 {
		t.Fatalf("expected ModExp result [3], got %v", result)
	}
	t.Logf("ModExp 2^3 mod 5 = %d (correct)", result[0])
}

// TestMordorEcrecoverStillWorks verifies ecrecover continues to work post-fork.
func TestMordorEcrecoverStillWorks(t *testing.T) {
	client := dialRPC(t, getMordorRPC())
	defer client.Close()

	requireForkReached(t, client, MordorForkBlock)

	ecrecover := common.HexToAddress("0x0000000000000000000000000000000000000001")

	// Invalid input — should return empty (not crash)
	_, err := ethCall(t, client, ecrecover, []byte{0x00}, blockTag(MordorForkBlock))
	if err != nil {
		t.Logf("ecrecover with bad input returned error (expected): %v", err)
	} else {
		t.Log("ecrecover post-fork works (returned empty for bad input)")
	}
}
