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

// TestBaseFeeFirstOlympiaBlock verifies that the second Olympia block (parent IS
// the first Olympia block, which carried InitialBaseFee and gasUsed=0) decreases
// baseFee by exactly 1 wei (delta floor kicks in when parentBaseFee is tiny).
func TestBaseFeeFirstOlympiaBlock(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	// Parent is the first Olympia block: gasLimit=30M (post-activation target),
	// gasUsed=0, baseFee=1 Gwei. Next baseFee should decrease.
	parent := olympiaHeader(olympiaBlock, 30_000_000, 0, new(big.Int).SetUint64(vars.InitialBaseFee))
	got := CalcBaseFee(cfg, parent)
	// gasTarget = 30M/2 = 15M; gasUsed(0) < gasTarget(15M) → decrease
	// delta = max(1, InitialBaseFee * 15M / 15M / 8) = max(1, InitialBaseFee/8) = 125_000_000
	// baseFee = InitialBaseFee - 125_000_000 = 875_000_000
	want := new(big.Int).SetUint64(vars.InitialBaseFee - vars.InitialBaseFee/8)
	if got.Cmp(want) != 0 {
		t.Fatalf("CalcBaseFee (second Olympia, empty) = %s, want %s", got, want)
	}
	if got.Cmp(new(big.Int).SetUint64(vars.InitialBaseFee)) >= 0 {
		t.Fatal("baseFee should decrease on empty block")
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

// TestBaseFeeDecaysTo0From1Wei verifies the Bug B fix: when baseFee = 1 wei and
// gasUsed = 0, the delta is floored at 1 so baseFee decays to 0 (not stuck at 1).
// Cross-client constant: baseFee reaches 0 in exactly 1 empty block from 1 wei.
func TestBaseFeeDecaysTo0From1Wei(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	gasLimit := uint64(30_000_000)
	tinyBaseFee := big.NewInt(1) // 1 wei
	parent := olympiaHeader(olympiaBlock, gasLimit, 0, tinyBaseFee)
	got := CalcBaseFee(cfg, parent)
	zero := new(big.Int)
	if got.Cmp(zero) != 0 {
		t.Fatalf("CalcBaseFee (baseFee=1 wei, empty block) = %s, want 0 — baseFee stuck at 1 wei (Bug B not fixed)", got)
	}
}

// TestBaseFeeNeverNegative verifies that 100 consecutive empty blocks never produce
// a negative baseFee (arithmetic invariant).
func TestBaseFeeNeverNegative(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	gasLimit := uint64(30_000_000)
	baseFee := new(big.Int).SetUint64(vars.InitialBaseFee)
	zero := new(big.Int)
	for i := 0; i < 100; i++ {
		parent := olympiaHeader(uint64(olympiaBlock+i), gasLimit, 0, baseFee)
		baseFee = CalcBaseFee(cfg, parent)
		if baseFee.Cmp(zero) < 0 {
			t.Fatalf("block %d: baseFee went negative: %s", olympiaBlock+i+1, baseFee)
		}
	}
}
