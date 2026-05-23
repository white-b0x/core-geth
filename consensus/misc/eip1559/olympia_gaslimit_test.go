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

	"github.com/ethereum/go-ethereum/params/vars"
)

// calcNextGasLimit replicates core.CalcGasLimit: one ±1/1024 step toward ceil.
// Duplicated here to avoid an import cycle (core imports consensus).
func calcNextGasLimit(parent, ceil uint64) uint64 {
	const minGasLimit = 5000
	if ceil < minGasLimit {
		ceil = minGasLimit
	}
	delta := parent/vars.GasLimitBoundDivisor - 1
	if parent < ceil {
		next := parent + delta
		if next > ceil {
			return ceil
		}
		return next
	}
	if parent > ceil {
		next := parent - delta
		if next < ceil {
			return ceil
		}
		return next
	}
	return parent
}

// TestForkGasTargetReturns8M verifies ForkGasTarget returns SpiralGasTarget (8 M)
// for every Spiral-era block (spiralBlock ≤ bn < olympiaBlock).
func TestForkGasTargetReturns8M(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock) // spiralBlock = 50

	for _, bn := range []uint64{50, 75, 99} {
		target := ForkGasTarget(cfg, new(big.Int).SetUint64(bn))
		if target == nil {
			t.Fatalf("block %d: ForkGasTarget returned nil, want 8_000_000", bn)
		}
		if *target != 8_000_000 {
			t.Fatalf("block %d: ForkGasTarget = %d, want 8_000_000", bn, *target)
		}
	}
}

// TestForkGasTargetReturns60M verifies ForkGasTarget returns OlympiaGasTarget (60 M)
// for every Olympia-era block (bn ≥ olympiaBlock).
func TestForkGasTargetReturns60M(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)

	for _, bn := range []uint64{100, 101, 10_000} {
		target := ForkGasTarget(cfg, new(big.Int).SetUint64(bn))
		if target == nil {
			t.Fatalf("block %d: ForkGasTarget returned nil, want 60_000_000", bn)
		}
		if *target != 60_000_000 {
			t.Fatalf("block %d: ForkGasTarget = %d, want 60_000_000", bn, *target)
		}
	}
}

// TestGasLimitStepAtOlympiaActivation verifies VerifyEIP1559Header accepts a header with
// exactly one ±1/1024 step increase from 8 M at the first Olympia block (8,007,811).
// delta = 8_000_000/1024 - 1 = 7811 → valid range is (7M+7811, 8M+7811).
func TestGasLimitStepAtOlympiaActivation(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	parent := preOlympiaHeader(olympiaBlock-1, 8_000_000, 0)
	child := olympiaHeader(olympiaBlock, 8_007_811, 0, new(big.Int).SetUint64(vars.InitialBaseFee))
	if err := VerifyEIP1559Header(cfg, parent, child); err != nil {
		t.Fatalf("VerifyEIP1559Header rejected valid Olympia first block: %v", err)
	}
}

// TestGasLimitStepTwoAtOlympia verifies VerifyEIP1559Header accepts the second Olympia
// block with gasLimit derived from 8,007,811 (two steps from 8 M).
func TestGasLimitStepTwoAtOlympia(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	// Block 100 as parent: 8,007,811 gas, empty block.
	// delta = 8,007,811/1024 - 1 = 7820 - 1 = 7819 → next = 8,015,630
	parentBaseFee := new(big.Int).SetUint64(vars.InitialBaseFee)
	parent := olympiaHeader(olympiaBlock, 8_007_811, 0, parentBaseFee)
	childBaseFee := CalcBaseFee(cfg, parent)
	child := olympiaHeader(olympiaBlock+1, 8_015_630, 0, childBaseFee)
	if err := VerifyEIP1559Header(cfg, parent, child); err != nil {
		t.Fatalf("VerifyEIP1559Header rejected valid second Olympia block: %v", err)
	}
}

// TestNo2xGasJumpAtOlympiaActivation verifies VerifyEIP1559Header rejects a 16 M gasLimit
// at the first Olympia block. This is the direct regression test for Bug A: ETH London
// doubles the parentGasLimit at EIP-1559 activation, but ETC Olympia must NOT do this.
func TestNo2xGasJumpAtOlympiaActivation(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	parent := preOlympiaHeader(olympiaBlock-1, 8_000_000, 0)
	// 16M is the ETH-style 2× jump; should be rejected on ETC.
	child := olympiaHeader(olympiaBlock, 16_000_000, 0, new(big.Int).SetUint64(vars.InitialBaseFee))
	if err := VerifyEIP1559Header(cfg, parent, child); err == nil {
		t.Fatal("VerifyEIP1559Header accepted 16M gasLimit at first Olympia block — Bug A not fixed")
	}
}

// TestGasLimitConvergesTo60MIn2055Blocks verifies that starting from 8 M, the gas limit
// converges to ≥59,400,000 (99% of 60 M) within 2,055 blocks.
// Cross-client constant: all compliant clients must match this trajectory exactly.
func TestGasLimitConvergesTo60MIn2055Blocks(t *testing.T) {
	const (
		olympiaGasTarget  = uint64(60_000_000)
		convergenceBlocks = 2055
		minConverged      = uint64(59_400_000) // 99% of 60 M
	)
	gasLimit := uint64(8_000_000)
	for i := 0; i < convergenceBlocks; i++ {
		gasLimit = calcNextGasLimit(gasLimit, olympiaGasTarget)
	}
	if gasLimit < minConverged {
		t.Fatalf("after %d blocks, gasLimit = %d, want ≥%d (99%% of 60 M)",
			convergenceBlocks, gasLimit, minConverged)
	}
}
