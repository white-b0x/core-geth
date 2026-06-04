// Copyright 2026 The core-geth Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package eip1559

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params/types/coregeth"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/vars"
)

// newETCTestConfig returns a minimal ETC-style chain config with Olympia at
// the given block number. spiralBlock is olympiaBlock/2 (arbitrary pre-fork
// point to exercise the Spiral-era gas schedule path).
func newETCTestConfig(olympiaBlock uint64) *coregeth.CoreGethChainConfig {
	spiralBlock := olympiaBlock / 2
	if spiralBlock == 0 {
		spiralBlock = 1
	}
	gasTarget8M := uint64(8_000_000)
	gasTarget60M := uint64(60_000_000)
	return &coregeth.CoreGethChainConfig{
		Ethash:           new(ctypes.EthashConfig),
		EIP3855FBlock:    big.NewInt(int64(spiralBlock)),  // Spiral (PUSH0)
		EIP1559FBlock:    big.NewInt(int64(olympiaBlock)), // Olympia
		EIP3198FBlock:    big.NewInt(int64(olympiaBlock)), // BASEFEE opcode
		SpiralGasTarget:  &gasTarget8M,
		OlympiaGasTarget: &gasTarget60M,
		BaseFeeMinValue:  big.NewInt(int64(vars.InitialBaseFee)), // ECIP-1111: 1 gwei floor
	}
}

// newETHTestConfig returns a chain config without a baseFee floor (ETH mainnet behaviour).
func newETHTestConfig(eip1559Block uint64) *coregeth.CoreGethChainConfig {
	return &coregeth.CoreGethChainConfig{
		Ethash:        new(ctypes.EthashConfig),
		EIP1559FBlock: big.NewInt(int64(eip1559Block)),
		EIP3198FBlock: big.NewInt(int64(eip1559Block)),
		// BaseFeeMinValue intentionally nil — no floor on ETH
	}
}

// preOlympiaHeader returns a header that is one block before Olympia (no BaseFee).
func preOlympiaHeader(blockNum, gasLimit, gasUsed uint64) *types.Header {
	return &types.Header{
		Number:   big.NewInt(int64(blockNum)),
		GasLimit: gasLimit,
		GasUsed:  gasUsed,
	}
}

// olympiaHeader returns a header at or after Olympia activation (has BaseFee).
func olympiaHeader(blockNum, gasLimit, gasUsed uint64, baseFee *big.Int) *types.Header {
	return &types.Header{
		Number:   big.NewInt(int64(blockNum)),
		GasLimit: gasLimit,
		GasUsed:  gasUsed,
		BaseFee:  baseFee,
	}
}

// TestInitialBaseFeeIs1Gwei verifies the InitialBaseFee constant equals 1 Gwei.
func TestInitialBaseFeeIs1Gwei(t *testing.T) {
	want := uint64(1_000_000_000)
	if vars.InitialBaseFee != want {
		t.Fatalf("InitialBaseFee = %d, want %d (1 Gwei)", vars.InitialBaseFee, want)
	}
}

// TestBaseFeePreOlympia_GasZero verifies that CalcBaseFee returns InitialBaseFee
// for the first Olympia block regardless of gas usage — fee market suppressed pre-Olympia.
func TestBaseFeePreOlympia_GasZero(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	// Parent is block 99 (pre-Olympia, no baseFee), gasUsed = 0
	parent := preOlympiaHeader(olympiaBlock-1, 8_000_000, 0)
	got := CalcBaseFee(cfg, parent)
	want := new(big.Int).SetUint64(vars.InitialBaseFee)
	if got.Cmp(want) != 0 {
		t.Fatalf("CalcBaseFee (first Olympia, gasUsed=0) = %s, want %s", got, want)
	}
}

// TestBaseFeePreOlympia_FullyUsed verifies that CalcBaseFee returns InitialBaseFee
// even when the parent was fully utilised — fee market is suppressed pre-Olympia.
func TestBaseFeePreOlympia_FullyUsed(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	parent := preOlympiaHeader(olympiaBlock-1, 8_000_000, 8_000_000)
	got := CalcBaseFee(cfg, parent)
	want := new(big.Int).SetUint64(vars.InitialBaseFee)
	if got.Cmp(want) != 0 {
		t.Fatalf("CalcBaseFee (first Olympia, fully used parent) = %s, want %s", got, want)
	}
}

// TestBaseFeeFirstOlympiaBlock verifies that when the parent is the first Olympia
// block (baseFee=InitialBaseFee, gasUsed=0), the computed decrease is clamped to
// InitialBaseFee by the ECIP-1111 floor. The raw decrease would be 875_000_000
// (InitialBaseFee - InitialBaseFee/8), but the 1 gwei floor prevents it.
func TestBaseFeeFirstOlympiaBlock(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	// Parent is the first Olympia block: gasLimit=30M, gasUsed=0, baseFee=1 Gwei.
	parent := olympiaHeader(olympiaBlock, 30_000_000, 0, new(big.Int).SetUint64(vars.InitialBaseFee))
	got := CalcBaseFee(cfg, parent)
	// gasTarget = 30M/2 = 15M; gasUsed(0) < gasTarget(15M) → computed decrease to 875_000_000
	// ECIP-1111 floor clamps result back to InitialBaseFee (1 gwei).
	want := new(big.Int).SetUint64(vars.InitialBaseFee)
	if got.Cmp(want) != 0 {
		t.Fatalf("CalcBaseFee (ETC, second Olympia, empty) = %s, want %s (ECIP-1111 floor)", got, want)
	}
}

// TestBaseFeeStable verifies that baseFee is unchanged when gasUsed == gasTarget.
func TestBaseFeeStable(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	gasLimit := uint64(30_000_000)
	gasTarget := gasLimit / 2 // = 15_000_000
	parent := olympiaHeader(olympiaBlock, gasLimit, gasTarget, big.NewInt(1_000_000_000))
	got := CalcBaseFee(cfg, parent)
	want := big.NewInt(1_000_000_000)
	if got.Cmp(want) != 0 {
		t.Fatalf("CalcBaseFee (stable) = %s, want %s (baseFee must not change at target)", got, want)
	}
}

// TestBaseFeeIncrease verifies that baseFee increases when gasUsed > gasTarget.
func TestBaseFeeIncrease(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	gasLimit := uint64(30_000_000)
	parent := olympiaHeader(olympiaBlock, gasLimit, gasLimit, big.NewInt(1_000_000_000)) // fully used
	got := CalcBaseFee(cfg, parent)
	if got.Cmp(big.NewInt(1_000_000_000)) <= 0 {
		t.Fatalf("CalcBaseFee (overused) = %s, want > 1 Gwei", got)
	}
}

// TestBaseFeeDecrease verifies that baseFee decreases when gasUsed < gasTarget.
func TestBaseFeeDecrease(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	gasLimit := uint64(30_000_000)
	parent := olympiaHeader(olympiaBlock, gasLimit, 0, big.NewInt(2_000_000_000)) // 2 Gwei, empty
	got := CalcBaseFee(cfg, parent)
	if got.Cmp(big.NewInt(2_000_000_000)) >= 0 {
		t.Fatalf("CalcBaseFee (underused) = %s, want < 2 Gwei", got)
	}
}

// TestBaseFeeFloorsAtInitialBaseFee verifies that on ETC chains the baseFee floor
// clamps any decrease to InitialBaseFee (1 gwei) per ECIP-1111.
// Note: Bug B (delta floor at 1) still applies — the decrease delta is computed
// correctly; the chain-config floor then ensures the result >= InitialBaseFee.
func TestBaseFeeFloorsAtInitialBaseFee(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	gasLimit := uint64(30_000_000)
	tinyBaseFee := big.NewInt(1) // 1 wei — well below InitialBaseFee
	parent := olympiaHeader(olympiaBlock, gasLimit, 0, tinyBaseFee)
	got := CalcBaseFee(cfg, parent)
	floor := new(big.Int).SetUint64(vars.InitialBaseFee)
	if got.Cmp(floor) < 0 {
		t.Fatalf("CalcBaseFee (ETC, baseFee=1 wei) = %s, want >= %s (ECIP-1111 floor not applied)", got, floor)
	}
	if got.Cmp(floor) != 0 {
		t.Fatalf("CalcBaseFee (ETC, baseFee=1 wei) = %s, want exactly %s (clamped to floor)", got, floor)
	}
}

// TestBaseFeeNoFloorOnETH verifies that without a configured floor (ETH mainnet),
// baseFee can decay to 0 — the Bug B delta-floor behaviour is preserved on ETH.
func TestBaseFeeNoFloorOnETH(t *testing.T) {
	const eip1559Block = 100
	cfg := newETHTestConfig(eip1559Block)
	// ETH uses DefaultElasticityMultiplier=2; no OlympiaGasTarget, use a simple gas limit.
	gasLimit := uint64(18_000_000) // 2× 9M target
	tinyBaseFee := big.NewInt(1)   // 1 wei
	parent := olympiaHeader(eip1559Block, gasLimit, 0, tinyBaseFee)
	got := CalcBaseFee(cfg, parent)
	if got.Sign() != 0 {
		t.Fatalf("CalcBaseFee (ETH, no floor, baseFee=1 wei) = %s, want 0", got)
	}
}

// TestBaseFeeNeverBelowFloor verifies that 100 consecutive empty blocks on ETC
// never produce a baseFee below InitialBaseFee (arithmetic invariant with floor).
func TestBaseFeeNeverBelowFloor(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	gasLimit := uint64(30_000_000)
	baseFee := new(big.Int).SetUint64(vars.InitialBaseFee)
	floor := new(big.Int).SetUint64(vars.InitialBaseFee)
	for i := 0; i < 100; i++ {
		parent := olympiaHeader(uint64(olympiaBlock+i), gasLimit, 0, baseFee)
		baseFee = CalcBaseFee(cfg, parent)
		if baseFee.Cmp(floor) < 0 {
			t.Fatalf("block %d: baseFee %s fell below InitialBaseFee floor %s", olympiaBlock+i+1, baseFee, floor)
		}
	}
}

// TestBaseFeeSustainedEmpty1000 verifies that 1000 consecutive empty blocks on ETC
// all produce exactly InitialBaseFee (1 gwei) — never below.
func TestBaseFeeSustainedEmpty1000(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	gasLimit := uint64(30_000_000)
	baseFee := new(big.Int).SetUint64(vars.InitialBaseFee)
	floor := new(big.Int).SetUint64(vars.InitialBaseFee)
	for i := 0; i < 1000; i++ {
		parent := olympiaHeader(uint64(olympiaBlock+i), gasLimit, 0, baseFee)
		baseFee = CalcBaseFee(cfg, parent)
		if baseFee.Cmp(floor) < 0 {
			t.Fatalf("block %d: baseFee %s fell below InitialBaseFee floor %s after 1000 empty blocks",
				olympiaBlock+i+1, baseFee, floor)
		}
	}
}
